package main

import (
	"fmt"
	"net"
	"github.com/impadalko/CES27Projeto/blockchain"
	"github.com/impadalko/CES27Projeto/network"
)

type Node struct {
	Network    *network.Network
	BlockChain *blockchain.BlockChain
}

var node Node // FIXME find some way to share the node between handlers without global...

func NewNode(nodeId string, timestamp int64) *Node {
	node = Node{
		network.NewNode(nodeId),
		blockchain.New(timestamp, []byte{}),
	}
	node.Network.AddHandler("REQUEST-BLOCKCHAIN", HandleRequestBlockchain)
	node.Network.AddHandler("BLOCK-ADD", HandleBlockAddMessage)
	return &node
}

func HandleRequestBlockchain(connInfo *network.ConnInfo, args []string) {
	// the peer requested for all the blocks of the blockchain of the current node to be sent back
	node.BlockChain.Lock.RLock()
	for _, block := range node.BlockChain.Blocks {
		msg := fmt.Sprintf("BLOCK-ADD %s\n", block.String())
		connInfo.SendMessage(msg)
	}
	node.BlockChain.Lock.RUnlock()
}

func HandleBlockAddMessage(connInfo *network.ConnInfo, args []string) {
	// FIXME node.BlockChain must be handled with mutexes

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

	} else if block.Index >= node.BlockChain.NextIndex {

		// the current node is behind the blockchain of the peer,
		// so request peer to send the full blockchain
		connInfo.SendMessage("REQUEST-BLOCKCHAIN\n")

	} else {

		fmt.Println("WARNING: Ignored Invalid block")
		fmt.Println("  Index:     ", block.Index)
		fmt.Println("  Hash:      ", block.Hash().String()[:8])
		fmt.Println("  Timestamp: ", block.Timestamp)
		fmt.Println("  Data:      ", string(block.Data))
		fmt.Println()
		
	}
}

func (node *Node) PrintInfo() {
	fmt.Println("NodeId:  ", node.Network.NodeId)
	fmt.Println("NodeAddr:", node.Network.NodeAddr)
	fmt.Println()
}

func (node *Node) PrintPeers() {
	node.Network.PeersLock.RLock()
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
	node.Network.PeersLock.RUnlock()
}

func (node *Node) PrintConns() {
	node.Network.ConnsLock.RLock()
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
	node.Network.ConnsLock.RUnlock()
}

func (node *Node) AddBlockFromData(timestamp int64, data []byte) {
	node.BlockChain.AddBlockFromData(int64, data)
}

func (node *Node) Broadcast(msg string) {
	node.Network.Broadcast(msg)
}

func (node *Node) Listen() error {
	return node.Network.Listen()
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

func (node *Node) PrintBlocks() {
	node.BlockChain.PrintBlocks()
}