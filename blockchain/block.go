package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"strconv"
	"time"
)

type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

func (b *Block) GenerateHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

// Create new block by running the proof of work algorithm

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: prevBlockHash,
	}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

func GenerateGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

func (b *Block) Print() {
	fmt.Printf("Prev. hash: %x\n", b.PrevBlockHash)
	fmt.Printf("Data: %s\n", b.Data)
	fmt.Printf("Hash: %x\n", b.Hash)
	fmt.Println()
}

func (b *Block) Serialize() []byte {
	var r bytes.Buffer
	encoder := gob.NewEncoder(&r)
	err := encoder.Encode(b)
	if err != nil {
		fmt.Println("Error serializing block chain", err)
	}
	return r.Bytes()
}

func Deserialize(data []byte) *Block {
	var r Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&r)
	if err != nil {
		fmt.Println("Error deserializing block chain", err)
	}
	return &r
}
