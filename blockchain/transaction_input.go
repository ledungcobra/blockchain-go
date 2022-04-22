package blockchain

import "bytes"

type TXInput struct {
	// TXid is a utils of the transaction
	Txid []byte
	// Vout is the index of the output in the transaction
	Vout int
	// ScriptSig is the signature of the input
	Signature []byte
	PubKey    []byte
}

func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}
