package evmbabble

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type AccountMap map[string]struct {
	Code    string
	Storage map[string]string
	Balance string
}
type JsonAccount struct {
	Address string
	Balance *big.Int
}

type JsonAccountList struct {
	Accounts []JsonAccount
}

// SendTxArgs represents the arguments to sumbit a new transaction into the transaction pool.
type SendTxArgs struct {
	From     common.Address  `json:"from"`
	To       *common.Address `json:"to"`
	Gas      *big.Int        `json:"gas"`
	GasPrice *big.Int        `json:"gasPrice"`
	Value    *big.Int        `json:"value"`
	Data     string          `json:"data"`
	Nonce    *uint64         `json:"nonce"`
}
