package evmbabble

import (
	"encoding/json"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

func accountsHandler(w http.ResponseWriter, r *http.Request, m *Service) {
	m.logger.Debug("GET accounts")

	state, err := m.getState()
	if err != nil {
		m.logger.WithError(err).Error("Getting State")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var al JsonAccountList

	for _, account := range m.keyStore.Accounts() {
		balance := state.GetBalance(account.Address)
		al.Accounts = append(al.Accounts,
			JsonAccount{
				Address: account.Address.Hex(),
				Balance: balance,
			})
	}

	js, err := json.Marshal(al)
	if err != nil {
		m.logger.WithError(err).Error("Marshaling JSON response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func transactionHandler(w http.ResponseWriter, r *http.Request, m *Service) {
	m.logger.WithField("request", r).Debug("POST tx")

	state, err := m.getState()
	if err != nil {
		m.logger.WithError(err).Error("Getting State")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var txArgs SendTxArgs
	err = decoder.Decode(&txArgs)
	if err != nil {
		m.logger.WithError(err).Error("Decoding JSON txArgs")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	tx, err := prepareTransaction(txArgs, state, m.keyStore)
	if err != nil {
		m.logger.WithError(err).Error("Preparing Transaction")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		m.logger.WithError(err).Error("Encoding Transaction")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = m.proxy.SubmitTransaction(data)
	if err != nil {
		m.logger.WithError(err).Error("Submitting Transaction")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := struct{ TxHash string }{TxHash: tx.Hash().Hex()}
	js, err := json.Marshal(res)
	if err != nil {
		m.logger.WithError(err).Error("Marshalling JSON response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

func transactionReceiptHandler(w http.ResponseWriter, r *http.Request, m *Service) {
	param := r.URL.Path[len("/tx/"):]
	txHash := common.HexToHash(param)
	m.logger.WithField("tx_hash", txHash.Hex()).Debug("GET tx")

	state, err := m.getState()
	if err != nil {
		m.logger.WithError(err).Error("Getting State")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tx, err := state.GetTransaction(txHash)
	if err != nil {
		m.logger.WithError(err).Error("Getting Transaction")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	receipt, err := state.GetReceipt(txHash)
	if err != nil {
		m.logger.WithError(err).Error("Getting Receipt")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	signer := ethTypes.NewEIP155Signer(big.NewInt(1))
	from, err := ethTypes.Sender(signer, tx)
	if err != nil {
		m.logger.WithError(err).Error("Getting Tx Sender")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fields := map[string]interface{}{
		"root":              common.BytesToHash(receipt.PostState),
		"transactionHash":   txHash,
		"from":              from,
		"to":                tx.To(),
		"gasUsed":           receipt.GasUsed,
		"cumulativeGasUsed": receipt.CumulativeGasUsed,
		"contractAddress":   nil,
		"logs":              receipt.Logs,
		"logsBloom":         receipt.Bloom,
	}
	if receipt.Logs == nil {
		fields["logs"] = [][]*ethTypes.Log{}
	}
	// If the ContractAddress is 20 0x0 bytes, assume it is not a contract creation
	if receipt.ContractAddress != (common.Address{}) {
		fields["contractAddress"] = receipt.ContractAddress
	}

	js, err := json.Marshal(fields)
	if err != nil {
		m.logger.WithError(err).Error("Marshaling JSON response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

//------------------------------------------------------------------------------

func prepareTransaction(args SendTxArgs, state *State, ks *keystore.KeyStore) (*ethTypes.Transaction, error) {
	var err error
	args, err = prepareSendTxArgs(args)
	if err != nil {
		return nil, err
	}

	if args.Nonce == nil {
		args.Nonce = new(uint64)
		*args.Nonce = state.GetNonce(args.From)
	}

	var tx *ethTypes.Transaction
	if args.To == nil {
		tx = ethTypes.NewContractCreation(*args.Nonce,
			args.Value,
			args.Gas,
			args.GasPrice,
			common.FromHex(args.Data))
	} else {
		tx = ethTypes.NewTransaction(*args.Nonce,
			*args.To,
			args.Value,
			args.Gas,
			args.GasPrice,
			common.FromHex(args.Data))
	}

	signer := ethTypes.NewEIP155Signer(big.NewInt(1))

	account, err := ks.Find(accounts.Account{Address: args.From})
	if err != nil {
		return nil, err
	}
	signature, err := ks.SignHash(account, signer.Hash(tx).Bytes())
	if err != nil {
		return nil, err
	}
	signedTx, err := tx.WithSignature(signer, signature)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

func prepareSendTxArgs(args SendTxArgs) (SendTxArgs, error) {
	if args.Gas == nil {
		args.Gas = defaultGas
	}
	if args.GasPrice == nil {
		args.GasPrice = big.NewInt(0)
	}
	if args.Value == nil {
		args.Value = big.NewInt(0)
	}
	return args, nil
}
