package blockchain

import (
	. "blockchaincore/hash"
	"flag"
	"fmt"
	"log"
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

	printChainDebugCmd := flag.NewFlagSet("printchaindebug", flag.ExitOnError)

	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)

	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
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
	case "printchaindebug":
		err := printChainDebugCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}

	default:
		cli.printUsage()
		os.Exit(1)
	}

	switch {
	case printChainCmd.Parsed():
		cli.printChain()
	case getBalanceCmd.Parsed():
		if *addressFlag == "" {
			fmt.Println("An error occur")
			os.Exit(1)
		}
		cli.getBalance(*addressFlag)
	case sendCmd.Parsed():
		if *toFlag == "" || *fromFlag == "" || *amountFlag == 0 {
			fmt.Println("An error occur")
			os.Exit(1)
		}
		cli.send(*fromFlag, *toFlag, *amountFlag)

	case createBlockChainCmd.Parsed():
		if *createBlockChainAddressFlag == "" {
			fmt.Println("An error occur")
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockChainAddressFlag)
	case printChainDebugCmd.Parsed():
		cli.printChainDebug()
	case createWalletCmd.Parsed():
		cli.createWallet()
	case listAddressesCmd.Parsed():
		cli.listAddresses()
	}

}

func (cli *CLI) createBlockchain(address string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
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
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  listaddresses - Lists all addresses from the wallet file")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}

func (cli *CLI) getBalance(address string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}

	bc := NewBlockChain(address)
	defer bc.Close()

	balance := 0
	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]

	UTXOs := bc.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)

}

func (cli *CLI) send(from, to string, amount int) {

	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}

	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := NewBlockChain(from)
	defer bc.Close()

	tx := NewUTXOTransaction(from, to, amount, bc)
	bc.MineBlock([]*Transaction{tx})
	fmt.Println("Success!")
}

func (cli *CLI) printChainDebug() {
	bc := NewBlockChain("")
	defer bc.Close()
	bci := bc.Iterator()

	for {
		block := bci.Next()
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %+v\n", block.Hash)
		fmt.Printf("Transaction: ")
		for _, tx := range block.Transactions {
			fmt.Printf("%s\n", tx.ToString())
		}
		fmt.Printf("Timestamp: %+v\n", block.Timestamp)
		fmt.Printf("Nonce: %+v\n", block.Nonce)
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) createWallet() {

	wallet, err := NewWallets()
	if err != nil {
		log.Printf("%v", err)
	}
	address := wallet.CreateWallet()
	wallet.SaveToFile()

	fmt.Println("Your address: ", address)
}

func (cli *CLI) listAddresses() {
	wallet, err := NewWallets()
	if err != nil {
		log.Printf("%v", err)
	}
	addresses := wallet.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}
