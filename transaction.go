package main

import(
	"fmt"
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	// "encoding/hex"
	"log"
)

const subsidy = 10

type Transaction struct{
	ID []byte
	Vin []TXInput
	Vout []TXOutput
}

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

type TXOutput struct{
	Value int
    ScriptPubKey string
}

func (out *TXOutput)CanBeUnlockedWith(unlockingData string) bool{
	return out.ScriptPubKey == unlockingData
}

type TXInput struct{
	Txid []byte
	//存储的是之前交易的 ID
    Vout int
    //存储的是该输出在那笔交易中所有输出的索引
    ScriptSig string
}

func (in *TXInput)CanUnlockOutputWith(unlockingData string) bool{
	return in.ScriptSig == unlockingData
}

func NewCoinbaseTX(to string, data string) *Transaction{
	if data == ""{
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, data}
    txout := TXOutput{subsidy, to}
    tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
    tx.SetID()

    return &tx
}

// SetID sets ID of a transaction
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}
