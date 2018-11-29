package blockchain

import (
	"fmt"
	"errors"
)

type BlockChain struct {
	NextIndex int64
	LastHash  HashVal
	Blocks    []Block
}

func New(timestamp int64, Data []byte) BlockChain {
	block := Block{
		0,
		HashVal{},
		timestamp,
		int32(len(Data)),
		Data,
	}

	bc := BlockChain{}
	bc.LastHash = block.Hash()
	bc.Blocks = append(bc.Blocks, block)
	bc.NextIndex = 1
	return bc
}

func NewFromBlock(block Block) BlockChain {
	fmt.Println("Genesis Block:", block)
	return BlockChain{
		1,
		block.Hash(),
		[]Block{block},
	}
}

func (bc *BlockChain) AddBlockFromData(timestamp int64, Data []byte) {
	block := Block{
		bc.NextIndex,
		bc.LastHash,
		timestamp,
		int32(len(Data)),
		Data,
	}
	bc.Blocks = append(bc.Blocks, block)
	bc.NextIndex++
	bc.LastHash = block.Hash()
}

func (bc *BlockChain) AddBlock(block Block) error {
	if block.PreviousHash != bc.LastHash || block.Index != bc.NextIndex {
		return errors.New("Block can't be added to blockchain")
	}
	fmt.Println("Added Block:", block)
	bc.Blocks = append(bc.Blocks, block)
	bc.NextIndex++
	bc.LastHash = block.Hash()
	return nil
}

func (bc BlockChain) VerifyConsistency() bool {
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