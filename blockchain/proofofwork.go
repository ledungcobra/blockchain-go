package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
)

const targetBits = 16

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func NewProofOfWork(block *Block) *ProofOfWork {
	pow := &ProofOfWork{block: block}
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	pow.target = target
	return pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	return bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{})
}

func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		panic(err)
	}
	return buff.Bytes()
}

const maxNonce = math.MaxInt64

func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte

	nonce := 0
	fmt.Printf("Mining the block with data %v\n", pow.block.Transactions)

	for nonce < maxNonce {
		// Prepare data from block and nonce
		data := pow.prepareData(nonce)

		// Hash data with sha256
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)

		// Convert hash to big integer
		hashInt.SetBytes(hash[:])

		// Compare hashInt and target
		if hashInt.Cmp(pow.target) == -1 { // hashInt < target
			break
		} else { // hashInt >= target
			nonce++
		}
	}
	fmt.Print("\n\n")
	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	return hashInt.Cmp(pow.target) == -1 // hashInt < target
}