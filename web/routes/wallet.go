package routes

import (
	"blockchaincore/blockchain"
	"blockchaincore/cli"
	"blockchaincore/utils"
	"encoding/json"
	"log"
	"net/http"
)

type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

type WalletRequest struct {
	utils.WalletData
	Password string `json:"password"`
}

type CreateWalletRequest struct {
	Password string `json:"password"`
	NodePort string `json:"node_port"`
}

var Cli = cli.NewCLI()

func CreateWalletHandler(w http.ResponseWriter, r *http.Request) {

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
	wallets, _ := blockchain.NewWallets(request.NodePort)
	address, privateKey, publicKey := wallets.CreateWallet()
	wallets.SaveToFile(request.NodePort)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=wallet.json")
	wallet := utils.WalletData{
		Address:    address,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}
	encryptedWallet := utils.EncryptData(wallet, request.Password)

	//w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	//_, err = w.Write(data)
	ec := json.NewEncoder(w)
	ec.Encode(encryptedWallet)
	if err != nil {
		return
	}

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

	decodedWallet := utils.DecryptData(&walletRequest.WalletData, walletRequest.Password)
	res, _ := json.Marshal(decodedWallet)
	_, _ = w.Write(res)
}
