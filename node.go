package main

import (
	"fmt"
	"net"
	"github.com/impadalko/CES27Projeto/blockchain"
	"github.com/impadalko/CES27Projeto/network"
)

type Node struct {
	Network    network.Network
	BlockChain blockchain.BlockChain
}

var node Node // find some way to share the node between handlers without global...

func NewNode(nodeId string, timestamp int64) *Node {
	node = Node{
		*network.NewNode(nodeId),
		blockchain.New(timestamp, []byte{}),
	}
	node.Network.AddHandler("REQUEST-BLOCKCHAIN", HandleRequestBlockchain)
	node.Network.AddHandler("BLOCK-ADD", HandleBlockAddMessage)
	return &node
}

func HandleRequestBlockchain(connInfo *network.ConnInfo, args []string) {
	// the peer requested for all the blocks of the blockchain of the current node to be sent back
	for _, block := range node.BlockChain.Blocks {
		msg := fmt.Sprintf("BLOCK-ADD %s\n", block.String())
		connInfo.SendMessage(msg)
	}
}

func HandleBlockAddMessage(connInfo *network.ConnInfo, args []string) {
	// the peer sent a block to be added to the blockchain of the current node
	block, err := blockchain.BlockFromString(args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	if block.Index == 0 {

		// replace the blockchain of the current node with and empty blockchain
		// starting with the received block
		node.BlockChain = blockchain.NewFromBlock(block)

	} else if block.Index == node.BlockChain.NextIndex &&
		block.PreviousHash == node.BlockChain.LastHash {

		// add the new block to the end of the blockchain of the current node
		err = node.BlockChain.AddBlock(block)
		if err != nil {
			fmt.Println(err)
		}

	} else if block.Index > node.BlockChain.NextIndex {

		// the current node is behind the blockchain of the peer,
		// so request peer to send the full blockchain
		connInfo.SendMessage("REQUEST-BLOCKCHAIN")

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