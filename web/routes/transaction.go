package routes

import (
	"blockchaincore/blockchain"
	"net/http"
	"os"
	"strconv"
)

type TransactionInfo struct {
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	Amount          int    `json:"amount"`
	Timestamp       int64  `json:"timestamp"`
	BlockHeight     int    `json:"block_height"`
	TransactionHash string `json:"transaction_hash"`
	TransactionFee  int    `json:"transaction_fee"`
}

type TransactionResponse struct {
	Transactions []TransactionInfo `json:"transactions"`
	Count        int               `json:"count"`
}

func GetTransaction(w http.ResponseWriter, request *http.Request) {
	countPr := request.URL.Query().Get("count")
	if countPr == "" {
		countPr = "2"
	}
	count, _ := strconv.Atoi(countPr)

	bc := blockchain.NewBlockchain(os.Getenv("NODE_ID"))
	defer bc.Close()
	txResponse := TransactionResponse{}
	it := bc.Iterator()
	//for {
	//	block := it.Next()
	//	for _, tx := range block.Transactions {
	//		txResponse.Transactions = append(txResponse.Transactions, TransactionInfo{
	//			FromAddress:     tx.,
	//			ToAddress:       tx.ToAddress,
	//			Amount:          tx.Amount,
	//			Timestamp:       tx.Timestamp,
	//			BlockHeight:     block.Height,
	//			TransactionHash: tx.TransactionHash,
	//			TransactionFee:  tx.TransactionFee,
	//		})
	//	}
	//	if len(txResponse.Transactions) >= count {
	//		break
	//	}
	//}
}
