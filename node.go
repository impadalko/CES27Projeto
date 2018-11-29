package main

import (
	"fmt"
	"net"
	"github.com/impadalko/CES27Projeto/blockchain"
	"github.com/impadalko/CES27Projeto/network"
)

type Node struct {
	Network network.Network
	BlockChain blockchain.BlockChain
}

var node Node // find some way to share the node between handlers without global...

func NewNode(nodeId string, timestamp int64) *Node {
	node = Node{
		*network.NewNode(nodeId),
		blockchain.New(timestamp, []byte{}),
	}
	node.Network.AddHandler("BLOCKCHAIN", HandleBlockchain)
	node.Network.AddHandler("BLOCK", HandleBlock)
	return &node
}

func HandleBlockchain(connInfo *network.ConnInfo, args []string) {
	for _, block := range node.BlockChain.Blocks {
		msg := fmt.Sprintf("BLOCK %s\n", block.String())
		connInfo.SendMessage(msg)
	}
}

func HandleBlock(connInfo *network.ConnInfo, args []string) {
	block, err := blockchain.BlockFromString(args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	if block.Index == 0 {
		node.BlockChain = blockchain.NewFromBlock(block)
	} else {
		err = node.BlockChain.AddBlock(block)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (node *Node) Listen() error {
	return node.Network.Listen()
}

func (node *Node) NodeId() string {
	return node.Network.NodeId
}

func (node *Node) NodeAddr() string {
	return node.Network.NodeAddr
}

func (node *Node) JoinNetwork(peerAddr string) (net.Conn, error) {
	return node.Network.JoinNetwork(peerAddr)
}

func (node *Node) StartHandleConnection(conn net.Conn) {
	node.Network.StartHandleConnection(conn)
}

func (node *Node) Start() {
	node.Network.Start()
}