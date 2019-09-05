package txnvalidating

import (
	"strings"

	ethereum "github.com/monetha/go-ethereum"
)

func getTxHashesByStatus(txs ethereum.Transactions, status ethereum.TransactionStatus) []string {
	txByHash := make(transactionsByHash)
	for _, tx := range txs {
		if tx == nil {
			continue
		}

		if *tx.Status == status {
			txHash := strings.ToLower(tx.Hash.Hex())
			txByHash[txHash] = tx
		}
	}

	keys := make([]string, len(txByHash))
	i := 0
	for txHash := range txByHash {
		keys[i] = txHash
		i++
	}

	return keys
}
