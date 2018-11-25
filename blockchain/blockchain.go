package blockchain

import (
	"github.com/impadalko/CES27Projeto/blockchain/block"
)

type BlockChain struct {
	NextIndex int64
	LastHash  block.HashVal
	Blocks    []block.Block
}

func New(timestamp int64, GenesisData []byte) BlockChain {
	genesis := block.Block{
		0,
		block.HashVal{},
		timestamp,
		int32(len(GenesisData)),
		GenesisData,
	}

	bc := BlockChain{}
	bc.LastHash = genesis.Hash()
	bc.Blocks = append(bc.Blocks, genesis)
	bc.NextIndex = 1
	return bc
}

func (bc *BlockChain) AddBlock(timestamp int64, Data []byte) {
	block := block.Block{
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