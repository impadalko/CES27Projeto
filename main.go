package main

import (
	"fmt"
	"os"
	"bufio"
	"strings"
	"strconv"

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
		fmt.Fprintf(conn, "REQUEST-BLOCKCHAIN\n")
		go node.StartHandleConnection(conn)
	}

	go node.Start()

	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		fmt.Println()
		text = strings.TrimSpace(text)
		split := strings.Split(text, " ")
		if len(split) == 0 {
			continue
		}
		command := split[0]

		if command == "info" {

			fmt.Println("NodeId:  ", node.NodeId())
			fmt.Println("NodeAddr:", node.NodeAddr())
			fmt.Println()

		} else if command == "peers" {

			if len(node.Network.Peers) == 0 {
				fmt.Println("No Peers")
				fmt.Println()
			} else {
				fmt.Printf("%-10s %s\n", "PeerId", "PeerAddr")
				for _, peer := range node.Network.Peers {
					fmt.Printf("%-10s %s\n", peer.Id, peer.Addr)
				}
				fmt.Println()
			}

		} else if command == "conns" {

			if len(node.Network.Conns) == 0 {
				fmt.Println("No Connections")
				fmt.Println()
			} else {
				fmt.Printf("%-22s %-22s %-10s %s\n", "RemoteAddr", "LocalAddr", "PeerId", "PeerAddr")
				for _, conn := range node.Network.Conns {
					fmt.Printf("%-22s %-22s %-10s %s\n",
						conn.Conn.RemoteAddr().String(), conn.Conn.LocalAddr().String(), conn.PeerId, conn.PeerAddr)
				}
				fmt.Println()
			}

		} else if command == "blocks" {

			fmt.Printf("%5s %-8s %-8s %-10s %s\n", "Index", "Hash", "PrevHash", "Timestamp", "Data")
			for _, block := range node.BlockChain.Blocks {
				fmt.Printf("%5d %8s %8s %10d %s\n", block.Index, block.Hash().String()[:8],
					block.PreviousHash.String()[:8], block.Timestamp, block.Data)
			}
			fmt.Println()

		} else if len(split) >= 2 && command == "add" {

			message := strings.Join(split[1:], " ")
			node.BlockChain.AddBlockFromData(util.Now(), []byte(message))

			fmt.Printf("%5s %-8s %-8s %-10s %s\n", "Index", "Hash", "PrevHash", "Timestamp", "Data")
			for _, block := range node.BlockChain.Blocks {
				fmt.Printf("%5d %8s %8s %10d %s\n", block.Index, block.Hash().String()[:8],
					block.PreviousHash.String()[:8], block.Timestamp, block.Data)
			}
			fmt.Println()

		} else if len(split) == 2 && command == "broadcast" {

			blockIndex, err := strconv.Atoi(split[1])
			if err == nil {
				block := node.BlockChain.Blocks[blockIndex]
				message := fmt.Sprintf("BLOCK-ADD %s\n", block.String())
				node.Network.Broadcast(message)
			}

		} else {
			fmt.Println("Invalid command")
			fmt.Println()
		}
	}
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