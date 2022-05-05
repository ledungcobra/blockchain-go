package p2pserver

import (
	"blockchaincore/blockchain"
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
)

///////////////////////////////////////////
//SEND ADDRESS AND HANDLE RECEIVE ADDRESS
///////////////////////////////////////////
func GetKnownNodes() []string {
	var nodes []string
	for node := range KnownNodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// SendGetAddress SendGetAddr returns true if already have nodes in KnownNodes
func SendGetAddress() bool {

	if len(KnownNodes) > 1 {
		return true
	}

	// Fetching all the nodes from the file
	log.Println("Sending GetAddress to all the nodes from central node")
	var r SendGetAddr
	r.AddrFrom = myAddress
	payload := GobEncode(r)
	request := append(getAddressesSerial, payload...)
	SendData(CentralNode, request)
	return false
}

func SendAddr(address string) {
	nodes := Addr{GetKnownNodes()}
	nodes.AddrList = append(nodes.AddrList, myAddress)
	payload := GobEncode(nodes)
	request := append(sendAddrCmdSerial, payload...)
	SendData(address, request)
}

func ReceiveAddress(data []byte) {
	var buff bytes.Buffer
	var addr Addr
	buff.Write(data[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&addr)
	if err != nil {
		log.Panic(err)
	}
	for _, add := range addr.AddrList {
		KnownNodes[add] = true
	}
	fmt.Printf("There are %d known nodes now!\n", len(KnownNodes))
}

///////////////////////////////////////////
//SEND BLOCK AND HANDLE RECEIVE BLOCK
///////////////////////////////////////////

func SendBlock(addr string, b *blockchain.Block) {
	d := Block{myAddress, b.Serialize()}
	payload := GobEncode(d)
	request := append(sendBlockCmdSerial, payload...)
	SendData(addr, request)
}

func ReceiveBlock(data []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload Block
	buff.Write(data[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	blockData := payload.Block
	block := blockchain.DeserializeBlock(blockData)
	fmt.Println("Received a new block!")
	bc.AddBlock(block)
	fmt.Printf("Added block %x\n", block.Hash)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		SendGetData(payload.AddrFrom, kindBlock, blockHash)

		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := blockchain.UTXOSet{bc}
		UTXOSet.Reindex()
	}
}

///////////////////////////////////////////
//SEND BLOCK AND HANDLE RECEIVE BLOCK
///////////////////////////////////////////

func SendInventory(address, kind string, items [][]byte) {
	inv := Inventory{myAddress, kind, items}
	payload := GobEncode(inv)
	request := append(sendInventoryCmdSerial, payload...)
	SendData(address, request)
}

func ReceiveInventory(data []byte) {
	var buff bytes.Buffer
	var payload Inventory
	buff.Write(data[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("Received inventory with %d %s\n", len(payload.Items), payload.Type)
	if payload.Type == kindBlock {
		blocksInTransit = payload.Items
		blockHash := payload.Items[0]
		SendGetData(payload.AddrFrom, kindBlock, blockHash)
	}

	if payload.Type == kindTx {
		txID := payload.Items[0]
		if memPool[hex.EncodeToString(txID)].ID == nil {
			SendGetData(payload.AddrFrom, kindTx, txID)
		}
	}
}

///////////////////////////////////////////
//SEND BLOCKS AND HANDLE RECEIVE BLOCKS
///////////////////////////////////////////

func SendGetBlocks(address string) {
	payload := GobEncode(GetBlocks{myAddress})
	request := append(getBlocksCmdSerial, payload...)
	SendData(address, request)
}

func ReceiveBlocks(data []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload GetBlocks
	buff.Write(data[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	blocks := bc.GetBlockHashes()
	SendInventory(payload.AddrFrom, kindBlock, blocks)
}

///////////////////////////////////////////
//SEND DATA AND HANDLE RECEIVE DATA
///////////////////////////////////////////

func SendGetData(address string, kind string, id []byte) {
	if kind != kindBlock && kind != kindTx {
		log.Panic("SendGetData: unknown kind")
	}
	payload := GobEncode(GetData{myAddress, kind, id})
	request := append(getDataCmdSerial, payload...)
	SendData(address, request)
}

func ReceiveGetData(data []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload GetData
	buff.Write(data[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	id, rType, addrFrom := payload.ID, payload.Type, payload.AddrFrom
	if rType == kindBlock {
		block, err := bc.GetBlock(id)
		if err != nil {
			return
		}
		SendBlock(addrFrom, &block)
	} else if rType == kindTx {
		txID := hex.EncodeToString(id)
		tx := memPool[txID]
		SendTx(addrFrom, &tx)
	}
}

///////////////////////////////////////////
//SEND TRANSACTION AND HANDLE RECEIVE TRANSACTION
///////////////////////////////////////////

func SendTx(addr string, tx *blockchain.Transaction) {
	data := Tx{myAddress, tx.Serialize()}
	payload := GobEncode(data)
	request := append(sendTxCmdSerial, payload...)
	SendData(addr, request)
}

func ReceiveTransaction(data []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload Tx

	buff.Write(data[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Data
	tx := blockchain.DeserializeTransaction(txData)
	memPool[hex.EncodeToString(tx.ID)] = tx

	log.Printf("My address is %s size of mempool: %d\n", myAddress, len(memPool))
	isCentralNode := myAddress == CentralNode

	if isCentralNode {
		// Broadcast transaction to all peers
		log.Printf("Number of known nodes: %d\n", len(KnownNodes))
		for node := range KnownNodes {
			if node != myAddress && node != payload.AddrFrom {
				SendInventory(node, kindTx, [][]byte{tx.ID})
			}
		}
	} else {
		// Has miner address to receive reward
		if len(mineAddr) != 0 {
			if len(memPool) >= mineTxCount {
				log.Println("Mining a new block")
				MineTx(bc)
			}
		} else {
			log.Println("Mining is off")
		}
	}
}

func MineTx(bc *blockchain.Blockchain) {
	var validTxs []*blockchain.Transaction
	for id := range memPool {
		fmt.Printf("Mining txid = %x\n", memPool[id].ID)
		tx := memPool[id]

		// Transaction is valid
		if bc.VerifyTransaction(&tx) {
			log.Printf("Transaction id %s is valid\n", id)
			validTxs = append(validTxs, &tx)
		} else {
			log.Printf("Transaction id %s is invalid\n", id)
		}
	}

	if len(validTxs) == 0 {
		fmt.Println("All transactions are invalid")
		return
	}
	totalFee := 0

	for _, transaction := range validTxs {
		totalFee += transaction.TransactionFee
	}
	log.Println("Total fee:", totalFee)

	coinBaseTx := blockchain.NewCoinbaseTX(mineAddr, "", totalFee)
	validTxs = append(validTxs, coinBaseTx)
	newBlock := bc.MineBlock(validTxs)
	utxoSet := blockchain.UTXOSet{bc}
	utxoSet.Reindex()
	fmt.Println("New block mined")
	for _, tx := range validTxs {
		txId := hex.EncodeToString(tx.ID)
		delete(memPool, txId)
	}

	for node := range KnownNodes {
		if node != myAddress {
			SendInventory(node, kindBlock, [][]byte{newBlock.Hash})
			SendDeleteTxFromPool(node, validTxs)
		}
	}

	if len(memPool) > 0 {
		MineTx(bc)
	}
}

type DeleteTX struct {
	AddrFrom string
	ID       [][]byte
}

func SendDeleteTxFromPool(node string, txs []*blockchain.Transaction) {
	txIDs := make([][]byte, len(txs))
	for i, tx := range txs {
		txIDs[i] = tx.ID
	}
	r := DeleteTX{myAddress, txIDs}
	payload := GobEncode(r)
	request := append(deleteTxPoolCmdSerial, payload...)
	SendData(node, request)
}

///////////////////////////////////////////
//SEND VERSION AND HANDLE RECEIVE VERSION
///////////////////////////////////////////

func SendVersion(addr string, bc *blockchain.Blockchain) {
	bestHeight := bc.GetBestHeight()
	lastHash := bc.GetLastHash()
	payload := GobEncode(Version{nodeVersion, bestHeight, myAddress, lastHash})
	request := append(sendVersionCmdSerial, payload...)
	SendData(addr, request)
}

func ReceiveVersion(data []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload Version
	buff.Write(data[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
		return
	}
	myHeight := bc.GetBestHeight()
	otherHeight := payload.BestHeight

	log.Printf("My height is %d, other height is: %d", myHeight+1, otherHeight+1)

	if myHeight < otherHeight {
		SendGetBlocks(payload.AddrFrom)
	} else if myHeight > otherHeight {
		SendVersion(payload.AddrFrom, bc)
	} else {
		// Height is equal do nothing
	}

	KnownNodes[payload.AddrFrom] = true
}

func RequestBlocks() {
	centralNodeAddr := CentralNode
	buildBlockChain := BuildBlockChain{myAddress}
	payload := GobEncode(buildBlockChain)
	request := append(getBlockChainCmdSerial, payload...)
	SendData(centralNodeAddr, request)
}

func HandleSendBuildBlockchain(data []byte) {

	var buff bytes.Buffer
	var payload BuildBlockChain
	buff.Write(data[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	addrFrom := payload.AddrFrom
	nodeID := os.Getenv("NODE_ID")
	file := fmt.Sprintf(blockchain.DbFile, nodeID)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		log.Panic("File not found")
		return
	}
	blockChainData, err := ioutil.ReadFile(file)
	log.Println("Found file, and will be transfer to other node")
	if err != nil {
		log.Panic(err)
	}
	result := bytes.Join([][]byte{receiveBlockChainCmdSerial, blockChainData}, []byte{})
	SendData(addrFrom, result)

}

func ReceiveBuildBlockChain(data []byte, ln net.Listener) {
	defer ln.Close()
	blockChainData := data[commandLength:]
	log.Println("Build blockchain from other node")
	nodeID := os.Getenv("NODE_ID")
	err := ioutil.WriteFile(fmt.Sprintf(blockchain.DbFile, nodeID), blockChainData, 0644)
	if err != nil {
		log.Panic("Error writing blockchain to file")
		return
	}
	doneWritingBlockChain <- true
}
