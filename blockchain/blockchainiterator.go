package blockchain

import "github.com/boltdb/bolt"

type BlockChainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func (iter *BlockChainIterator) Next() *Block {
	var block *Block
	iter.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(iter.currentHash)
		block = Deserialize(encodedBlock)
		return nil
	})
	iter.currentHash = block.PrevBlockHash
	return block
}
