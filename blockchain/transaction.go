package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"
)

type Transaction struct {
	ID             []byte
	Vin            []TXInput
	Vout           []TXOutput
	Timestamp      int64
	FromAddress    string
	ToAddress      string
	Amount         int
	TransactionFee int
}

const rewardInitValue = 100
const randomFactor = 20

func Now() int64 {
	return time.Now().Unix()
}

// NewCoinbaseTX  creates a new coinbase transaction
// If fee = -1 mean that is genesis transaction
func NewCoinbaseTX(to, data string, fee int) *Transaction {
	if data == "" {
		randData := make([]byte, randomFactor)
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}
		data = fmt.Sprintf("%x", randData)

	}

	txin := TXInput{[]byte{}, -1, nil, []byte(data)}

	// Change reward amount to fee
	rewardAmount := rewardInitValue
	if fee != -1 {
		rewardAmount = int(float32(fee) * 1.5)
	}

	txout := NewTXOutput(rewardAmount, to)

	log.Println("Reward amount: ", rewardAmount)
	// it is transaction normal, not genesis block
	tx := Transaction{
		ID:             nil,
		Vin:            []TXInput{txin},
		Vout:           []TXOutput{*txout},
		Timestamp:      Now(),
		FromAddress:    "Base Reward",
		ToAddress:      to,
		TransactionFee: 0,
		Amount:         rewardAmount,
	}
	tx.ID = tx.Hash()
	return &tx
}

// NewUTXOTransaction new transaction for sending money from, to address with amount of money
// include process of sign transaction
// and returns a brand new transaction
func NewUTXOTransaction(wallet *Wallet, to string, amount int, UTXOSet *UTXOSet) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	fee := CalcTxFee(amount)
	pubKeyHash := HashPubKey(wallet.PublicKey)
	from := fmt.Sprintf("%s", wallet.GetAddress())
	totalAmount := amount + fee
	acc, validOutputs := UTXOSet.FindSpendableOutputs(pubKeyHash, totalAmount)

	if acc < totalAmount {
		log.Panic("Insufficient funds")
		return nil
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)

		if err != nil {
			fmt.Println(err)
		}

		for _, out := range outs {
			input := TXInput{txID, out, nil, wallet.PublicKey}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, *NewTXOutput(amount, to))
	remainder := acc - totalAmount

	if remainder > 0 {
		outputs = append(outputs, *NewTXOutput(remainder, from))
	}

	tx := Transaction{
		ID:             nil,
		Vin:            inputs,
		Vout:           outputs,
		Timestamp:      Now(),
		FromAddress:    from,
		ToAddress:      to,
		Amount:         amount,
		TransactionFee: fee,
	}

	tx.ID = tx.Hash()
	UTXOSet.Blockchain.SignTransaction(&tx, wallet.PrivateKey)

	return &tx
}

func CalcTxFee(amount int) int {
	txFee := 0
	if amount < 50 {
		return 1
	} else if amount > 50 {
		txFee = amount * 10 / 100
	} else if amount > 200 {
		txFee = amount * 20 / 100
	} else if amount > 500 {
		txFee = amount * 30 / 100
	}
	return txFee
}

// IsCoinbase check if transaction is coinbase
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// Hash hashes entire transaction
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())
	return hash[:]
}

// String returns a human-readable representation of a transaction
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))
	lines = append(lines, fmt.Sprintf("	Timestamp: %d", tx.Timestamp))
	lines = append(lines, fmt.Sprintf("	From:      %s", tx.FromAddress))
	lines = append(lines, fmt.Sprintf("	To:        %s", tx.ToAddress))

	for i, input := range tx.Vin {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}

// Serialize serialize transaction into byte slice
func (tx *Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing transaction
func (tx *Transaction) TrimmedCopy() Transaction {

	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}

	txCopy := Transaction{ID: tx.ID, Vin: inputs, Vout: outputs, Timestamp: tx.Timestamp, FromAddress: tx.FromAddress, ToAddress: tx.ToAddress}

	return txCopy
}

// Sign signs transaction with private key and previous transactions
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil

		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		dataToSign := fmt.Sprintf("%x\n", txCopy)

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, []byte(dataToSign))
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inID].Signature = signature
		txCopy.Vin[inID].PubKey = nil

	}
}

// Verify verifies transaction
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash

		r, s, x, y := big.Int{}, big.Int{}, big.Int{}, big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])

		dataToVerify := fmt.Sprintf("%x\n", txCopy)

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) == false {
			return false
		}
		txCopy.Vin[inID].PubKey = nil
	}

	return true
}

// DeserializeTransaction deserializes a transaction
func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	if err != nil {
		log.Panic(err)
	}

	return transaction
}
