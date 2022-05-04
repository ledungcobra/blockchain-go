package routes

import (
	"blockchaincore/blockchain"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
)

type BlockInfo struct {
	BlockHash     string `json:"block_hash"`
	BlockHeight   int    `json:"block_height"`
	Timestamp     int64  `json:"timestamp"`
	TxCount       int    `json:"tx_count"`
	MineByAddress string `json:"mine_by_address"`
	BlockReward   int    `json:"block_reward"`
}

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
	//Cli.SynBlockChain()
	blockChain := blockchain.NewBlockchain(os.Getenv("NODE_ID"))
	defer blockChain.Close()

	countPr := request.URL.Query().Get("count")
	if countPr == "" {
		countPr = "2"
	}
	count, _ := strconv.Atoi(countPr)

	it := blockChain.Iterator()
	var getBlockResp GetBlockResponse

	for {
		block := it.Next()
		var coinbaseTx blockchain.Transaction
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
				MineByAddress: string(coinbaseTx.Vout[0].PubKeyHash),
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
