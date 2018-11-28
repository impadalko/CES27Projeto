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
	Id     string
	Addr   string
	Conn   net.Conn
}

type ConnInfo struct {
	PeerId     string        // the empty string is used when the peerId has not been resolved yet
	PeerAddr   string        // the empty string is used when the peerAddr has not been resolved yet
	Conn       net.Conn
	Reader     *bufio.Reader
}

type Node struct {
	Id        string
	Addr      string
	Listener  net.Listener

	// map: peerId string => peer Peer
	Peers     map[string]Peer
	PeersLock sync.RWMutex

	// map: conn net.Conn => conn Conn
	Conns     map[net.Conn]*ConnInfo
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
	node.Conns     = map[net.Conn]*ConnInfo{}
	node.ConnsLock = sync.RWMutex{}
	return &node
}

func (node *Node) Listen() error {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}
	node.Listener = listener
	node.Addr = listener.Addr().String()
	return nil
}

func (node *Node) Start() error {
	for {
		conn, err := node.AcceptConnection()
		if err != nil {
			return err
		}
		go node.StartHandleConnection(conn)
	}
}

func (node *Node) StartHandleConnection(conn net.Conn) {
	connInfo := node.HandleConnection(conn)
	for {
		msg, err := node.ReadNextMessage(connInfo)
		if err != nil {
			break
		}
		newConn, err := node.HandleMessage(connInfo, msg)
		if err != nil {
			fmt.Println(err)
		}
		if err == nil && newConn != nil {
			go node.StartHandleConnection(newConn)
		}
	}
	node.CleanupConnection(connInfo)
}

func (node *Node) AcceptConnection() (net.Conn, error) {
	return node.Listener.Accept()
}

func (node *Node) Close() error {
	return node.Listener.Close()
}

type ProtocolError struct {
	code string
	msg string
}

func (err *ProtocolError) Error() string {
	return fmt.Sprintf("%s: %s", err.code, err.msg)
}

// may return a new connection that must be handled
func (node *Node) HandleMessage(connInfo *ConnInfo, message string) (net.Conn, error) {
	// received message
	message = strings.TrimSpace(message)
	split := strings.Split(message, " ")

	if len(split) == 3 && split[0] == "REQUEST" {
		// the other peer is requesting the current node to add it as peer

		connInfo.PeerId = split[1]
		connInfo.PeerAddr = split[2]

		if connInfo.PeerId == node.Id {
			connInfo.Conn.Close()
			return nil, &ProtocolError{"SelfPeer", "Can't add itself as peer"}
		} else {
			_, ok := node.PeerGet(connInfo.PeerId)
			if ok {
				connInfo.Conn.Close()
				return nil, &ProtocolError{"AlreadyPeer", "Requesting peer is already a peer"}
			} else {
				// accept requesting peer as peer
				node.PeerSet(connInfo.PeerId, Peer{connInfo.PeerId, connInfo.PeerAddr, connInfo.Conn})
				fmt.Fprintf(connInfo.Conn, "ACCEPTED %s %s\n", node.Id, node.Addr)
				return nil, nil
			}
		}
	} else if len(split) == 3 && split[0] == "ACCEPTED" {
		// the other peer accepted the current node as a peer
		
		connInfo.PeerId = split[1]
		connInfo.PeerAddr = split[2]

		if connInfo.PeerId == node.Id {
			connInfo.Conn.Close()
			return nil, &ProtocolError{"SelfPeer", "Can't add itself as peer"}
		} else {
			_, ok := node.PeerGet(connInfo.PeerId)
			if ok {
				connInfo.Conn.Close()
				return nil, &ProtocolError{"AlreadyPeer", "Requesting peer is already a peer"}
			} else {
				// add accepting peer as peer
				node.PeerSet(connInfo.PeerId, Peer{connInfo.PeerId, connInfo.PeerAddr, connInfo.Conn})
				return nil, nil
			}
		}
	} else if len(split) == 1 && split[0] == "LIST" {
		// the other peer is requesting a list of all the other peers of the current node

		node.PeersLock.RLock()
		for _, peer := range node.Peers {
			if peer.Id == connInfo.PeerId {
				continue
			}
			fmt.Fprintf(connInfo.Conn, "PEER %s %s\n", peer.Id, peer.Addr)
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
					return conn, nil
				}
			}
		}
	}
	return nil, &ProtocolError{"InvalidMessage", "The message is invalid"}
}

// connections are handled concurrently to each other and stay open during the lifetime of the newtork
func (node *Node) HandleConnection(conn net.Conn) *ConnInfo {
	connInfo := ConnInfo{}
	connInfo.Conn = conn

	// allows to get a list of connections
	node.ConnSet(conn, &connInfo)

	// the protocol of communication bewteen peers is in plain-text format,
	// with newlines '\n' at the end of each message
	connInfo.Reader = bufio.NewReader(conn)

	return &connInfo
}

func (node *Node) ReadNextMessage(connInfo *ConnInfo) (string, error) {
	line, err := connInfo.Reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return line, nil
}

func (node *Node) CleanupConnection(connInfo *ConnInfo) {
	node.PeerDelete(connInfo.PeerId)
	node.ConnDelete(connInfo.Conn)
}

/* SYNCHRONIZED OPERATIONS ON MAPS */

func (node *Node) ConnSet(conn net.Conn, connInfo *ConnInfo) {
	node.ConnsLock.Lock()
	node.Conns[conn] = connInfo
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

// may return a new connection that must be handled
func (node *Node) JoinNetwork(peerAddr string) (net.Conn, error) {
	// the current peer will request to join the network of the target peer

	conn, err := net.Dial("tcp", peerAddr)
	if err != nil {
		return nil, &ProtocolError{"FailConnect", "Failed to connect to peer"}
	} else {
		fmt.Fprintf(conn, "REQUEST %s %s\n", node.Id, node.Addr)
		fmt.Fprintf(conn, "LIST\n")
		return conn, nil
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

/*func (node *Node) GetConns() []ConnInfo {
	node.ConnsLock.RLock()
	conns := make([]ConnInfo, len(node.Conns))
	i := 0
	for conn, connInfo := range node.Conns {
		conns[i] = *connInfo
		conns[i].reader = nil
		i++
	}
	node.ConnsLock.RUnlock()
	return conns
}*/