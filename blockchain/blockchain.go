package blockchain

import (
	"blockchaincore/utils"
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

type Blockchain struct {
	lastHash []byte
	Db       *bolt.DB
}

const DbFile = "./db/blockchain_%s.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "Testing"

var TransactionNotFoundError = errors.New("transaction not found")

func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		log.Println("Database file not exist: ", dbFile)
		return false
	}
	return true
}

// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(address, nodeID string) *Blockchain {

	file := fmt.Sprintf(DbFile, nodeID)
	if dbExists(file) {
		fmt.Println("Blockchain already exists.")
		utils.HandleError(errors.New("blockchain already exists"))
	}

	var tip []byte

	cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
	genesis := NewGenesisBlock(cbtx)
	db, err := bolt.Open(file, 0600, nil)
	utils.HandleError(err)

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		utils.HandleError(err)

		err = b.Put(genesis.Hash, genesis.Serialize())
		utils.HandleError(err)
		err = b.Put([]byte("l"), genesis.Hash)
		utils.HandleError(err)
		tip = genesis.Hash

		return nil
	})
	utils.HandleError(err)

	bc := Blockchain{tip, db}
	return &bc
}

// NewBlockchain creates a new Blockchain with genesis Block and reward coinbase transaction to the first miner
func NewBlockchain(nodeID string) *Blockchain {
	file := fmt.Sprintf(DbFile, nodeID)

	if dbExists(file) == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(file, 0600, nil)
	utils.HandleError(err)
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})
	utils.HandleError(err)
	bc := Blockchain{tip, db}

	return &bc
}

// MineBlock mine a block by adding new transactions to a new created block
func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	for _, tx := range transactions {
		if !bc.VerifyTransaction(tx) {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	err := bc.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		block := DeserializeBlock(blockData)

		lastHeight = block.Height
		return nil
	})

	utils.HandleError(err)

	// create new block and do proof of work
	newBlock := NewBlock(transactions, lastHash, lastHeight+1)

	// Store new block to local database
	err = bc.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		err = b.Put([]byte("l"), newBlock.Hash)
		bc.lastHash = newBlock.Hash
		utils.HandleError(err)
		return nil
	})
	return newBlock
}

func (bc *Blockchain) AddBlock(block *Block) {
	err := bc.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockInDb := b.Get(block.Hash)

		if blockInDb != nil {
			return nil
		}

		blockData := block.Serialize()
		err := b.Put(block.Hash, blockData)
		utils.HandleError(err)

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := DeserializeBlock(lastBlockData)

		if block.Height > lastBlock.Height {
			err = b.Put([]byte("l"), block.Hash)
			utils.HandleError(err)
			bc.lastHash = block.Hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

}

func (bc *Blockchain) GetBestHeight() int {
	var lastBlock Block

	err := bc.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		lastBlock = *DeserializeBlock(blockData)

		return nil
	})
	utils.HandleError(err)
	return lastBlock.Height
}

func (bc *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := bc.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("Block is not found.")
		}

		block = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

// Iterator create new iterator to traverse the blockchain
func (bc *Blockchain) Iterator() *BlockChainIterator {
	bci := &BlockChainIterator{bc.lastHash, bc.Db}
	return bci
}

func (bc *Blockchain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.Iterator()

	for {
		block := bci.Next()

		blocks = append(blocks, block.Hash)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return blocks
}

// Close the underlying database of blockchain
func (bc *Blockchain) Close() {
	err := bc.Db.Close()
	utils.HandleError(err)
}

// SignTransaction Sign transaction by private key
func (bc *Blockchain) SignTransaction(t *Transaction, key ecdsa.PrivateKey) error {

	// key is transaction id mapping to transaction
	prevTXs := make(map[string]Transaction)

	for _, vin := range t.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			return fmt.Errorf("%w Not found transaction by id", err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	t.Sign(key, prevTXs)
	return nil
}

// FindTransaction find transaction by transaction id
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		// Hit the genesis block
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, TransactionNotFoundError
}

// FindPreviousTransactions Find previous transaction related to the current transaction
func (bc *Blockchain) FindPreviousTransactions(tx *Transaction) map[string]Transaction {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		utils.HandleError(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	return prevTXs
}

// VerifyTransaction verify transaction
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	prevTXs := bc.FindPreviousTransactions(tx)
	return tx.Verify(prevTXs)
}

// FindUnspentTransactions Find unspent transactions by address
func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTxs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

	Transactions:
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent ?

				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							// Skip that output since it was already referenced in an input
							continue Outputs
						}
					}
				}
				// Check for the OutputTx is belonged to address
				if out.IsLockedWithKey(pubKeyHash) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}

			if tx.IsCoinbase() { // Skip the coinbase transaction
				continue Transactions
			}

			for _, in := range tx.Vin {
				if in.UsesKey(pubKeyHash) {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		// Break if that is genesis block
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return unspentTxs
}

// FindUTXO Find all unspent transaction outputs in blockchain
func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXO
}
