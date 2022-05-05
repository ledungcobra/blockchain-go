package routes

import (
	"blockchaincore/blockchain"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
)

type TypeResult struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func Search(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			data, _ := json.Marshal(Response{Status: http.StatusInternalServerError, Message: err.(error).Error()})
			_, err := w.Write(data)
			if err != nil {
				return
			}
		}
	}()
	search := r.URL.Query().Get("search")
	if search == "" {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(Response{Status: http.StatusBadRequest, Message: "Search is empty"})
		_, _ = w.Write(data)
	}

	bc := blockchain.NewBlockchain(os.Getenv("NODE_ID"))
	defer bc.Close()
	heightInt, e := strconv.Atoi(search)
	if e != nil {
		// Is hashing of transaction
		tx, err := bc.GetTransaction(search)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			data, _ := json.Marshal(Response{Status: http.StatusNotFound, Message: "Transaction not found"})
			_, _ = w.Write(data)
		} else {
			data, _ := json.Marshal(TypeResult{Type: "tx", Data: tx})
			_, _ = w.Write(data)
		}
	} else {
		block, e := bc.GetBlockByHeight(heightInt)
		if e != nil {
			// Not found
			w.WriteHeader(http.StatusNotFound)
			data, _ := json.Marshal(Response{Status: http.StatusNotFound, Message: "Block not found"})
			_, _ = w.Write(data)
		} else {
			data, _ := json.Marshal(TypeResult{Type: "block", Data: block})
			_, _ = w.Write(data)
		}

	}
}
