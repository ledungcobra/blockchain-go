package routes

import (
	"blockchaincore/blockchain"
	. "blockchaincore/types"
	"net/http"
	"os"
	"strconv"
)

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
	for {
		block := it.Next()
		for _, tx := range block.Transactions {
			txResponse.Transactions = append(txResponse.Transactions, TransactionInfo{
				FromAddress:     tx.FromAddress,
				ToAddress:       tx.ToAddress,
				Timestamp:       tx.Timestamp,
				BlockHeight:     block.Height,
				TransactionHash: string(tx.ID),
				Amount:          tx.Amount,
				TransactionFee:  tx.TransactionFee,
			})
		}
		count--
		if len(block.PrevBlockHash) == 0 || len(txResponse.Transactions) >= count {
			break
		}
	}
}
