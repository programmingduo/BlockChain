package main

import(
	// "bytes"
	"time"
	// "crypto/sha256"
	// "strconv"
)

type Block struct{
	TimeStamp int64
	Data []byte
	PrevBlockHash []byte
	Hash []byte
	Nonce int
}

// //Hash = SHA256(PrevBlockHash + Timestamp + Data)
// func (b *Block) SetHash(){
// 	timestamp := []byte(strconv.FormatInt(b.TimeStamp, 10))
// 	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
// 	hash := sha256.Sum256(headers)

// 	b.Hash = hash[:]
// }

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

