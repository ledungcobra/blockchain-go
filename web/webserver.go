package web

import (
	"blockchaincore/blockchain"
	"blockchaincore/utils"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"
)

const pathStatic = "./web/static/"

func StartWebServer(port string) {
	r := mux.NewRouter()
	r.HandleFunc("/", IndexHandler).Methods("GET")
	r.HandleFunc("/create-blockchain", CreateBlockChainHandler).Methods("POST")
	r.HandleFunc("/create-wallet", CreateWalletHandler).Methods("POST")
	r.HandleFunc("/get-balance", GetBalanceHandler).Methods("POST")

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir(pathStatic+"css"))))
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir(pathStatic+"js"))))
	log.Println("Starting web server on port " + port)
	if err := srv.ListenAndServe(); err != nil {
		log.Println("Server start fail")
		return
	}
}

type GetBalanceResponse struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func GetBalanceHandler(writer http.ResponseWriter, request *http.Request) {
	address := request.FormValue("address")
	nodePort := request.FormValue("node-port")
	log.Println("Get balance for address " + address + " on node port " + nodePort)
	writer.Header().Set("Content-Type", "application/json")
	if !blockchain.ValidateAddress(address) {
		writer.WriteHeader(http.StatusBadRequest)
		resp, _ := json.Marshal(GetBalanceResponse{
			Address: address,
			Balance: "0",
			Message: "Invalid address",
			Success: false,
		})
		writer.Write(resp)

		return
	}
	bc := blockchain.NewBlockchain(nodePort)
	utxoSet := blockchain.UTXOSet{bc}
	defer bc.Db.Close()
	balance := 0
	pubKeyHash := utils.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := utxoSet.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}
	writer.WriteHeader(http.StatusOK)
	resp, err := json.Marshal(GetBalanceResponse{
		Address: address,
		Balance: strconv.Itoa(balance),
		Message: "Get balance success",
		Success: true,
	})
	if err != nil {
		log.Println("Get balance fail")
		return
	}
	writer.Write(resp)

}

type CreateWalletResponse struct {
	Address    string `json:"address"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

func CreateWalletHandler(w http.ResponseWriter, r *http.Request) {
	nodePort := r.FormValue("node-port-wallet")
	wallets, _ := blockchain.NewWallets(nodePort)
	address, privateKey, publicKey := wallets.CreateWallet()
	wallets.SaveToFile(nodePort)
	w.Header().Set("Content-Disposition", "attachment; filename=wallet.json")
	w.Header().Set("Content-Type", "application/octet-stream")
	data, _ := json.Marshal(CreateWalletResponse{Address: address, PrivateKey: privateKey, PublicKey: publicKey})
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	_, err := w.Write(data)
	if err != nil {
		return
	}
}

type CreateBlockResponse struct {
	Message string `json:"message"`
}

func CreateBlockChainHandler(writer http.ResponseWriter, request *http.Request) {

	defer func() {
		if r := recover(); r != nil {
			writer.WriteHeader(http.StatusOK)
			log.Println(r)
			writer.Write([]byte("Server error"))
		}
	}()

	address := request.FormValue("address")
	nodePort := request.FormValue("node-port")
	if address == "" || nodePort == "" {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	bc := blockchain.CreateBlockchain(address, nodePort)
	defer bc.Db.Close()

	UTXOSet := blockchain.UTXOSet{bc}
	UTXOSet.Reindex()

	fmt.Println("Done!")
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	data, _ := json.Marshal(CreateBlockResponse{Message: "Blockchain created"})
	writer.Write(data)
}

type Data struct {
	Name   string
	Titles []string
}

func IndexHandler(w http.ResponseWriter, request *http.Request) {
	w.WriteHeader(http.StatusOK)

	err := template.Must(template.ParseFiles(pathStatic+"templates/index.html")).Execute(w, nil)
	if err != nil {
		log.Fatal("Load template fail")
		return
	}
	if err != nil {
		return
	}
}
