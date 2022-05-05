package routes

import (
	"blockchaincore/blockchain"
	. "blockchaincore/types"
	"encoding/hex"
	"encoding/json"
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
		countPr = "50"
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
				TransactionHash: hex.EncodeToString(tx.ID),
				Amount:          tx.Amount,
				TransactionFee:  tx.TransactionFee,
			})
		}
		count--
		if len(block.PrevBlockHash) == 0 || len(txResponse.Transactions) >= count {
			break
		}
	}
	txResponse.Count = len(txResponse.Transactions)
	data, _ := json.Marshal(txResponse)
	_, _ = w.Write(data)
}
