package p2pserver

import (
	"blockchaincore/blockchain"
	"bytes"
	"encoding/gob"
	"io"
	"log"
	"net"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12
const CentralNode = "localhost:3000"

var myAddress string
var mineAddr string
var KnownNodes = map[string]bool{CentralNode: true}
var blocksInTransit = [][]byte{}
var memPool = make(map[string]blockchain.Transaction)

const mineTxCount = 1
const kindBlock = "block"
const kindTx = "tx"

func ExtractCmd(request []byte) []byte {
	return request[:commandLength]
}

func GobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func GobDecode[DataType any](data []byte, to *DataType) error {
	buff := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buff)
	return dec.Decode(to)
}

func SendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		// Node offline remove that node from the node list
		log.Printf("%s is not available, remaining nodes: %d\n", addr, len(KnownNodes))
		UpdateKnownNodes(addr)
		return
	}
	defer conn.Close()
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func UpdateKnownNodes(addr string) {
	delete(KnownNodes, addr)
}
