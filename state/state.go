package state

import (
	"bytes"
	"encoding/json"
	"math/big"
	"sync"
	"syscall"

	"github.com/babbleio/babble/hashgraph"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core"
	ethState "github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/sirupsen/logrus"

	bcommon "github.com/nic0lae/evm-babble/common"
)

var (
	chainID        = big.NewInt(1)
	gasLimit       = big.NewInt(1000000000000000000)
	txMetaSuffix   = []byte{0x01}
	receiptsPrefix = []byte("receipts-")
	MIPMapLevels   = []uint64{1000000, 500000, 100000, 50000, 1000}
	headTxKey      = []byte("LastTx")
)

type State struct {
	db          ethdb.Database
	commitMutex sync.Mutex
	statedb     *ethState.StateDB
	was         *WriteAheadState

	signer      ethTypes.Signer
	chainConfig params.ChainConfig //vm.env is still tightly coupled with chainConfig
	vmConfig    vm.Config

	logger *logrus.Logger
}

func NewState(logger *logrus.Logger, dbFile string, dbCache int) (*State, error) {

	handles, err := getFdLimit()
	if err != nil {
		return nil, err
	}

	db, err := ethdb.NewLDBDatabase(dbFile, dbCache, handles)
	if err != nil {
		return nil, err
	}

	s := new(State)
	s.logger = logger
	s.db = db
	s.signer = ethTypes.NewEIP155Signer(chainID)
	s.chainConfig = params.ChainConfig{ChainId: chainID}
	s.vmConfig = vm.Config{Tracer: vm.NewStructLogger(nil)}

	if err := s.InitState(); err != nil {
		return nil, err
	}

	s.resetWAS()

	return s, nil
}

// getFdLimit retrieves the number of file descriptors allowed to be opened by this
// process.
func getFdLimit() (int, error) {
	var limit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return 0, err
	}
	return int(limit.Cur), nil
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
	vmenv := vm.NewEVM(context, s.was.ethState.Copy(), &s.chainConfig, s.vmConfig)

	// Apply the transaction to the current state (included in the env)
	res, _, _, err := core.ApplyMessage(vmenv, callMsg, s.was.gp)
	if err != nil {
		s.logger.WithError(err).Error("Executing Call on WAS")
		return nil, err
	}

	return res, err
}

func (s *State) ProcessBlock(block hashgraph.Block) (common.Hash, error) {
	s.logger.Debug("Process Block")
	s.commitMutex.Lock()
	defer s.commitMutex.Unlock()

	blockHashBytes, _ := block.Hash()
	blockHash := common.BytesToHash(blockHashBytes)

	for txIndex, txBytes := range block.Transactions() {
		if err := s.applyTransaction(txBytes, txIndex, blockHash); err != nil {
			return common.Hash{}, err
		}
	}

	return s.commit()
}

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
func txAsString(tx ethTypes.Transaction) string {
	stuff, _ := json.Marshal(tx)
	return string(stuff)
}

