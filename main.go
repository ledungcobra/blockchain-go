package main

import "blockchaincore/blockchain"

func main() {
	cli := blockchain.NewCLI()
	cli.Run()
}
