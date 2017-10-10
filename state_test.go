package evmbabble

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

type Test struct {
	dataDir string
	pwdFile string

	keyStore *keystore.KeyStore
	state    *State
	logger   *logrus.Logger
}

func NewTest(dataDir, pwdFile string, logger *logrus.Logger) *Test {
	state, err := NewState(logger)
	if err != nil {
		os.Exit(1)
	}

	return &Test{
		dataDir: dataDir,
		pwdFile: pwdFile,
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
		Alloc AccountMap
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
	data string) (*ethTypes.Transaction, error) {

	nonce := test.state.GetNonce(from.Address)

	var tx *ethTypes.Transaction
	if to == nil {
		tx = ethTypes.NewContractCreation(nonce,
			value,
			gas,
			gasPrice,
			common.FromHex(data))
	} else {
		tx = ethTypes.NewTransaction(nonce,
			to.Address,
			value,
			gas,
			gasPrice,
			common.FromHex(data))
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

//------------------------------------------------------------------------------
func TestTransfer(t *testing.T) {
	test := NewTest("test_data/eth", "test_data/eth/pwd.txt", NewTestLogger(t))

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
		"")

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
}

/*

pragma solidity 0.4.8;

contract Test {

   uint localI = 1;

   event LocalChange(uint);

   function test(uint i) constant returns (uint){
        return i * 10;
   }

   function testAsync(uint i) {
        localI += i;
        LocalChange(localI);
   }
}

*/

//Compiled with solc v0.4.8 commit 60cc1668..
//the version of evm we are using does not support some bytecodes generated by
//later versions of solc
func dummyContract() Contract {
	return Contract{
		name: "Test",
		code: "6060604052600160005534610000575b6101158061001e6000396000f30060606040526000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806329e99f07146046578063cb0d1c76146074575b6000565b34600057605e6004808035906020019091905050608e565b6040518082815260200191505060405180910390f35b34600057608c6004808035906020019091905050609c565b005b6000600a820290505b919050565b806000600082825401925050819055507ffa753cb3413ce224c9858a63f9d3cf8d9d02295bdb4916a594b41499014bb57f6000546040518082815260200191505060405180910390a15b505600a165627a7a72305820c6efb8842641b4ae24d8981702d2f3edd59b71ed10abfde086697615bfb4af360029",
		abi:  "[{\"constant\":true,\"inputs\":[{\"name\":\"i\",\"type\":\"uint256\"}],\"name\":\"test\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\",\"stateMutability\":\"view\"},{\"constant\":false,\"inputs\":[{\"name\":\"i\",\"type\":\"uint256\"}],\"name\":\"testAsync\",\"outputs\":[],\"payable\":false,\"type\":\"function\",\"stateMutability\":\"nonpayable\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"LocalChange\",\"type\":\"event\"}]",
	}
}

func TestCreateContract(t *testing.T) {
	test := NewTest("test_data/eth", "test_data/eth/pwd.txt", NewTestLogger(t))

	err := test.Init()

	if err != nil {
		t.Fatal(err)
	}

	from := test.keyStore.Accounts()[0]

	//Create Contract transaction
	value := big.NewInt(0)
	gas := big.NewInt(1000000)
	gasPrice := big.NewInt(0)

	tx, err := test.prepareTransaction(&from,
		nil,
		value,
		gas,
		gasPrice,
		dummyContract().code)

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

	contractAddress := receipt.ContractAddress

	code := test.state.statedb.GetCode(contractAddress)

	t.Log(common.ToHex(code))

	abi, err := abi.JSON(strings.NewReader(dummyContract().abi))
	if err != nil {
		t.Fatal(err)
	}

	//call constant test method

	callData, err := abi.Pack("test", big.NewInt(10))
	if err != nil {
		t.Fatal(err)
	}

	callMsg := ethTypes.NewMessage(from.Address,
		&contractAddress,
		0,
		value,
		gas,
		gasPrice,
		callData,
		false)

	if err != nil {
		t.Fatal(err)
	}

	res, err := test.state.Call(callMsg)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(res)

	var parsedRes *big.Int
	err = abi.Unpack(&parsedRes, "test", res)
	if err != nil {
		t.Error(err)
	}

	t.Log(parsedRes)

	//execute state-altering testAsync method

	callData2, err := abi.Pack("testAsync", big.NewInt(10))
	if err != nil {
		t.Fatal(err)
	}

	tx2, err := test.prepareTransaction(&from,
		&accounts.Account{Address: contractAddress},
		value,
		gas,
		gasPrice,
		common.ToHex(callData2))

	if err != nil {
		t.Fatal(err)
	}

	//convert to raw bytes
	data2, err := rlp.EncodeToBytes(tx2)
	if err != nil {
		t.Fatal(err)
	}

	//try to append tx
	err = test.state.AppendTx(data2)
	if err != nil {
		t.Fatal(err)
	}

	err = test.state.Commit()
	if err != nil {
		t.Fatal(err)
	}

	receipt2, err := test.state.GetReceipt(tx2.Hash())
	if err != nil {
		t.Fatal(err)
	}

	t.Log(receipt2)

}
