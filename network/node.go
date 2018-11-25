package network

// Implementation of a network of symmetrical peers that can start, join or leave a network.
// The network is designed to be fully connected, that is, all peer connected to each other.

import (
	"fmt"
	"net"
	"bufio"
	"strings"
	"sync"
)

type Peer struct {
	Id   string
	Addr string
	Conn net.Conn
}

type Node struct {
	Id        string
	Addr      string
	Listener  net.Listener

	// map: peerId string => peer Peer
	Peers     map[string]Peer
	PeersLock sync.RWMutex

	// map: conn net.Conn => peerId string | ""
	// the empty string is used as the peerId when the peerId has not been resolved yet
	Conns     map[net.Conn]string
	ConnsLock sync.RWMutex

	// the native map implementation in Go is not thread-safe.
	// we may use multiple goroutines that are able to access and modify the same maps concurrently,
	// so we need locks for synchronization
}

func NewNode(id string) *Node {
	node := Node{}
	node.Id        = id
	node.Peers     = map[string]Peer{}
	node.PeersLock = sync.RWMutex{}
	node.Conns     = map[net.Conn]string{}
	node.ConnsLock = sync.RWMutex{}
	return &node
}

func (node *Node) Start() error {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}
	node.Listener = listener
	node.Addr = listener.Addr().String()
	return nil
}

func (node *Node) Close() {
	node.Listener.Close()
}

func (node *Node) AcceptConnection() (net.Conn, error) {
	return node.Listener.Accept()
}

func (node *Node) AcceptConnections() {
	for {
		conn, err := node.AcceptConnection()
		if err != nil {
			// TODO handle asynchronous errors somehow
			continue
		}
		go node.HandleConnection(conn)
	}
}

type ProtocolError struct {
	code string
	msg string
}

func (err *ProtocolError) Error() string {
	return fmt.Sprintf("%s: %s", err.code, err.msg)
}

// may return a new connection that must me handled
func (node *Node) HandleMessage(conn net.Conn, connPeerId *string, connPeerAddr *string, line string) (*net.Conn, error) {
	// received message
	line = strings.TrimSpace(line)
	split := strings.Split(line, " ")

	if len(split) == 3 && split[0] == "REQUEST" {
		// the other peer is requesting the current node to add it as peer

		*connPeerId = split[1]
		*connPeerAddr = split[2]
		node.ConnSet(conn, *connPeerId)

		if *connPeerId == node.Id {
			conn.Close()
			return nil, &ProtocolError{"SelfPeer", "Can't add itself as peer"}
		} else {
			_, ok := node.PeerGet(*connPeerId)
			if ok {
				conn.Close()
				return nil, &ProtocolError{"AlreadyPeer", "Requesting peer is already a peer"}
			} else {
				// accept requesting peer as peer
				node.PeerSet(*connPeerId, Peer{*connPeerId, *connPeerAddr, conn})
				fmt.Fprintf(conn, "ACCEPTED %s %s\n", node.Id, node.Addr)
				return nil, nil
			}
		}
	} else if len(split) == 3 && split[0] == "ACCEPTED" {
		// the other peer accepted the current node as a peer
		
		*connPeerId = split[1]
		*connPeerAddr = split[2]
		node.ConnSet(conn, *connPeerId)

		if *connPeerId == node.Id {
			conn.Close()
			return nil, &ProtocolError{"SelfPeer", "Can't add itself as peer"}
		} else {
			_, ok := node.PeerGet(*connPeerId)
			if ok {
				conn.Close()
				return nil, &ProtocolError{"AlreadyPeer", "Requesting peer is already a peer"}
			} else {
				// add accepting peer as peer
				node.PeerSet(*connPeerId, Peer{*connPeerId, *connPeerAddr, conn})
				return nil, nil
			}
		}
	} else if len(split) == 1 && split[0] == "LIST" {
		// the other peer is requesting a list of all the other peers of the current node

		node.PeersLock.RLock()
		for _, peer := range node.Peers {
			if peer.Id == *connPeerId {
				continue
			}
			fmt.Fprintf(conn, "PEER %s %s\n", peer.Id, peer.Addr)
		}
		node.PeersLock.RUnlock()
		return nil, nil

	} else if len(split) == 3 && split[0] == "PEER" {
		// the other peer sent information about one of his peers, as requested by
		// the current node with the LIST message
		
		peerId := split[1]
		peerAddr := split[2]

		if peerId == node.Id {
			return nil, &ProtocolError{"SelfPeer", "Can't add itself as peer"}
		} else {
			_, ok := node.PeerGet(peerId)
			if ok {
				return nil, &ProtocolError{"AlreadyPeer", "Requesting peer is already a peer"}
			} else {
				conn, err := net.Dial("tcp", peerAddr)
				if err != nil {
					return nil, &ProtocolError{"FailConnect", "Failed to connect to peer"}
				} else {
					fmt.Fprintf(conn, "REQUEST %s %s\n", node.Id, node.Addr)
					return &conn, nil
				}
			}
		}
	}
	return nil, &ProtocolError{"InvalidMessage", "The message is invalid"}
}

