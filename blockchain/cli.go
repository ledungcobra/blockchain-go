package blockchain

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type CLI struct {
}

func NewCLI() *CLI {
	return &CLI{}
}

func (cli *CLI) Run() {
	cli.ValidateArgs()
	createBlockChainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createBlockChainAddressFlag := createBlockChainCmd.String("address", "", "The address to send genesis block reward to")

	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)

	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	fromFlag := sendCmd.String("from", "", "Source wallet address")
	toFlag := sendCmd.String("to", "", "Destination wallet address")
	amountFlag := sendCmd.Int("amount", 0, "Amount to send")

	addressFlag := getBalanceCmd.String("address", "", "The address to get balance for")
	switch os.Args[1] {
	case "createblockchain":
		err := createBlockChainCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if getBalanceCmd.Parsed() {
		if *addressFlag == "" {
			fmt.Println("An error occur")
			os.Exit(1)
		}
		cli.getBalance(*addressFlag)
	}

	if sendCmd.Parsed() {
		if *toFlag == "" || *fromFlag == "" || *amountFlag == 0 {
			fmt.Println("An error occur")
			os.Exit(1)
		}
		cli.send(*fromFlag, *toFlag, *amountFlag)
	}

	if createBlockChainCmd.Parsed() {
		if *createBlockChainAddressFlag == "" {
			fmt.Println("An error occur")
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockChainAddressFlag)
	}

}

func (cli *CLI) createBlockchain(address string) {
	bc := NewBlockChain(address)
	defer bc.db.Close()
	fmt.Println("Done!")
}

func (cli *CLI) printChain() {
	bc := NewBlockChain("")
	defer bc.Close()
	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) ValidateArgs() bool {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./blockchain-cli [addblock]")
		os.Exit(1)
	}
	return true
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  addblock -data BLOCK_DATA - add a block to the blockchain")
	fmt.Println("  printchain - print all the blocks of the blockchain")
}

func (cli *CLI) getBalance(address string) {
	bc := NewBlockChain(address)
	defer bc.Close()

	balance := 0
	UTXOs := bc.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)

}

func (cli *CLI) send(from, to string, amount int) {
	bc := NewBlockChain(from)
	defer bc.Close()

	tx := NewUTXOTransaction(from, to, amount, bc)
	bc.MineBlock([]*Transaction{tx})
	fmt.Println("Success!")
}
