package routes

import (
	. "blockchaincore/blockchain"
	"blockchaincore/types"
	"blockchaincore/utils"
	utils2 "blockchaincore/web/utils"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Response struct {
	Message string      `json:"message"`
	Status  int         `json:"status"`
	Data    interface{} `json:"data"`
}

type WalletRequest struct {
	utils.WalletData
	Password string `json:"password"`
}

type CreateWalletRequest struct {
	Password string `json:"password"`
	NodePort string `json:"node_port"`
}

func CreateWalletHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("CreateWalletHandler")
	var request CreateWalletRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	log.Println("Request: ", request)
	if err != nil {
		log.Println(err)
		w.Header().Set("Content-Type", "application/json")
		data, _ := json.Marshal(Response{Message: "Invalid request", Status: http.StatusBadRequest})
		w.Write(data)
		return
	}
	wallets, _ := NewWallets(request.NodePort)
	address, privateKey, publicKey := wallets.CreateWallet()
	wallets.SaveToFile(request.NodePort)
	w.Header().Set("Content-Type", "application/json")

	wallet := utils.WalletData{
		Address:    address,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}
	wallet = *utils.EncryptData(wallet, request.Password)

	data, _ := json.Marshal(wallet)
	_, err = w.Write(data)
	if err != nil {
		log.Println(err)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func AccessWalletHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if err := recover(); err != nil {
			log.Println("AccessWalletHandler: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Internal Server Error"))
		}
	}()

	var walletRequest WalletRequest
	err := json.NewDecoder(r.Body).Decode(&walletRequest)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		data, _ := json.Marshal(Response{Message: "Invalid request body", Status: http.StatusBadRequest})
		_, _ = w.Write(data)
		return
	}
	// Status ok
	w.WriteHeader(http.StatusOK)
	decodedWallet := utils.DecryptData(&walletRequest.WalletData, walletRequest.Password)
	res, _ := json.Marshal(decodedWallet)
	_, _ = w.Write(res)
}

type SendMoneyRequest struct {
	PrivateAddress string `json:"private_address"`
	ToAddress      string `json:"to_address"`
	Amount         int    `json:"amount"`
}

func SendMoneyFromWallet(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if err := recover(); err != nil {
			log.Println("SendMoneyFromWallet: ", err)
			data, _ := json.Marshal(Response{Message: "Server error ", Status: http.StatusBadRequest})
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write(data)
		}
	}()
	w.Header().Set("Content-Type", "application/json")

	var request SendMoneyRequest
	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		data, _ := json.Marshal(Response{Message: "Invalid request body", Status: http.StatusBadRequest})
		_, _ = w.Write(data)
		return
	}
	utils2.SendMoney(request.PrivateAddress, request.ToAddress, request.Amount)
	// Ok status
	w.WriteHeader(http.StatusOK)
}

type GetBalanceRequest struct {
	Address string `json:"address"`
}

func GetBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	defer func() {
		if err := recover(); err != nil {
			log.Println("GetBalance: ", err)
			data, _ := json.Marshal(Response{Message: "Server error ", Status: http.StatusBadRequest})
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write(data)
		}
	}()
	log.Println("Go getbalance")
	nodeID := os.Getenv("NODE_ID")
	w.Header().Set("Content-Type", "application/json")
	var addr GetBalanceRequest
	err := json.NewDecoder(r.Body).Decode(&addr)
	if err != nil {
		data, _ := json.Marshal(Response{Message: "Invalid request body", Status: http.StatusBadRequest})
		_, _ = w.Write(data)
		return
	}
	bc := NewBlockchain(nodeID)
	if !ValidateAddress(addr.Address) {
		log.Panic("ERROR: Address is not valid")
	}

	UTXOSet := UTXOSet{bc}
	defer bc.Db.Close()
	balance := 0
	pubKeyHash := utils.Base58Decode([]byte(addr.Address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)
	for _, out := range UTXOs {
		balance += out.Value
	}
	// Status ok
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(Response{Message: "Balance is", Status: http.StatusOK, Data: balance})
	if err != nil {
		log.Println("Error when get balance: ", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		res, _ := json.Marshal(Response{Message: "Server error", Status: http.StatusInternalServerError})
		_, _ = w.Write(res)
		return
	}
}

type BlockHeightResponse struct {
	Height    int             `json:"height"`
	BlockInfo types.BlockInfo `json:"data"`
}

// TODO: Test

func GetBlockByHeight(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("GetBlockByHeight: ", err)
			data, _ := json.Marshal(Response{Message: "Server error ", Status: http.StatusBadRequest})
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write(data)

		}
	}()
	height := mux.Vars(r)["height"]
	w.Header().Set("Content-Type", "application/json")

	bc := NewBlockchain(os.Getenv("NODE_ID"))
	defer bc.Close()
	heightInt, err := strconv.Atoi(height)
	if err != nil {
		log.Println("GetBlockByHeight: ", err)
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(Response{Message: "Invalid height", Status: http.StatusBadRequest})
		if err != nil {
			log.Println("Error when get block by height: ", err)
			return
		}
	}
	block, err := bc.GetBlockByHeight(heightInt)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(Response{Message: "Block not found", Status: http.StatusNotFound})
		return
	}
	_ = json.NewEncoder(w).Encode(block)
}

// TODO: Test

func GetTransactionByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	w.Header().Set("Content-Type", "application/json")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(Response{Message: "Invalid transaction id", Status: http.StatusBadRequest})
		return
	}
	bc := NewBlockchain(os.Getenv("NODE_ID"))
	defer bc.Close()
	tx, err := bc.GetTransaction(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(Response{Message: "Transaction not found", Status: http.StatusNotFound})
		return
	}
	if tx == nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(Response{Message: "Transaction not found", Status: http.StatusNotFound})
		return
	}
	_ = json.NewEncoder(w).Encode(tx)
}
