package p2pserver

import (
	"blockchaincore/blockchain"
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/vrecan/death"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"
	"syscall"
	"time"
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
		// Blocking the rest of the program until the blockchain is sync from the central node
		GetBlockFromCentralNode(ln)
	}

	bc := blockchain.NewBlockchain(nodeID)
	defer bc.Close()
	go HandleClose(bc)

	// If not the central node
	if myAddress != CentralNode {
		SendVersion(CentralNode, bc)
	}
	log.Println("Listening on port :", nodeID)

	for {
		log.Println("Waiting for connection ")
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go HandleConnection(conn, bc, ln)
	}

}

func GetBlockFromCentralNode(ln net.Listener) {
	if myAddress == "" {
		myAddress = fmt.Sprintf("localhost:%s", os.Getenv("NODE_ID"))
	}
	fmt.Println("Requesting blockchain from central node")
	RequestBlocks()
	connCh := make(chan net.Conn)

	go func(c chan net.Conn) {
		for {
			// Blocking
			conn, err := ln.Accept()
			if err != nil {
				log.Println(err)
				log.Println("Stop listening")
				break
			}
			c <- conn
			fmt.Println("Handle connection")
		}
	}(connCh)

	for {
		select {
		case <-doneWritingBlockChain:
			fmt.Println("Done writing blockchain")
			return
		case conn := <-connCh:
			fmt.Println("Got connection")
			go HandleConnection(conn, nil, ln)
		case <-time.After(time.Second * 3):
			fmt.Println("Timeout 3s")
		}
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

func HandleConnection(conn net.Conn, bc *blockchain.Blockchain, ln net.Listener) {

	defer conn.Close()
	data, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}

	command := ByteToCmd(data[:commandLength])
	log.Printf("Receive %s command\n", command)

	switch command {
	case sendVersionCmd.Command:
		ReceiveVersion(data, bc)
	case sendAddrCmd.Command:
		ReceiveAddress(data)
	case sendBlockCmd.Command:
		ReceiveBlock(data, bc)
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
		ReceiveBuildBlockChain(data, ln)
	case deleteTxPoolCmd.Command:
		ReceiveDeleteTxPool(data)
	default:
		fmt.Printf("Unknown command %s\n", command)
	}
}

func ReceiveDeleteTxPool(data []byte) {
	var payload = data[commandLength:]
	var txPoolDelete = DeleteTX{}
	var bytesBuffer bytes.Buffer
	bytesBuffer.Write(payload)
	decoder := gob.NewDecoder(&bytesBuffer)
	err := decoder.Decode(&txPoolDelete)
	if err != nil {
		log.Panic(err)
	}

	for i := range txPoolDelete.ID {
		delete(memPool, hex.EncodeToString(txPoolDelete.ID[i]))
	}
	log.Println("Delete txs pool")
}
