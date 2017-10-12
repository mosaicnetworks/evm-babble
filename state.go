package evmbabble

import (
	"bytes"
	"math/big"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
)

var (
	chainID        = big.NewInt(1)
	gasLimit       = big.NewInt(1000000000000000000)
	txMetaSuffix   = []byte{0x01}
	receiptsPrefix = []byte("receipts-")
	MIPMapLevels   = []uint64{1000000, 500000, 100000, 50000, 1000}
)

type State struct {
	proxy *Proxy

	db          ethdb.Database
	commitMutex sync.Mutex
	statedb     *state.StateDB
	was         *WriteAheadState

	signer      ethTypes.Signer
	chainConfig params.ChainConfig //vm.env is still tightly coupled with chainConfig
	vmConfig    vm.Config

	logger *logrus.Logger
}

func NewState(logger *logrus.Logger) (*State, error) {

	db, err := ethdb.NewMemDatabase() //ephemeral database
	if err != nil {
		return nil, err
	}

	state, err := state.New(common.Hash{}, state.NewDatabase(db))
	if err != nil {
		return nil, err
	}

	s := new(State)
	s.logger = logger
	s.db = db
	s.statedb = state
	s.signer = ethTypes.NewEIP155Signer(chainID)
	s.chainConfig = params.ChainConfig{ChainId: chainID}
	s.vmConfig = vm.Config{Tracer: vm.NewStructLogger(nil)}

	s.resetWAS(state.Copy())

	return s, nil
}

func (s *State) Init(proxy *Proxy) error {
	s.proxy = proxy
	return nil
}

//------------------------------------------------------------------------------

func (s *State) Call(callMsg ethTypes.Message) ([]byte, error) {
	s.logger.Debug("Call")
	s.commitMutex.Lock()
	defer s.commitMutex.Unlock()

	context := vm.Context{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     func(uint64) common.Hash { return common.Hash{} },
		// Message information
		Origin:   callMsg.From(),
		GasPrice: callMsg.GasPrice(),
	}

	// The EVM should never be reused and is not thread safe.
	// Call is done on a copy of the state...we dont want any changes to be persisted
	// Call is a readonly operation
	vmenv := vm.NewEVM(context, s.was.state.Copy(), &s.chainConfig, s.vmConfig)

	// Apply the transaction to the current state (included in the env)
	res, _, _, err := core.ApplyMessage(vmenv, callMsg, s.was.gp)
	if err != nil {
		s.logger.WithError(err).Error("Executing Call on WAS")
		return nil, err
	}

	return res, err
}

// AppendTx applies the tx to the WAS
func (s *State) AppendTx(tx []byte) error {
	s.logger.Debug("AppendTx")
	s.commitMutex.Lock()
	defer s.commitMutex.Unlock()

	var t ethTypes.Transaction
	if err := rlp.Decode(bytes.NewReader(tx), &t); err != nil {
		s.logger.WithError(err).Error("Decoding Transaction")
		return err
	}
	s.logger.WithField("hash", t.Hash().Hex()).Debug("Decoded tx")

	msg, err := t.AsMessage(s.signer)
	if err != nil {
		s.logger.WithError(err).Error("Converting Transaction to Message")
		return err
	}

	context := vm.Context{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     func(uint64) common.Hash { return common.Hash{} },
		// Message information
		Origin:      msg.From(),
		GasPrice:    msg.GasPrice(),
		BlockNumber: big.NewInt(0), //the vm has a dependency on this..
	}

	//XXX
	s.was.state.Prepare(t.Hash(), common.Hash{}, 0)

	// The EVM should never be reused and is not thread safe.
	vmenv := vm.NewEVM(context, s.was.state, &s.chainConfig, s.vmConfig)

	// Apply the transaction to the current state (included in the env)
	_, gas, failed, err := core.ApplyMessage(vmenv, msg, s.was.gp)
	if err != nil {
		s.logger.WithError(err).Error("Applying transaction to WAS")
		return err
	}

	s.was.totalUsedGas.Add(s.was.totalUsedGas, gas)

	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing wether the root touch-delete accounts.
	root := s.was.state.IntermediateRoot(true) //this has side effects. It updates StateObjects (SmartContract memory)
	receipt := ethTypes.NewReceipt(root.Bytes(), failed, s.was.totalUsedGas)
	receipt.TxHash = t.Hash()
	receipt.GasUsed = new(big.Int).Set(gas)
	// if the transaction created a contract, store the creation address in the receipt.
	if msg.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(vmenv.Context.Origin, t.Nonce())
	}
	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = s.was.state.GetLogs(t.Hash())
	//receipt.Logs = s.was.state.Logs()
	receipt.Bloom = ethTypes.CreateBloom(ethTypes.Receipts{receipt})

	s.was.txIndex++
	s.was.transactions = append(s.was.transactions, &t)
	s.was.receipts = append(s.was.receipts, receipt)
	s.was.allLogs = append(s.was.allLogs, receipt.Logs...)

	s.logger.WithField("hash", t.Hash().Hex()).Debug("Applied tx to WAS")

	return nil
}

