package routes

import (
	. "blockchaincore/blockchain"
	. "blockchaincore/types"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
)

type GetBlockResponse struct {
	Blocks []BlockInfo `json:"blocks"`
	Count  int         `json:"count"`
}

func GetBlock(w http.ResponseWriter, request *http.Request) {

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}
	}()
	blockChain := NewBlockchain(os.Getenv("NODE_ID"))
	defer blockChain.Close()

	countPr := request.URL.Query().Get("count")
	if countPr == "" {
		countPr = ""
	}
	count, _ := strconv.Atoi(countPr)

	it := blockChain.Iterator()
	var getBlockResp GetBlockResponse

	for {
		block := it.Next()
		var coinbaseTx Transaction
		for _, tx := range block.Transactions {
			if tx.IsCoinbase() {
				coinbaseTx = *tx
				break
			}
		}
		hasMiner := len(coinbaseTx.Vout) > 0
		var blockInfo BlockInfo
		if hasMiner {
			blockInfo = BlockInfo{
				BlockHash:     string(block.Hash),
				BlockHeight:   block.Height,
				Timestamp:     block.Timestamp,
				TxCount:       len(block.Transactions),
				MineByAddress: hex.EncodeToString(coinbaseTx.Vout[0].PubKeyHash),
				BlockReward:   coinbaseTx.Vout[0].Value,
			}
		} else {
			blockInfo = BlockInfo{
				BlockHash:     string(block.Hash),
				BlockHeight:   block.Height,
				Timestamp:     block.Timestamp,
				TxCount:       len(block.Transactions),
				MineByAddress: "",
				BlockReward:   block.Transactions[0].Vout[0].Value,
			}
		}
		getBlockResp.Blocks = append(getBlockResp.Blocks, blockInfo)
		count--
		if block == nil || len(block.PrevBlockHash) == 0 || count == 0 {
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	getBlockResp.Count = len(getBlockResp.Blocks)
	err := json.NewEncoder(w).Encode(getBlockResp)

	if err != nil {
		log.Println(err)
		return
	}
}
