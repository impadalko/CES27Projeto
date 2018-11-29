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

func NewNode(nodeId string, timestamp int64) *Node {
	node := Node{
		*network.NewNode(nodeId),
		blockchain.New(timestamp, []byte{}),
	}
	node.network.AddHandler("BLOCKCHAIN", HandleBlockchain)
	return &node
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

func HandleBlockchain(connInfo *network.ConnInfo, args []string) {
	fmt.Println("BLOCHAIN Handler fired")
	connInfo.SendMessage("ACK")
}