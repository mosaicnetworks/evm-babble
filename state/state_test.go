package state

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	bcommon "github.com/babbleio/evm-babble/common"
)

var (
	_defaultValue    = big.NewInt(0)
	_defaultGas      = big.NewInt(1000000)
	_defaultGasPrice = big.NewInt(0)
)

type Test struct {
	dataDir string
	pwdFile string
	dbFile  string
	cache   int

	keyStore *keystore.KeyStore
	state    *State
	logger   *logrus.Logger
}

func NewTest(dataDir string, logger *logrus.Logger, t *testing.T) *Test {
	pwdFile := filepath.Join(dataDir, "pwd.txt")
	dbFile := filepath.Join(dataDir, "chaindata")
	cache := 128

	state, err := NewState(logger, dbFile, cache)
	if err != nil {
		t.Fatal(err)
	}

	return &Test{
		dataDir: dataDir,
		pwdFile: pwdFile,
		dbFile:  dbFile,
		cache:   cache,
		state:   state,
		logger:  logger,
	}
}

func (test *Test) readPwd() (pwd string, err error) {
	text, err := ioutil.ReadFile(test.pwdFile)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(text), "\n")
	// Sanitise DOS line endings.
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}
	return lines[0], nil
}

func (test *Test) initKeyStore() error {
	scryptN := keystore.StandardScryptN
	scryptP := keystore.StandardScryptP

	keydir := filepath.Join(test.dataDir, "keystore")
	if err := os.MkdirAll(keydir, 0700); err != nil {
		return err
	}

	test.keyStore = keystore.NewKeyStore(keydir, scryptN, scryptP)

	return nil
}

func (test *Test) unlockAccounts() error {
	pwd, err := test.readPwd()
	if err != nil {
		test.logger.WithError(err).Error("Reading PwdFile")
		return err
	}

	for _, ac := range test.keyStore.Accounts() {
		if err := test.keyStore.Unlock(ac, string(pwd)); err != nil {
			return err
		}
		test.logger.WithField("address", ac.Address.Hex()).Debug("Unlocked account")
	}
	return nil
}

func (test *Test) createGenesisAccounts() error {
	genesisFile := filepath.Join(test.dataDir, "genesis.json")

	contents, err := ioutil.ReadFile(genesisFile)
	if err != nil {
		return err
	}

	var genesis struct {
		Alloc bcommon.AccountMap
	}

	if err := json.Unmarshal(contents, &genesis); err != nil {
		return err
	}

	if err := test.state.CreateAccounts(genesis.Alloc); err != nil {
		return err
	}
	return nil
}

func (test *Test) Init() error {
	if err := test.initKeyStore(); err != nil {
		return err
	}

	if err := test.unlockAccounts(); err != nil {
		return err
	}

	if err := test.createGenesisAccounts(); err != nil {
		return err
	}

	return nil
}

func (test *Test) prepareTransaction(from, to *accounts.Account,
	value, gas, gasPrice *big.Int,
	data []byte) (*ethTypes.Transaction, error) {

	nonce := test.state.GetNonce(from.Address)

	var tx *ethTypes.Transaction
	if to == nil {
		tx = ethTypes.NewContractCreation(nonce,
			value,
			gas,
			gasPrice,
			data)
	} else {
		tx = ethTypes.NewTransaction(nonce,
			to.Address,
			value,
			gas,
			gasPrice,
			data)
	}

	signer := ethTypes.NewEIP155Signer(big.NewInt(1))

	signature, err := test.keyStore.SignHash(*from, signer.Hash(tx).Bytes())
	if err != nil {
		return nil, err
	}
	signedTx, err := tx.WithSignature(signer, signature)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

func (test *Test) deployContract(from accounts.Account, contract *Contract, t *testing.T) {

	//Create Contract transaction
	tx, err := test.prepareTransaction(&from,
		nil,
		_defaultValue,
		_defaultGas,
		_defaultGasPrice,
		common.FromHex(contract.code))

	if err != nil {
		t.Fatal(err)
	}

	//convert to raw bytes
	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		t.Fatal(err)
	}

	//try to append tx
	err = test.state.AppendTx(data)
	if err != nil {
		t.Fatal(err)
	}

	err = test.state.Commit()
	if err != nil {
		t.Fatal(err)
	}

	receipt, err := test.state.GetReceipt(tx.Hash())
	if err != nil {
		t.Fatal(err)
	}

	contract.address = receipt.ContractAddress
}

