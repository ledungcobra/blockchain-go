package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/boltdb/bolt"
	"log"
)

type Blockchain struct {
	lastHash []byte
	db       *bolt.DB
}

const dbFile = "./db/blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "Testing"

func NewBlockChain(address string) *Blockchain {

	//if dbExists() == false {
	//	fmt.Println("No existing blockchain found. Create one first.")
	//	os.Exit(1)
	//}

	var lastHash []byte

	db, err := bolt.Open(dbFile, 0600, nil)
	HandleError(err)
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			baseTx := NewCoinBaseTX(address, genesisCoinbaseData)
			genesis := NewGenesisBlock(baseTx)
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

func dbExists() bool {
	return true
}

func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	var lastHash []byte

	for _, tx := range transactions {
		if bc.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid transaction")
		}
	}

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

func (bc *Blockchain) SignTransaction(t *Transaction, key ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)
	for _, vin := range t.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	t.Sign(key, prevTXs)
}

func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}

func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inID].Signature = signature
	}
}
