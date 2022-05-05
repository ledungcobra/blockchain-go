package cli

import (
	"blockchaincore/blockchain"
	"blockchaincore/p2pserver"
	"blockchaincore/utils"
	"blockchaincore/web"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
)

type CLI struct {
}

func NewCLI() *CLI {
	return &CLI{}
}
func (cli *CLI) ValidateArgs() bool {
	if len(os.Args) < 2 {
		cli.printUsage()
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
	fmt.Println("  reindexutxo - Rebuilds the UTXO set")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT -mine - Send AMOUNT of coins from FROM address to TO. Mine on the same node, when -mine is set.")
	fmt.Println("  startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining")
}

func (cli *CLI) Run() {
	cli.ValidateArgs()

	// Node id is node port
	nodeID := os.Getenv("NODE_ID")

	if nodeID == "" {
		cli.printUsage()
		fmt.Println("NODE_ID is not set!")
		os.Exit(1)
	}

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)
	runWebCmd := flag.NewFlagSet("runweb", flag.ExitOnError)
	clearBlockChainCmd := flag.NewFlagSet("clear", flag.ExitOnError)
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)

	// Flags
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
	startNodeMiner := startNodeCmd.String("miner", "", "Enable mining mode and send reward to ADDRESS")
	syncBlockChainCmd := flag.NewFlagSet("sync", flag.ExitOnError)
	portStartWebServer := runWebCmd.String("port", "8080", "Port to start web server on")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "reindexutxo":
		err := reindexUTXOCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "sync":
		err := syncBlockChainCmd.Parse(os.Args[1:])
		if err != nil {
			log.Panic(err)
		}
	case "runweb":
		err := runWebCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "clear":
		err := clearBlockChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "init":
		err := initCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	default:
		cli.printUsage()
		os.Exit(1)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress, nodeID)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress, nodeID)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(nodeID)
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses(nodeID)
	}

	if printChainCmd.Parsed() {
		cli.printChain(nodeID)
	}

	if reindexUTXOCmd.Parsed() {
		cli.reindexUTXO(nodeID)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendFrom, *sendTo, *sendAmount, nodeID, *sendMine)
	}

	if startNodeCmd.Parsed() {
		cli.startNode(nodeID, *startNodeMiner)
	}
	if syncBlockChainCmd.Parsed() {
		cli.SynBlockChain()
	}

	if runWebCmd.Parsed() {
		log.Println("Start web")
		web.StartWebServer(*portStartWebServer)
	}

	if clearBlockChainCmd.Parsed() {
		cli.ClearBlockChain()
	}

	if initCmd.Parsed() {
		cli.InitBlockChain()
	}

}

func (cli *CLI) getBalance(address string, nodeID string) int {
	if !blockchain.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := blockchain.NewBlockchain(nodeID)
	UTXOSet := blockchain.UTXOSet{bc}
	defer bc.Db.Close()

	balance := 0
	pubKeyHash := utils.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
	return balance
}

func (cli *CLI) printChain(nodeID string) {
	bc := blockchain.NewBlockchain(nodeID)
	defer bc.Db.Close()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		pow := blockchain.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) createBlockchain(address string, nodeID string) {
	if !blockchain.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := blockchain.CreateBlockchain(address, nodeID)
	defer bc.Db.Close()

	UTXOSet := blockchain.UTXOSet{bc}
	UTXOSet.Reindex()

	fmt.Println("Done!")
}

func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !blockchain.ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !blockchain.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := blockchain.NewBlockchain(nodeID)
	UTXOSet := blockchain.UTXOSet{bc}
	defer bc.Db.Close()

	wallets, err := blockchain.NewWallets(nodeID)
	utils.HandleError(err)
	// TODO: using private key to get wallet
	wallet := wallets.GetWallet(from)
	if wallet == nil {
		log.Println("ERROR: Sender address is not found in wallet file")
		return
	}

	tx := blockchain.NewUTXOTransaction(wallet, to, amount, &UTXOSet)

	if mineNow {
		cbTx := blockchain.NewCoinbaseTX(from, "", 0)
		txs := []*blockchain.Transaction{cbTx, tx}
		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		log.Println("Sending tx to the network...")
		p2pserver.SendTx(p2pserver.CentralNode, tx)
		log.Println("Sent tx to transaction pools")
	}

	log.Println("Success!")

}

func (cli *CLI) createWallet(nodeID string) {
	wallets, _ := blockchain.NewWallets(nodeID)
	address, pri, pub := wallets.CreateWallet()
	wallets.SaveToFile(nodeID)

	log.Printf("Your new wallet:\n address: %s\nprivate key: %s, public key: %s", address, pri, pub)
}

func (cli *CLI) reindexUTXO(nodeID string) {
	bc := blockchain.NewBlockchain(nodeID)
	UTXOSet := blockchain.UTXOSet{bc}
	UTXOSet.Reindex()

	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}

func (cli *CLI) listAddresses(nodeID string) {
	wallet, err := blockchain.NewWallets(nodeID)
	if err != nil {
		log.Printf("%v", err)
	}
	addresses := wallet.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CLI) startNode(nodeID, minerAddress string) {
	fmt.Printf("Starting node %s\n", nodeID)
	if len(minerAddress) > 0 {
		if blockchain.ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	p2pserver.StartServer(nodeID, minerAddress)
}

func (cli *CLI) SynBlockChain() {
	nodePort := os.Getenv("NODE_ID")
	if nodePort == "" {
		log.Panic("NODE_ID not set")
	}
	myAddress := "localhost:" + nodePort
	ln, err := net.Listen("tcp", myAddress)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()
	p2pserver.GetBlockFromCentralNode(ln)
}

func (cli *CLI) ClearBlockChain() {
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		log.Panic("NODE_ID not set")
	}
	err := os.Remove(fmt.Sprintf(blockchain.DbFile, nodeID))
	if err != nil {
		log.Panic(err)
	}
	log.Println("Blockchain is cleared")
	err = os.Remove(fmt.Sprintf(blockchain.WalletFile, nodeID))
	if err != nil {
		log.Panic(err)
		return
	}
}

func (cli *CLI) InitBlockChain() {

	nodeID := os.Getenv("NODE_ID")
	wallets, _ := blockchain.NewWallets(nodeID)
	address, pri, pub := wallets.CreateWallet()
	wallets.SaveToFile(nodeID)

	cli.createBlockchain(address, nodeID)
	balance := cli.getBalance(address, nodeID)
	log.Printf("Your \nNew address: %s\n pubKey: %s\n, priKey: %s\n and be reward %d",
		address, pub, pri, balance)

	data, _ := json.Marshal(utils.WalletData{
		Address:    address,
		PrivateKey: pri,
		PublicKey:  pub,
	})
	ioutil.WriteFile("wallet1.json", data, 0644)

	address, pri, pub = wallets.CreateWallet()

	data, _ = json.Marshal(utils.WalletData{
		Address:    address,
		PrivateKey: pri,
		PublicKey:  pub,
	})
	ioutil.WriteFile("wallet2.json", data, 0644)
	address, pri, pub = wallets.CreateWallet()
	data, _ = json.Marshal(utils.WalletData{
		Address:    address,
		PrivateKey: pri,
		PublicKey:  pub,
	})
	ioutil.WriteFile("wallet3.json", data, 0644)

}
