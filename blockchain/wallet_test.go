package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"testing"
)

func TestFromPrivateKey(t *testing.T) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	_ = fmt.Sprintf("%x", privateKey.D)
	pKey, _ := ToECDSAFromHex(fmt.Sprintf("%x", privateKey.D))
	fmt.Printf("pKey: %x\n", pKey.PublicKey)
	fmt.Printf("privateKey: %x\n", privateKey.PublicKey)

	p1 := PrivateKeyToBytes(privateKey)
	p2 := PrivateKeyToBytes(pKey)

	if bytes.Compare(p1, p2) != 0 {
		t.Error("privateKeyToBytes not equal")
	}
}
