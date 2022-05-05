package p2pserver

import "fmt"

const addr = "addr"
const block = "block"
const inv = "inv"
const tx = "tx"
const version = "version"
const getBlocks = "getblocks"
const getData = "getdata"
const getBlockchain = "getbc"
const receiveBlock = "recv_bc"

const getAddr = "getaddr"

const deleteTxPool = "del_tx_pool"

type Command struct {
	Command string
}

func NewCommand(command string) *Command {
	return &Command{Command: command}
}

func (c *Command) Bytes() []byte {
	var bytes [commandLength]byte
	for i, c := range c.Command {
		bytes[i] = byte(c)
	}
	return bytes[:]
}

func ByteToCmd(data []byte) string {
	// TODO: Refactor
	var cmd []byte
	for _, b := range data {
		if b != 0x0 {
			cmd = append(cmd, b)
		}
	}
	return fmt.Sprintf("%s", cmd)
}

var sendAddrCmd = NewCommand(addr)
var sendAddrCmdSerial = sendAddrCmd.Bytes()

var sendBlockCmd = NewCommand(block)
var sendBlockCmdSerial = sendBlockCmd.Bytes()

var sendInventoryCmd = NewCommand(inv)
var sendInventoryCmdSerial = sendInventoryCmd.Bytes()

var sendTxCmd = NewCommand(tx)
var sendTxCmdSerial = sendTxCmd.Bytes()

var sendVersionCmd = NewCommand(version)
var sendVersionCmdSerial = sendVersionCmd.Bytes()

var getBlocksCmd = NewCommand(getBlocks)
var getBlocksCmdSerial = getBlocksCmd.Bytes()

var getDataCmd = NewCommand(getData)
var getDataCmdSerial = getDataCmd.Bytes()

var getBlockChainCmd = NewCommand(getBlockchain)
var getBlockChainCmdSerial = getBlockChainCmd.Bytes()

var receiveBlockChainCmd = NewCommand(receiveBlock)
var receiveBlockChainCmdSerial = receiveBlockChainCmd.Bytes()

var getAddresses = NewCommand(getAddr)
var getAddressesSerial = getAddresses.Bytes()

var deleteTxPoolCmd = NewCommand(deleteTxPool)
var deleteTxPoolCmdSerial = deleteTxPoolCmd.Bytes()
