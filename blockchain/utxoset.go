package blockchain

import (
	"blockchaincore/utils"
	"encoding/hex"
	"github.com/boltdb/bolt"
	"log"
)

type UTXOSet struct {
	Blockchain *Blockchain
}

const utxoBucket = "utxo"

// Reindex rebuilds the UTXO set
func (u UTXOSet) Reindex() {
	db := u.Blockchain.Db
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bolt.ErrBucketNotFound {
			log.Panic(err)
		}

		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			log.Panic(err)
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	UTXO := u.Blockchain.FindUTXO()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(key, outs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
}

func (u *UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.Db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txId := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txId] = append(unspentOutputs[txId], outIdx)
				}
			}
		}
		return nil
	})
	utils.HandleError(err)
	return accumulated, unspentOutputs
}

func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := u.Blockchain.Db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}

		return nil
	})
	utils.HandleError(err)
	return UTXOs
}

// TODO: NOT WORK
// Update When new block is mined UTXO set is updated
// Update by removing spent outputs and adding unspent outputs from newly mined transactions
func (u *UTXOSet) Update(block *Block) {
	db := u.Blockchain.Db
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.Transactions {
			if !tx.IsCoinbase() {
				for _, vin := range tx.Vin {
					updatedOuts := TXOutputs{}
					outsBytes := b.Get(vin.Txid)
					log.Println("outsBytes: ", outsBytes)
					outs := DeserializeOutputs(outsBytes)

					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						err := b.Delete(vin.Txid)
						utils.HandleError(err)
					} else {
						err := b.Put(vin.Txid, updatedOuts.Serialize())
						utils.HandleError(err)
					}
				}
			}

			newOutputs := TXOutputs{}
			for _, out := range tx.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			err := b.Put(tx.ID, newOutputs.Serialize())
			utils.HandleError(err)
		}
		return nil
	})
	utils.HandleError(err)

}

func (u UTXOSet) CountTransactions() int {
	db := u.Blockchain.Db
	counter := 0
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}
		return nil
	})
	utils.HandleError(err)
	return counter
}
