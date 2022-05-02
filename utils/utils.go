package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"log"
)

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

func PubBytes(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(elliptic.P256(), pub.X, pub.Y)
}
