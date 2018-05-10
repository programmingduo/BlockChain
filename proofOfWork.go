package main

import(
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
	// "strings"
	// "encoding/hex"
	// "encoding/binary"
)

const targetBits = 24

var (
	maxNonce = math.MaxInt64
)

// ProofOfWork represents a proof-of-work
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork builds and returns a ProofOfWork
func NewProofOfWork(b *Block) *ProofOfWork{
	target := big.NewInt(1)
	target.Lsh(target, uint(256 - targetBits))
	
	return &ProofOfWork{b, target}
}

//准备数据
func (pow *ProofOfWork) PrepareData(nonce int) []byte{
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			IntToHex(pow.block.TimeStamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
		)

	return data
}

// Run performs a proof-of-work
func (pow *ProofOfWork) Run() (int, []byte) {
	var hash [32]byte
	var hashInt big.Int
	nonce := 0

	// fmt.Println("Mining the block containing \" %s \"", pow.block.Data)
	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)
	for nonce < maxNonce{
		data := pow.PrepareData(nonce)
		hash = sha256.Sum256(data)
		// fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Printf("\r%x\n\n", hash)

	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool{
	var hashInt big.Int
	data := pow.PrepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}