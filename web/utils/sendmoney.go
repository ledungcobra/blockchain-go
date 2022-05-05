package utils

import (
	"blockchaincore/blockchain"
	"blockchaincore/p2pserver"
	"log"
	"os"
)

func SendMoney(priKeyFrom, toAddress string, amount int) {
	nodeID := os.Getenv("NODE_ID")
	if !blockchain.ValidateAddress(toAddress) {
		log.Panic("ERROR: Recipient address is not valid")
	}
	bc := blockchain.NewBlockchain(nodeID)
	defer bc.Db.Close()
	wallets, err := blockchain.NewWallets(nodeID)
	wallet := wallets.FromPrivateKey(priKeyFrom)

	if err != nil || wallet == nil {
		// Handle sync wallet from other node
	}

	UTXOSet := blockchain.UTXOSet{bc}
	tx := blockchain.NewUTXOTransaction(wallet, toAddress, amount, &UTXOSet)
	p2pserver.SendTx(p2pserver.CentralNode, tx)
}