// connections handled concurrently to each other and stay open during the lifetime of the newtork
func (node *Node) HandleConnection(conn net.Conn) {
	// allows to print a list of connections
	node.ConnSet(conn, "")

	// Id and Addr of the peer on the other side of the connection, if it is known
	connPeerId := ""
	connPeerAddr := ""

	// the protocol of communication bewteen peers is in plain-text format,
	// with newlines '\n' at the end of each message
	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		node.HandleMessage(conn, &connPeerId, &connPeerAddr, line)
	}
	// Connection closed
	node.ConnDelete(conn)
	node.PeerDelete(connPeerId)
}

/* SYNCHRONIZED OPERATIONS ON MAPS */

func (node *Node) ConnSet(conn net.Conn, peerId string) {
	node.ConnsLock.Lock()
	node.Conns[conn] = peerId
	node.ConnsLock.Unlock()
}

func (node *Node) ConnDelete(conn net.Conn) {
	node.ConnsLock.Lock()
	if _, ok := node.Conns[conn]; ok {
		delete(node.Conns, conn)
	}
	node.ConnsLock.Unlock()
}

func (node *Node) PeerGet(peerId string) (Peer, bool) {
	node.PeersLock.RLock()
	peer, ok := node.Peers[peerId]
	node.PeersLock.RUnlock()
	return peer, ok
}

func (node *Node) PeerSet(peerId string, peer Peer) {
	node.PeersLock.Lock()
	node.Peers[peerId] = peer
	node.PeersLock.Unlock()
}

func (node *Node) PeerDelete(peerId string) {
	node.PeersLock.Lock()
	if _, ok := node.Peers[peerId]; ok {
		delete(node.Peers, peerId)
	}
	node.PeersLock.Unlock()
}

/* NODE CONTROL */

// may return a new connection that must me handled
func (node *Node) JoinPeer(peerAddr string) (*net.Conn, error) {
	// the current peer will request to join the network of the target peer

	conn, err := net.Dial("tcp", peerAddr)
	if err != nil {
		return nil, &ProtocolError{"FailConnect", "Failed to connect to peer"}
	} else {
		fmt.Fprintf(conn, "REQUEST %s %s\n", node.Id, node.Addr)
		fmt.Fprintf(conn, "LIST\n")
		return &conn, nil
	}
}

func (node *Node) GetPeers() []Peer {
	node.PeersLock.RLock()
	peers := make([]Peer, len(node.Peers))
	i := 0
	for _, peer := range node.Peers {
		peers[i] = peer
		i++
	}
	node.PeersLock.RUnlock()
	return peers
}

type ConnInfo struct {
	LocalAddr  string
	RemoteAddr string
	PeerId     string
}

func (node *Node) GetConns() []ConnInfo {
	node.ConnsLock.RLock()
	conns := make([]ConnInfo, len(node.Conns))
	i := 0
	for conn, peerId := range node.Conns {
		conns[i] = ConnInfo{conn.LocalAddr().String(), conn.RemoteAddr().String(), peerId}
		i++
	}
	node.ConnsLock.RUnlock()
	return conns
}