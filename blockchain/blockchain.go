package blockchain

import (
	"github.com/boltdb/bolt"
)

type Blockchain struct {
	lastHash []byte
	db       *bolt.DB
}

const dbFile = "./db/blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "Testing"

func NewBlockChain(address string) *Blockchain {
	var lastHash []byte

	db, err := bolt.Open(dbFile, 0600, nil)
	HandleError(err)
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			baseTx := NewCoinBaseTX(address, genesisCoinbaseData)
			genesis := GenerateGenesisBlock(baseTx)
			b, _ := tx.CreateBucket([]byte(blocksBucket))
			err = b.Put(genesis.Hash, genesis.Serialize())
			err = b.Put([]byte("l"), genesis.Hash)
			lastHash = genesis.Hash
		} else {
			lastHash = b.Get([]byte("l"))
		}
		return nil
	})
	b := &Blockchain{lastHash: lastHash, db: db}
	return b
}

func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	var lastHash []byte
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	HandleError(err)
	newBlock := NewBlock(transactions, lastHash)

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

func (bc *Blockchain) Iterator() *BlockChainIterator {
	bci := &BlockChainIterator{bc.lastHash, bc.db}
	return bci
}

func (bc *Blockchain) Close() {
	err := bc.db.Close()
	if err != nil {
		panic(err)
		return
	}
}