// Commit then reset the WAS
func (s *State) Commit() error {
	s.logger.Debug("Commit")
	s.commitMutex.Lock()
	defer s.commitMutex.Unlock()

	return s.commit()
}

//------------------------------------------------------------------------------

func (s *State) commit() error {
	//commit all state changes to the database
	root, err := s.was.Commit()
	if err != nil {
		s.logger.WithError(err).Error("Committing WAS")
		return err
	}

	// reset the write ahead state for the next block
	// with the latest eth state
	s.statedb = s.was.state
	s.logger.WithField("root", root.Hex()).Debug("Committed")
	s.resetWAS(s.statedb.Copy())

	return nil
}

func (s *State) resetWAS(state *state.StateDB) {
	s.was = &WriteAheadState{
		db:           s.db,
		state:        state,
		txIndex:      0,
		totalUsedGas: big.NewInt(0),
		gp:           new(core.GasPool).AddGas(gasLimit),
		logger:       s.logger,
	}
	s.logger.Debug("Reset Write Ahead State")
}

//------------------------------------------------------------------------------

func (s *State) CreateAccounts(accounts AccountMap) error {
	s.commitMutex.Lock()
	defer s.commitMutex.Unlock()

	for addr, account := range accounts {
		address := common.HexToAddress(addr)
		s.was.state.AddBalance(address, math.MustParseBig256(account.Balance))
		s.was.state.SetCode(address, common.Hex2Bytes(account.Code))
		for key, value := range account.Storage {
			s.was.state.SetState(address, common.HexToHash(key), common.HexToHash(value))
		}
		s.logger.WithField("address", addr).Debug("Adding account")
	}

	return s.commit()
}

func (s *State) GetBalance(addr common.Address) *big.Int {
	return s.statedb.GetBalance(addr)
}

func (s *State) GetNonce(addr common.Address) uint64 {
	return s.was.state.GetNonce(addr)
}

func (s *State) GetTransaction(hash common.Hash) (*ethTypes.Transaction, error) {
	// Retrieve the transaction itself from the database
	data, err := s.db.Get(hash.Bytes())
	if err != nil {
		s.logger.WithError(err).Error("GetTransaction")
		return nil, err
	}
	var tx ethTypes.Transaction
	if err := rlp.DecodeBytes(data, &tx); err != nil {
		s.logger.WithError(err).Error("Decoding Transaction")
		return nil, err
	}

	return &tx, nil
}

func (s *State) GetReceipt(txHash common.Hash) (*ethTypes.Receipt, error) {
	data, err := s.db.Get(append(receiptsPrefix, txHash[:]...))
	if err != nil {
		s.logger.WithError(err).Error("GetReceipt")
		return nil, err
	}
	var receipt ethTypes.ReceiptForStorage
	if err := rlp.DecodeBytes(data, &receipt); err != nil {
		s.logger.WithError(err).Error("Decoding Receipt")
		return nil, err
	}

	return (*ethTypes.Receipt)(&receipt), nil
}
