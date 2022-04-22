package blockchain

import (
	"blockchaincore/utils"
	"bytes"
	"encoding/gob"
)

type TXOutput struct {
	// Store number of coins
	Value int
	// Store public key of the receiver
	PubKeyHash []byte
}

type TXOutputs struct {
	Outputs []TXOutput
}

// IsLockedWithKey checks if the output can be used by the owner of the pubkey
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// NewTXOutput create a new TXOutput and locking that to the given address
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))
	return txo
}

// Lock locks the address to the output transaction
func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := utils.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// Serialize serializes the TXOutputs into byte slice
func (o *TXOutputs) Serialize() []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(o)
	utils.HandleError(err)
	return buff.Bytes()
}

func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	utils.HandleError(err)
	return outputs
}