//------------------------------------------------------------------------------
func TestTransfer(t *testing.T) {

	os.RemoveAll("test_data/eth/chaindata")
	defer os.RemoveAll("test_data/eth/chaindata")

	test := NewTest("test_data/eth", bcommon.NewTestLogger(t), t)
	defer test.state.db.Close()

	err := test.Init()

	if err != nil {
		t.Fatal(err)
	}

	from := test.keyStore.Accounts()[0]
	fromBalanceBefore := test.state.GetBalance(from.Address)
	to := test.keyStore.Accounts()[1]
	toBalanceBefore := test.state.GetBalance(to.Address)

	//Create transfer transaction
	value := big.NewInt(1000000)
	gas := big.NewInt(21000) //a value transfer transaction costs 21000 gas
	gasPrice := big.NewInt(0)

	tx, err := test.prepareTransaction(&from,
		&to,
		value,
		gas,
		gasPrice,
		[]byte{})

	if err != nil {
		t.Fatal(err)
	}

	//convert to raw bytes
	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		t.Fatal(err)
	}

	//try to append tx
	err = test.state.AppendTx(data)
	if err != nil {
		t.Fatal(err)
	}

	err = test.state.Commit()
	if err != nil {
		t.Fatal(err)
	}

	fromBalanceAfter := test.state.GetBalance(from.Address)
	expectedFromBalanceAfter := big.NewInt(0)
	expectedFromBalanceAfter.Sub(fromBalanceBefore, value)
	toBalanceAfter := test.state.GetBalance(to.Address)
	expectedToBalanceAfter := big.NewInt(0)
	expectedToBalanceAfter.Add(toBalanceBefore, value)

	if fromBalanceAfter.Cmp(expectedFromBalanceAfter) != 0 {
		t.Fatalf("fromBalanceAfter should be %v, not %v", expectedFromBalanceAfter, fromBalanceAfter)
	}

	if toBalanceAfter.Cmp(expectedToBalanceAfter) != 0 {
		t.Fatalf("toBalanceAfter should be %v, not %v", expectedToBalanceAfter, toBalanceAfter)
	}
}

//------------------------------------------------------------------------------
type Contract struct {
	name    string
	address common.Address
	code    string
	abi     string
	jsonABI abi.ABI
}

func (c *Contract) parseABI(t *testing.T) {
	jABI, err := abi.JSON(strings.NewReader(c.abi))
	if err != nil {
		t.Fatal(err)
	}
	c.jsonABI = jABI
}

/*

pragma solidity 0.4.8;

contract Test {

   uint localI = 1;

   event LocalChange(uint);

   function test(uint i) constant returns (uint){
        return localI * i;
   }

   function testAsync(uint i) {
        localI += i;
        LocalChange(localI);
   }
}

*/

func dummyContract() *Contract {
	return &Contract{
		name: "Test",
		code: "6060604052600160005534610000575b6101168061001e6000396000f30060606040526000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806329e99f07146046578063cb0d1c76146074575b6000565b34600057605e6004808035906020019091905050608e565b6040518082815260200191505060405180910390f35b34600057608c6004808035906020019091905050609d565b005b6000816000540290505b919050565b806000600082825401925050819055507ffa753cb3413ce224c9858a63f9d3cf8d9d02295bdb4916a594b41499014bb57f6000546040518082815260200191505060405180910390a15b505600a165627a7a723058203f0887849cabeb36c6f72cc345c5ff3521d889356357e6815dd8dbe9f7c41bbe0029",
		abi:  "[{\"constant\":true,\"inputs\":[{\"name\":\"i\",\"type\":\"uint256\"}],\"name\":\"test\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\",\"stateMutability\":\"view\"},{\"constant\":false,\"inputs\":[{\"name\":\"i\",\"type\":\"uint256\"}],\"name\":\"testAsync\",\"outputs\":[],\"payable\":false,\"type\":\"function\",\"stateMutability\":\"nonpayable\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"LocalChange\",\"type\":\"event\"}]",
	}
}

