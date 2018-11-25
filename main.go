package main

import (
	"fmt"

	"github.com/impadalko/CES27Projeto/blockchain"
	"github.com/impadalko/CES27Projeto/blockchain/block"
	//"github.com/impadalko/CES27Projeto/util"
)

func main() {
	bc := blockchain.New(0xFFFF)
	bc.AddBlock(0xFFFF, []byte{16, 32, 64})

	for _, b := range bc.Blocks {
		serial := b.ToString()
		fmt.Println(serial)
		fmt.Println()
		b2, _ := block.FromString(serial)
		fmt.Println(b2.ToString())
		fmt.Println()
	}
}