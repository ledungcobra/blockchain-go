package p2pserver

import (
	"blockchaincore/blockchain"
	"fmt"
	"github.com/vrecan/death"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"
	"syscall"
)

var doneWritingBlockChain = make(chan bool)

func StartServer(nodeID, minerAddr string) {
	myAddress = fmt.Sprintf("localhost:%s", nodeID)
	mineAddr = minerAddr
	ln, err := net.Listen(protocol, myAddress)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()
	file := fmt.Sprintf(blockchain.DbFile, nodeID)

	if _, err := os.Stat(file); os.IsNotExist(err) {
		fmt.Println("Requesting blockchain from central node")
		RequestBlocks()
		for {
			select {
			case <-doneWritingBlockChain:
				fmt.Println("Done writing blockchain")
				break
			default:
				fmt.Println("Waiting for blockchain to be written")
				conn, err := ln.Accept()

				if err != nil {
					log.Panic(err)
				}
				HandleConnection(conn, nil)

			}
		}
	}

	bc := blockchain.NewBlockchain(nodeID)
	defer bc.Close()
	go HandleClose(bc)

	// If not the central node
	if myAddress != KnownNodes[0] {
		SendVersion(KnownNodes[0], bc.GetBestHeight())
	}
	log.Println("Listening on port :", nodeID)

	for {
		log.Println("Waiting for connection ")
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go HandleConnection(conn, bc)
	}

}

func HandleClose(bc *blockchain.Blockchain) {
	d := death.NewDeath(syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	d.WaitForDeathWithFunc(func() {
		defer os.Exit(1)
		defer runtime.Goexit()
		bc.Close()
	})
}

func HandleConnection(conn net.Conn, bc *blockchain.Blockchain) {

	defer conn.Close()
	data, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}

	command := ByteToCmd(data[:commandLength])
	fmt.Printf("Receive %s command\n", command)

	switch command {
	case sendAddrCmd.Command:
		ReceiveAddress(data)
	case sendBlockCmd.Command:
		ReceiveBlock(data, bc)
	case sendVersionCmd.Command:
		ReceiveVersion(data, bc)
	case sendInventoryCmd.Command:
		ReceiveInventory(data)
	case getBlocksCmd.Command:
		ReceiveBlocks(data, bc)
	case getDataCmd.Command:
		ReceiveGetData(data, bc)
	case sendTxCmd.Command:
		ReceiveTransaction(data, bc)
	case getBlockChainCmd.Command:
		HandleSendBuildBlockchain(data)
	case receiveBlockChainCmd.Command:
		ReceiveBuildBlockChain(data)
	default:
		fmt.Printf("Unknown command %s\n", command)
	}
}
