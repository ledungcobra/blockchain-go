package blockchain

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"
)

type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash Hash
	Hash          Hash
	Nonce         int
	Height        int
}

type Hash = []byte

// NewBlock Create new block by running the proof of work algorithm
func NewBlock(transactions []*Transaction, prevBlockHash Hash, height int) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  transactions,
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
		Height:        height,
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

// NewGenesisBlock create the genesis block for the blockchain
func NewGenesisBlock(coinBaseTx *Transaction) *Block {
	return NewBlock([]*Transaction{coinBaseTx}, []byte{}, 0)
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
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
	}
	mTree := NewMerkleTree(transactions)
	return mTree.RootNode.Data
}

// DeserializeBlock deserializes the block
func DeserializeBlock(data []byte) *Block {
	var r Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&r)
	if err != nil {
		fmt.Println("Error deserializing block chain", err)
	}
	return &r
}
