package blockchain

import (
	"github.com/boltdb/bolt"
)

type BlockChain struct {
	lastHash []byte
	db       *bolt.DB
	Blocks   []*Block
}

const dbFile = "blockchain.db"
const blocksBucket = "blocks"

func NewBlockChain() *BlockChain {
	var lastHash []byte

	db, err := bolt.Open(dbFile, 0600, nil)
	HandleError(err)
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			genesis := GenerateGenesisBlock()
			b, _ := tx.CreateBucket([]byte(blocksBucket))
			err = b.Put(genesis.Hash, genesis.Serialize())
			err = b.Put([]byte("l"), genesis.Hash)
			lastHash = genesis.Hash
		} else {
			lastHash = b.Get([]byte("l"))
		}
		return nil
	})
	b := &BlockChain{lastHash: lastHash, db: db}
	return b
}

func (bc *BlockChain) AddBlock(data string) {
	var lastHash []byte
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	HandleError(err)
	newBlock := NewBlock(data, lastHash)

	// Store new block to local database
	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		err = b.Put([]byte("l"), newBlock.Hash)
		bc.lastHash = newBlock.Hash
		HandleError(err)
		return nil
	})
}

func HandleError(e error) {
	if e != nil {
		panic(e)
	}
}

func (bc *BlockChain) Iterator() *BlockChainIterator {
	bci := &BlockChainIterator{bc.lastHash, bc.db}
	return bci
}
