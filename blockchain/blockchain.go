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
)

type Blockchain struct {
	lastHash []byte
	db       *bolt.DB
}

const dbFile = "./db/blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "Testing"

var TransactionNotFoundError = errors.New("transaction not found")

// NewBlockChain creates a new Blockchain with genesis Block and reward coinbase transaction to the first miner
func NewBlockChain(address string) *Blockchain {

	var lastHash []byte

	db, err := bolt.Open(dbFile, 0600, nil)
	utils.HandleError(err)
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

// MineBlock mine a block by adding new transactions to a new created block
func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
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

	utils.HandleError(err)

	// create new block and do proof of work
	newBlock := NewBlock(transactions, lastHash)

	// Store new block to local database
	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		err = b.Put([]byte("l"), newBlock.Hash)
		bc.lastHash = newBlock.Hash
		utils.HandleError(err)
		return nil
	})
	return newBlock
}

// Iterator create new iterator to traverse the blockchain
func (bc *Blockchain) Iterator() *BlockChainIterator {
	bci := &BlockChainIterator{bc.lastHash, bc.db}
	return bci
}

// Close the underlying database of blockchain
func (bc *Blockchain) Close() {
	err := bc.db.Close()
	if err != nil {
		panic(err)
		return
	}
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
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	return prevTXs
}

// VerifyTransaction verify transaction
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
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

// FindUTXOByPubHashKey Is used for get the balance of the public key
func (bc *Blockchain) FindUTXOByPubHashKey(pubHashKey []byte) []TXOutput {

	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(pubHashKey)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubHashKey) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
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
				// Was the output spent ?

				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							// Skip that output since it was already referenced in an input
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if tx.IsCoinbase() { // Skip the coinbase transaction
				continue
			}

			for _, in := range tx.Vin {
				inTxID := hex.EncodeToString(in.Txid)
				spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return UTXO
}
