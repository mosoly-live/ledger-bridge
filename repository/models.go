package repository

const (
	// TxnInProgress is transaction status - in progress
	TxnInProgress = iota + 1
	// TxnSuccessful is transaction status - successful completed
	TxnSuccessful
	// TxnFailed is transaction status - failed
	TxnFailed
)
