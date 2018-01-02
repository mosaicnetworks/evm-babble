package state

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethState "github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/sirupsen/logrus"
)

// write ahead state, updated with each AppendTx
// and reset on Commit
type WriteAheadState struct {
	db       ethdb.Database
	ethState *ethState.StateDB

	txIndex      int
	transactions []*ethTypes.Transaction
	receipts     []*ethTypes.Receipt
	allLogs      []*ethTypes.Log

	totalUsedGas *big.Int
	gp           *core.GasPool

	logger *logrus.Logger
}

func (was *WriteAheadState) Commit() (common.Hash, error) {
	//commit all state changes to the database
	hashArray, err := was.ethState.CommitTo(was.db, true)
	if err != nil {
		was.logger.WithError(err).Error("Committing state")
		return common.Hash{}, err
	}
	if err := was.writeHead(); err != nil {
		was.logger.WithError(err).Error("Writing head")
		return common.Hash{}, err
	}
	if err := was.writeTransactions(); err != nil {
		was.logger.WithError(err).Error("Writing txs")
		return common.Hash{}, err
	}
	if err := was.writeReceipts(); err != nil {
		was.logger.WithError(err).Error("Writing receipts")
		return common.Hash{}, err
	}
	return hashArray, nil
}

func (was *WriteAheadState) writeHead() error {
	head := &ethTypes.Transaction{}
	if len(was.transactions) > 0 {
		head = was.transactions[len(was.transactions)-1]
	}
	return was.db.Put(headTxKey, head.Hash().Bytes())
}

func (was *WriteAheadState) writeTransactions() error {
	batch := was.db.NewBatch()

	for _, tx := range was.transactions {
		data, err := rlp.EncodeToBytes(tx)
		if err != nil {
			return err
		}
		if err := batch.Put(tx.Hash().Bytes(), data); err != nil {
			return err
		}
	}

	// Write the scheduled data into the database
	return batch.Write()
}

func (was *WriteAheadState) writeReceipts() error {
	batch := was.db.NewBatch()

	for _, receipt := range was.receipts {
		storageReceipt := (*ethTypes.ReceiptForStorage)(receipt)
		data, err := rlp.EncodeToBytes(storageReceipt)
		if err != nil {
			return err
		}
		if err := batch.Put(append(receiptsPrefix, receipt.TxHash.Bytes()...), data); err != nil {
			return err
		}
	}

	return batch.Write()
}
