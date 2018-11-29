package main

import (
	"fmt"
	"os"

	"github.com/impadalko/CES27Projeto/blockchain"
	"github.com/impadalko/CES27Projeto/network"
	"github.com/impadalko/CES27Projeto/sign"
	"github.com/impadalko/CES27Projeto/util"
)

func main() {
	node := NewNode(util.RandomString(8), util.Now())
	err := node.Listen()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("NodeId:  ", node.NodeId())
	fmt.Println("NodeAddr:", node.NodeAddr())
	fmt.Println()
	
	if len(os.Args) == 2 {
		peerAddr := os.Args[1]
		conn, err := node.JoinNetwork(peerAddr)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Fprintf(conn, "BLOCKCHAIN\n")
		go node.StartHandleConnection(conn)
	} else {
		for i := 0; i < 10; i++ {
			b := byte(i)
			node.BlockChain.AddBlockFromData(util.Now(), []byte{ b, b * 10 })
		}
	}

	node.Start()
}

func Tests() {
	var err error

	err = blockchain.TestBlockToStringAndFromString()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = sign.TestWriteAndReadPemFile()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = sign.TestSignAndVerify()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	
	err = network.TestNodeJoinNetwork()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("ALL TESTS PASSED")
}