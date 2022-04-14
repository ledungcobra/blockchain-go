package main

import (
	"blockchaincore/blockchain"
)

func main() {
	bc := blockchain.NewBlockChain()
	defer bc.Close()

	cli := blockchain.NewCLI(bc)
	cli.Run()
}
