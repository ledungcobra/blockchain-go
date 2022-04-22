package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"time"
)

type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

// NewBlock Create new block by running the proof of work algorithm
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  transactions,
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

// NewGenesisBlock create the genesis block for the blockchain
func NewGenesisBlock(coinBaseTx *Transaction) *Block {
	return NewBlock([]*Transaction{coinBaseTx}, []byte{})
}

// Print prints the block in debug mode
func (b *Block) Print() {
	fmt.Printf("Prev. utils: %x\n", b.PrevBlockHash)
	fmt.Printf("Data: %v\n", b.Transactions)
	fmt.Printf("Hash: %x\n", b.Hash)
	fmt.Println()
}

// Serialize serializes the block
func (b *Block) Serialize() []byte {
	var r bytes.Buffer
	encoder := gob.NewEncoder(&r)
	err := encoder.Encode(b)
	if err != nil {
		fmt.Println("Error serializing block chain", err)
	}
	return r.Bytes()
}

// HashTransactions utils transactions by combine all utils transactions in block
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.Hash())
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}

// Deserialize deserializes the block
func Deserialize(data []byte) *Block {
	var r Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&r)
	if err != nil {
		fmt.Println("Error deserializing block chain", err)
	}
	return &r
}
