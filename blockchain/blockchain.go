package blockchain

import (
	"fmt"
	"sync"
	"errors"
	"encoding/hex"
)

type BlockChain struct {
	NextIndex int64
	LastHash  HashVal
	Blocks    []Block
	Lock      sync.RWMutex
}

func New(timestamp int64, Data []byte) *BlockChain {
	return NewFromBlock(Block{
		0,
		HashVal{},
		timestamp,
		int32(len(Data)),
		Data,
	})
}

func NewFromBlock(block Block) *BlockChain {
	return &BlockChain{
		1,
		block.Hash(),
		[]Block{block},
		sync.RWMutex{},
	}
}

func (bc *BlockChain) AddBlockFromData(timestamp int64, Data []byte) error {
	return bc.AddBlock(Block{
		bc.NextIndex,
		bc.LastHash,
		timestamp,
		int32(len(Data)),
		Data,
	})
}

func (bc *BlockChain) AddBlock(block Block) error {
	bc.Lock.Lock()
	defer bc.Lock.Unlock()
	if block.PreviousHash != bc.LastHash {
		return errors.New("Previous hash of the block doesn't match")
	}
	if block.Index != bc.NextIndex {
		return errors.New("Index of the block doesn't match")
	}
	bc.Blocks = append(bc.Blocks, block)
	bc.NextIndex++
	bc.LastHash = block.Hash()
	return nil
}

func (bc *BlockChain) VerifyConsistency() bool {
	// TODO verify if indexes follow 0, 1, 2 ...
	// TODO verify if timestamps are non-decreasing
	// TODO verify if DataLen = len(Data)
	bc.Lock.Lock()
	defer bc.Lock.Unlock()
	lastHash := bc.Blocks[0].Hash()
	for _, block := range bc.Blocks[1:] {
		if block.PreviousHash != lastHash {
			return false
		}
		lastHash = block.Hash()
	}
	if lastHash != bc.LastHash {
		return false
	}
	return true
}

func (bc *BlockChain) PrintBlocks() {
	bc.Lock.RLock()
	fmt.Printf("%5s %-8s %-8s %-10s %s\n", "Index", "Hash", "PrevHash", "Timestamp", "Data")
	for _, block := range bc.Blocks {
		hexData := hex.EncodeToString(block.Data)
		if len(hexData) > 8 {
			hexData = hexData[:8]
		}
		fmt.Printf("%5d %8s %8s %10d %s\n", block.Index, block.Hash().String()[:8],
			block.PreviousHash.String()[:8], block.Timestamp, hexData)
	}
	fmt.Println()
	bc.Lock.RUnlock()
}

func (bc *BlockChain) GetBlock(index int) Block {
	bc.Lock.RLock()
	block := bc.Blocks[index]
	bc.Lock.RUnlock()
	return block
}