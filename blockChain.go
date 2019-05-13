package main

import(
	"github.com/boltdb/bolt"
	"log"
	"fmt"
	"os"
	"encoding/hex"
)

type Blockchain struct{
	tip []byte
	db *bolt.DB
}

const blocksBucket = "blocks"  
const dbFile = "blockchain.db"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

// func (bc *Blockchain) AddBlock(block *Block) {
// 	var lastHash []byte 
// 	err := bc.db.View(func (tx *bolt.Tx) error{
// 			b := tx.Bucket([]byte(blocksBucket))
// 			lastHash = b.Get([]byte("1"))
// 			 return nil
// 		})
// 	if err != nil{
// 		log.Panic(err)
// 	}

// 	newBlock := NewBlock(data, lastHash)

// 	err = bc.db.Update(func (tx *bolt.Tx) error{
// 			b := tx.Bucket([]byte(blocksBucket))
// 			err := b.Put(newBlock.Hash, newBlock.Serialize())
// 			if err != nil{
// 				log.Panic(err)
// 			}
// 			err = b.Put([]byte("1"), newBlock.Hash)
// 			if err != nil{
// 				log.Panic(err)
// 			}
// 			bc.tip = newBlock.Hash

// 			return nil
// 		})
// 	if err != nil{
// 		log.Panic(err)
// 	}
// }

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

func GetBlockchain() *Blockchain{
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil{
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error{
			b := tx.Bucket([]byte(blocksBucket))
			tip = b.Get([]byte("1"))
			return nil
		})
	if err != nil{
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

func CreateBlockchain(address string) *Blockchain{
	if dbExists(){
		fmt.Println("Blockchain already exit")
		os.Exit(1)
	}

	var tip []byte
	db,err := bolt.Open(dbFile, 0600, nil)
	if err != nil{
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error{
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("1"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash

		return nil
		})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	var acc = 0

	txs := bc.FindUnspentTransactions(from)
	for _, tx := range txs {
		for txid, out := range tx.Vout {
			if out.CanBeUnlockedWith(from) && acc < amount{
				acc += out.Value
				input := TXInput{tx.ID, txid, from}
				inputs = append(inputs, input)
			}
		}
	}

	if acc < amount{
		log.Panic("ERROR: Not enough funds")
	}

	output := TXOutput{amount, to}
	outputs = append(outputs, output)

	if acc > amount{
		output.Value = acc - amount
		output.ScriptPubKey = from
		outputs = append(outputs, output)
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()
	return &tx
}

func (bc *Blockchain) MineBlock(txs []*Transaction) {
	var prvHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error{
		b := tx.Bucket([]byte(blocksBucket))
		prvHash = b.Get([]byte("1"))

		return nil
		})
	if err != nil{
		log.Panic(err)
	}
	
	block := NewBlock(txs, prvHash)

	err = bc.db.Update(func(tx *bolt.Tx) error{
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(block.Hash, block.Serialize())
		if err != nil{
			log.Panic(err)
		}

		err = b.Put([]byte("1"), block.Hash)
		if err != nil{
			log.Panic(err)
		}

		bc.tip = block.Hash

		return nil
		})
	if err != nil{
		log.Panic(err)
	}
}

func (bc *Blockchain) FindUnspentTransactions(address string)[]Transaction {
	bci := bc.Iterator()
	var unspentTXs []Transaction
  	spentTXOs := make(map[string][]int)

	for{
		//遍历区块链中的所有block
		block := bci.Next()
		//每一个block有可能存储着若干transactions
		for _, tx := range block.Transactions{
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			//每一个transaction又有很多vout
			for outIdx, out := range tx.Vout{
				//判断这一vout是否能够被address解密
				if out.CanBeUnlockedWith(address){
					//如果可以又要判断是否被花费
					if(spentTXOs[txID] != nil){
						for sTX := range spentTXOs[txID]{
							if sTX == outIdx{
								continue Outputs
							}
						}
					}

					unspentTXs = append(unspentTXs, *tx)
				}
			}

			//判断是否被花费的过程中又需要存储所有已经花费的transaction
			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin{
					if in.CanUnlockOutputWith(address) {
            			inTxID := hex.EncodeToString(in.Txid)
            			spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
        		  	}
				}
			}

		}
		if len(block.PrevBlockHash) == 0 {
     		break
   		}
	}

	return unspentTXs
}

func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	uTX := bc.FindUnspentTransactions(address)
	var UTXO []TXOutput

	for _, tx := range uTX{
		for _, out := range tx.Vout{
			if out.CanBeUnlockedWith(address){
				UTXO = append(UTXO, out)
			}
		}
	}
	return UTXO
}

func (bc *Blockchain) Iterator() *BlockchainIterator{
	bci := &BlockchainIterator{bc.tip, bc.db}
	return bci
}
