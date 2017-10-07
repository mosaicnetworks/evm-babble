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
pragma solidity ^0.4.0;

contract Simple {
    function Foo() returns (uint) {
        return 1;
    }
}
*/
//Compiled with solc v0.4.8 commit 60cc1668..
//the version of evm we are using does not support some bytecodes generated by
//later versions of solc
func dummyContract() Contract {
	return Contract{
		name: "Dummy",
		code: "6060604052346000575b6092806100176000396000f30060606040526000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff168063bfb4ebcf14603c575b6000565b346000576046605c565b6040518082815260200191505060405180910390f35b6000600190505b905600a165627a7a7230582020668019c3efbd40820d7161a0b7adc5b392ea0051dcc9626b3f3946aa02f2400029",
		abi:  "[{\"constant\":false,\"inputs\":[],\"name\":\"Foo\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\",\"stateMutability\":\"nonpayable\"}]",
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

	callData, err := abi.Pack("Foo")
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
}
