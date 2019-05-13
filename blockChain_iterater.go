package main

import(
	"github.com/boltdb/bolt"
	"log"
	// "fmt"
)

type BlockchainIterator struct {
	currentHash []byte
	db *bolt.DB
}

func (i *BlockchainIterator) Next() *Block{
	var block *Block
	err := i.db.View(func(tx *bolt.Tx) error{
			b := tx.Bucket([]byte(blocksBucket))
			encoder := b.Get(i.currentHash)
			block = DeserializeBlock(encoder)

			return nil
		})
	if err != nil{
		log.Panic(err)
	}

	i.currentHash = block.PrevBlockHash
	return block
}