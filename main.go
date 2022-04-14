package main

import (
	"blockchaincore/blockchain"
	"fmt"
	"strconv"
)

func main() {
	bc := blockchain.NewBlockChain()
	bc.AddBlock("Send 1 BTC to Ivan")
	bc.AddBlock("Send 2 more BTC to Ivan")

	for _, block := range bc.Blocks {
		block.Print()
		pow := blockchain.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Print("\n")
	}

}
