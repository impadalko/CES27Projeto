package main

import (
	"fmt"
	"net"
	"github.com/impadalko/CES27Projeto/blockchain"
	"github.com/impadalko/CES27Projeto/network"
)

type Node struct {
	network network.Network
	blockChain blockchain.BlockChain
}

var node Node // find some way to share the node between handlers without global...

func NewNode(nodeId string, timestamp int64) *Node {
	node = Node{
		*network.NewNode(nodeId),
		blockchain.New(timestamp, []byte{}),
	}
	node.network.AddHandler("BLOCKCHAIN", HandleBlockchain)
	node.network.AddHandler("BLOCK", HandleBlock)
	return &node
}

func HandleBlockchain(connInfo *network.ConnInfo, args []string) {
	for _, block := range node.blockChain.Blocks {
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
		node.blockChain = blockchain.NewFromBlock(block)
	} else {
		err = node.blockChain.AddBlock(block)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (node *Node) Listen() error {
	return node.network.Listen()
}

func (node *Node) NodeId() string {
	return node.network.NodeId
}

func (node *Node) NodeAddr() string {
	return node.network.NodeAddr
}

func (node *Node) JoinNetwork(peerAddr string) (net.Conn, error) {
	return node.network.JoinNetwork(peerAddr)
}

func (node *Node) StartHandleConnection(conn net.Conn) {
	node.network.StartHandleConnection(conn)
}

func (node *Node) Start() {
	node.network.Start()
}