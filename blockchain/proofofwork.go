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
const maxNonce = math.MaxInt64

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork Create new proof of work
func NewProofOfWork(block *Block) *ProofOfWork {
	pow := &ProofOfWork{block: block}
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	pow.target = target
	return pow
}

// prepareData Prepare data for utils data contains PreviousBlockHash, HashTransactions, CurrentTimeStamp, targetBits, nonce
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

// IntToHex convert int to hex
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		panic(err)
	}
	return buff.Bytes()
}

// Run runs the proof of work
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

		// Convert utils to big integer
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

// Validate validates utils of block bellow the target
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	return hashInt.Cmp(pow.target) == -1 // hashInt < target
}
