package utils

import (
	"encoding/json"
	jose "github.com/dvsekhvalnov/jose2go"
	"log"
)

type WalletData struct {
	Address    string `json:"address"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

func HandleError(e error) {
	if e != nil {
		log.Panicln(e)
	}
}

func EncryptData(wallet WalletData, password string) *WalletData {
	newWallet := WalletData{}
	dat, _ := json.Marshal(wallet)
	result, r := jose.Sign(string(dat), jose.HS256, []byte(password))
	if r != nil {
		log.Println(r)
		return nil
	}
	newWallet.PrivateKey = result[0:15]
	newWallet.PublicKey = result[15:30]
	newWallet.Address = result[30:]
	return &newWallet
}

func DecryptData(wallet *WalletData, password string) *WalletData {
	newWallet := WalletData{}
	payload, _, _ := jose.Decode(wallet.PrivateKey+wallet.PublicKey+wallet.Address, []byte(password))
	err := json.Unmarshal([]byte(payload), &newWallet)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &newWallet
}
