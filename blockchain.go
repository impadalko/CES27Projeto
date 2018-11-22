package main

import (
	"fmt"
	"crypto/sha256"
	"encoding/hex"
	
	"encoding/binary"
	"bytes"

	"time"
)

type HashVal [32]byte // SHA-256 outputs 256 bits (32 bytes)

// We will be using SHA-256 for hashing blocks
func HashFunc(data interface{}) HashVal {
	buffer := bytes.Buffer{}
	binary.Write(&buffer, binary.LittleEndian, data)
	return sha256.Sum256(buffer.Bytes())
}

func (hashVal HashVal) ToString() string {
	return hex.EncodeToString(hashVal[:])
}

func UnixTimestamp() int64 {
	return time.Now().Unix()
}

type Block struct {
	Index int64
	PreviousHash HashVal
	Timestamp int64
	Data [512]byte
}

func (block Block) Hash() HashVal {
	return HashFunc(block)
}

type BlockChain struct {
	NextIndex int64
	LastHash HashVal
	Blocks []Block
}

func CreateBlockChain() BlockChain {
	genesis := Block{}
	genesis.Timestamp = UnixTimestamp()

	blockChain := BlockChain{}
	blockChain.LastHash = genesis.Hash()
	blockChain.Blocks = append(blockChain.Blocks, genesis)
	blockChain.NextIndex = 1
	return blockChain
}

func (blockChain *BlockChain) AddBlock(Data [512]byte) {
	block := Block{
		blockChain.NextIndex,
		blockChain.LastHash,
		UnixTimestamp(),
		Data,
	}
	blockChain.Blocks = append(blockChain.Blocks, block)
	blockChain.NextIndex++
	blockChain.LastHash = block.Hash()
}

func (blockChain BlockChain) VerifyConsistency() bool {
	lastHash := blockChain.Blocks[0].Hash()
	for _, block := range blockChain.Blocks[1:] {
		if block.PreviousHash != lastHash {
			return false
		}
		lastHash = block.Hash()
	}
	if lastHash != blockChain.LastHash {
		return false
	}
	return true
}

func main() {
	bc := CreateBlockChain()

	for i := 0; i < 5; i++ {
		b := [512]byte{}
		b[0] = byte(i)
		bc.AddBlock(b)
	}

	fmt.Println(bc.VerifyConsistency())

	bc.Blocks[3].Data[1] = 1
	
	fmt.Println(bc.VerifyConsistency())
}