func callDummyContractTest(test *Test, from accounts.Account, contract *Contract, expected *big.Int, t *testing.T) {
	callData, err := contract.jsonABI.Pack("test", big.NewInt(10))
	if err != nil {
		t.Fatal(err)
	}

	callMsg := ethTypes.NewMessage(from.Address,
		&contract.address,
		0,
		_defaultValue,
		_defaultGas,
		_defaultGasPrice,
		callData,
		false)

	if err != nil {
		t.Fatal(err)
	}

	res, err := test.state.Call(callMsg)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("call res: %v", res)

	var parsedRes *big.Int
	err = contract.jsonABI.Unpack(&parsedRes, "test", res)
	if err != nil {
		t.Error(err)
	}
	t.Logf("parsed res: %v", parsedRes)

	if parsedRes.Cmp(expected) != 0 {
		t.Fatalf("Result should be %v, not %v", expected, parsedRes)
	}

}

func callDummyContractTestAsync(test *Test, from accounts.Account, contract *Contract, t *testing.T) {
	callData, err := contract.jsonABI.Pack("testAsync", big.NewInt(10))
	if err != nil {
		t.Fatal(err)
	}

	tx, err := test.prepareTransaction(&from,
		&accounts.Account{Address: contract.address},
		_defaultValue,
		_defaultGas,
		_defaultGasPrice,
		callData)

	if err != nil {
		t.Fatal(err)
	}

	//convert to raw bytes
	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		t.Fatal(err)
	}

	//try to append tx
	err = test.state.AppendTx(data)
	if err != nil {
		t.Fatal(err)
	}

	err = test.state.Commit()
	if err != nil {
		t.Fatal(err)
	}

	receipt, err := test.state.GetReceipt(tx.Hash())
	if err != nil {
		t.Fatal(err)
	}

	t.Log(receipt)
}

func TestCreateContract(t *testing.T) {

	os.RemoveAll("test_data/eth/chaindata")
	defer os.RemoveAll("test_data/eth/chaindata")

	test := NewTest("test_data/eth", bcommon.NewTestLogger(t), t)
	defer test.state.db.Close()

	err := test.Init()

	if err != nil {
		t.Fatal(err)
	}

	from := test.keyStore.Accounts()[0]

	contract := dummyContract()

	test.deployContract(from, contract, t)

	contract.parseABI(t)

	//call constant test method
	callDummyContractTest(test, from, contract, big.NewInt(10), t)

	//execute state-altering testAsync method
	callDummyContractTestAsync(test, from, contract, t)

	//call constant test method
	callDummyContractTest(test, from, contract, big.NewInt(110), t)

}

func TestDB(t *testing.T) {

	os.RemoveAll("test_data/eth/chaindata")
	defer os.RemoveAll("test_data/eth/chaindata")

	//initialise a fresh instance and commit stuff to the db
	test := NewTest("test_data/eth", bcommon.NewTestLogger(t), t)
	if err := test.Init(); err != nil {
		t.Fatal(err)
	}

	from := test.keyStore.Accounts()[0]

	contract := dummyContract()

	test.deployContract(from, contract, t)

	code := test.state.statedb.GetCode(contract.address)
	t.Logf("code: %s", common.ToHex(code))

	contract.parseABI(t)

	//execute state-altering testAsync method
	callDummyContractTestAsync(test, from, contract, t)

	//close the database
	test.state.db.Close()

	//initialise another instance from the existing db
	test2 := NewTest("test_data/eth", bcommon.NewTestLogger(t), t)
	if err := test2.Init(); err != nil {
		t.Fatal(err)
	}

	//check that state is the same

	//check that contract code is there
	code2 := test2.state.statedb.GetCode(contract.address)
	t.Logf("code2: %s", common.ToHex(code2))
	if !reflect.DeepEqual(code2, code) {
		t.Fatalf("contract code should be equal")
	}

	//check contract memory
	callDummyContractTest(test2, from, contract, big.NewInt(110), t)

}
