package txnvalidating

import (
	"io"
	"math/big"

	ethereum "github.com/monetha/go-ethereum"
)

type transactionsByHash map[string]*ethereum.Transaction

// BlockSource delivers blocks from Ethereum block-chain.
type BlockSource interface {
	io.Closer
	// Blocks returns the channel on which the blocks are delivered.
	Blocks() <-chan *ethereum.Block
}

// BlockSourceCreator creates an instance of BlockSource.
type BlockSourceCreator interface {
	// CreateBlockSource creates an instance of BlockSource.
	// startBlock is the number of the block from which to start the delivery of blocks.
	// If number is nil, the latest known block that has specified number of confirmations
	// is used as start block.
	// confirmations number indicates that the block must be delivered only when it has
	// the specified number of confirmations (number of blocks mined since delivered block).
	CreateBlockSource(startBlock *big.Int, confirmations uint) (BlockSource, error)
}
