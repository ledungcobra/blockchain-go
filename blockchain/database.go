package blockchain

import "github.com/boltdb/bolt"

type BlockChainDB interface {
	Open() error
	Close() error
	Update(callback func(tx *bolt.Tx) error) error
	Put(key []byte, value []byte) error
	View(callback func(tx *bolt.Tx) error) error
}

type MyDb struct {
	*bolt.DB
}

func (m MyDb) Open() error {
	return nil
}

func (m MyDb) Put(key []byte, value []byte) error {

	return nil
}