//applyTransaction applies a transaction to the WAS
func (s *State) applyTransaction(txBytes []byte, txIndex int, blockHash common.Hash) error {

	var t ethTypes.Transaction
	if err := rlp.Decode(bytes.NewReader(txBytes), &t); err != nil {
		s.logger.WithError(err).Error("Decoding Transaction")
		return err
	}
	s.logger.WithField("hash", t.Hash().Hex()).Debug("Decoded tx")
	s.logger.WithField("tx", txAsString(t)).Debug()

	msg, err := t.AsMessage(s.signer)
	if err != nil {
		s.logger.WithError(err).Error("Converting Transaction to Message")
		return err
	}

	context := vm.Context{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     func(uint64) common.Hash { return blockHash },
		// Message information
		Origin:      msg.From(),
		GasLimit:    msg.Gas(),
		GasPrice:    msg.GasPrice(),
		BlockNumber: big.NewInt(0), //the vm has a dependency on this..
	}

	//Prepare the ethState with transaction Hash so that it can be used in emitted
	//logs
	s.was.ethState.Prepare(t.Hash(), blockHash, txIndex)

	// The EVM should never be reused and is not thread safe.
	vmenv := vm.NewEVM(context, s.was.ethState, &s.chainConfig, s.vmConfig)

	// Apply the transaction to the current state (included in the env)
	_, gas, failed, err := core.ApplyMessage(vmenv, msg, s.was.gp)
	if err != nil {
		s.logger.WithError(err).Error("Applying transaction to WAS")
		return err
	}

	s.was.totalUsedGas.Add(s.was.totalUsedGas, big.NewInt(0).SetUint64(gas))

	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing wether the root touch-delete accounts.
	root := s.was.ethState.IntermediateRoot(true) //this has side effects. It updates StateObjects (SmartContract memory)
	receipt := ethTypes.NewReceipt(root.Bytes(), failed, s.was.totalUsedGas.Uint64())
	receipt.TxHash = t.Hash()
	receipt.GasUsed = gas
	// if the transaction created a contract, store the creation address in the receipt.
	if msg.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(vmenv.Context.Origin, t.Nonce())
	}
	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = s.was.ethState.GetLogs(t.Hash())
	//receipt.Logs = s.was.state.Logs()
	receipt.Bloom = ethTypes.CreateBloom(ethTypes.Receipts{receipt})

	s.was.txIndex++
	s.was.transactions = append(s.was.transactions, &t)
	s.was.receipts = append(s.was.receipts, receipt)
	s.was.allLogs = append(s.was.allLogs, receipt.Logs...)

	s.logger.WithField("hash", t.Hash().Hex()).Debug("Applied tx to WAS")

	return nil
}

func (s *State) commit() (common.Hash, error) {
	//commit all state changes to the database
	root, err := s.was.Commit()
	if err != nil {
		s.logger.WithError(err).Error("Committing WAS")
		return root, err
	}

	// reset the write ahead state for the next block
	// with the latest eth state
	s.statedb = s.was.ethState
	s.logger.WithField("root", root.Hex()).Debug("Committed")
	s.resetWAS()

	return root, nil
}

func (s *State) resetWAS() {
	state := s.statedb.Copy()
	s.was = &WriteAheadState{
		db:           s.db,
		ethState:     state,
		txIndex:      0,
		totalUsedGas: big.NewInt(0),
		gp:           new(core.GasPool).AddGas(gasLimit.Uint64()),
		logger:       s.logger,
	}
	s.logger.Debug("Reset Write Ahead State")
}

//------------------------------------------------------------------------------

func (s *State) InitState() error {

	rootHash := common.Hash{}

	//get head transaction hash
	headTxHash := common.Hash{}
	data, _ := s.db.Get(headTxKey)
	if len(data) != 0 {
		headTxHash = common.BytesToHash(data)
		s.logger.WithField("head_tx", headTxHash.Hex()).Debug("Loading state from existing head")
		//get head tx receipt
		headTxReceipt, err := s.GetReceipt(headTxHash)
		if err != nil {
			s.logger.WithError(err).Error("Head transaction receipt missing")
			return err
		}

		//extract root from receipt
		if len(headTxReceipt.PostState) != 0 {
			rootHash = common.BytesToHash(headTxReceipt.PostState)
			s.logger.WithField("root", rootHash.Hex()).Debug("Head transaction root")
		}
	}

	//use root to initialise the state
	var err error
	s.statedb, err = ethState.New(rootHash, ethState.NewDatabase(s.db))
	return err
}

func (s *State) CreateAccounts(accounts bcommon.AccountMap) error {
	s.commitMutex.Lock()
	defer s.commitMutex.Unlock()

	for addr, account := range accounts {
		address := common.HexToAddress(addr)
		s.was.ethState.AddBalance(address, math.MustParseBig256(account.Balance))
		s.was.ethState.SetCode(address, common.Hex2Bytes(account.Code))
		for key, value := range account.Storage {
			s.was.ethState.SetState(address, common.HexToHash(key), common.HexToHash(value))
		}
		s.logger.WithField("address", addr).Debug("Adding account")
	}

	_, err := s.commit()

	return err
}

func (s *State) GetBalance(addr common.Address) *big.Int {
	return s.statedb.GetBalance(addr)
}

func (s *State) GetNonce(addr common.Address) uint64 {
	return s.was.ethState.GetNonce(addr)
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
