package network

// Implementation of a network of symmetrical peers that can start, join or leave a network.
// The network is designed to be fully connected, that is, all peer connected to each other.

import (
	"fmt"
	"net"
	"bufio"
	"strings"
	"sync"
	"errors"
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

func (connInfo *ConnInfo) SendMessage(message string) error {
	_, err := fmt.Fprint(connInfo.Conn, message)
	return err
}

func (network *Network) SendMessage(peerId string, message string) error {
	peer, ok := network.PeerGet(peerId)
	if !ok {
		return errors.New("Peer not found")
	}
	_, err := fmt.Fprint(peer.Conn, message)
	return err
}

type Network struct {
	NodeId    string
	NodeAddr  string
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

	Handlers  map[string]func(connInfo *ConnInfo, args []string)
}

func NewNode(nodeId string) *Network {
	network := Network{}
	network.NodeId    = nodeId
	network.Peers     = map[string]Peer{}
	network.PeersLock = sync.RWMutex{}
	network.Conns     = map[net.Conn]*ConnInfo{}
	network.ConnsLock = sync.RWMutex{}
	network.Handlers = map[string]func(connInfo *ConnInfo, args []string){}
	return &network
}

func (network *Network) Listen() error {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}
	network.Listener = listener
	network.NodeAddr = listener.Addr().String()
	return nil
}

func (network *Network) Start() error {
	for {
		conn, err := network.AcceptConnection()
		if err != nil {
			return err
		}
		go network.StartHandleConnection(conn)
	}
}

func (network *Network) StartHandleConnection(conn net.Conn) {
	connInfo := network.HandleConnection(conn)
	for {
		msg, err := network.ReadNextMessage(connInfo)
		if err != nil {
			break
		}
		newConn, err := network.HandleMessage(connInfo, msg)
		if err != nil {
			fmt.Println(err)
		}
		if err == nil && newConn != nil {
			go network.StartHandleConnection(newConn)
		}
	}
	network.CleanupConnection(connInfo)
}

func (network *Network) AcceptConnection() (net.Conn, error) {
	return network.Listener.Accept()
}

func (network *Network) Close() error {
	return network.Listener.Close()
}

type ProtocolError struct {
	code string
	msg string
}

func (err *ProtocolError) Error() string {
	return fmt.Sprintf("%s: %s", err.code, err.msg)
}

// may return a new connection that must be handled
func (network *Network) HandleMessage(connInfo *ConnInfo, message string) (net.Conn, error) {
	// received message
	message = strings.TrimSpace(message)
	args := strings.Split(message, " ")

	messageType := args[0]

	if len(args) == 3 && messageType == "REQUEST" {
		// the other peer is requesting the current network to add it as peer

		connInfo.PeerId = args[1]
		connInfo.PeerAddr = args[2]

		if connInfo.PeerId == network.NodeId {
			connInfo.Conn.Close()
			return nil, &ProtocolError{"SelfPeer", "Can't add itself as peer"}
		} else {
			_, ok := network.PeerGet(connInfo.PeerId)
			if ok {
				connInfo.Conn.Close()
				return nil, &ProtocolError{"AlreadyPeer", "Requesting peer is already a peer"}
			} else {
				// accept requesting peer as peer
				network.PeerSet(connInfo.PeerId, Peer{connInfo.PeerId, connInfo.PeerAddr, connInfo.Conn})
				fmt.Fprintf(connInfo.Conn, "ACCEPTED %s %s\n", network.NodeId, network.NodeAddr)
				return nil, nil
			}
		}
	} else if len(args) == 3 && messageType == "ACCEPTED" {
		// the other peer accepted the current network as a peer
		
		connInfo.PeerId = args[1]
		connInfo.PeerAddr = args[2]

		if connInfo.PeerId == network.NodeId {
			connInfo.Conn.Close()
			return nil, &ProtocolError{"SelfPeer", "Can't add itself as peer"}
		} else {
			_, ok := network.PeerGet(connInfo.PeerId)
			if ok {
				connInfo.Conn.Close()
				return nil, &ProtocolError{"AlreadyPeer", "Requesting peer is already a peer"}
			} else {
				// add accepting peer as peer
				network.PeerSet(connInfo.PeerId, Peer{connInfo.PeerId, connInfo.PeerAddr, connInfo.Conn})
				return nil, nil
			}
		}
	} else if len(args) == 1 && messageType == "LIST" {
		// the other peer is requesting a list of all the other peers of the current network

		network.PeersLock.RLock()
		for _, peer := range network.Peers {
			if peer.Id == connInfo.PeerId {
				continue
			}
			fmt.Fprintf(connInfo.Conn, "PEER %s %s\n", peer.Id, peer.Addr)
		}
		network.PeersLock.RUnlock()
		return nil, nil

	} else if len(args) == 3 && messageType == "PEER" {
		// the other peer sent information about one of his peers, as requested by
		// the current network with the LIST message
		
		peerId := args[1]
		peerAddr := args[2]

		if peerId == network.NodeId {
			return nil, &ProtocolError{"SelfPeer", "Can't add itself as peer"}
		} else {
			_, ok := network.PeerGet(peerId)
			if ok {
				return nil, &ProtocolError{"AlreadyPeer", "Requesting peer is already a peer"}
			} else {
				conn, err := net.Dial("tcp", peerAddr)
				if err != nil {
					return nil, &ProtocolError{"FailConnect", "Failed to connect to peer"}
				} else {
					fmt.Fprintf(conn, "REQUEST %s %s\n", network.NodeId, network.NodeAddr)
					return conn, nil
				}
			}
		}
	} else if handler, ok := network.Handlers[messageType]; ok {
		handler(connInfo, args)
	} else {
		return nil, &ProtocolError{"InvalidMessage", "The message is invalid"}
	}
	return nil, nil
}

// connections are handled concurrently to each other and stay open during the lifetime of the newtork
func (network *Network) HandleConnection(conn net.Conn) *ConnInfo {
	connInfo := ConnInfo{}
	connInfo.Conn = conn

	// allows to get a list of connections
	network.ConnSet(conn, &connInfo)

	// the protocol of communication bewteen peers is in plain-text format,
	// with newlines '\n' at the end of each message
	connInfo.Reader = bufio.NewReader(conn)

	return &connInfo
}

func (network *Network) ReadNextMessage(connInfo *ConnInfo) (string, error) {
	line, err := connInfo.Reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return line, nil
}

func (network *Network) CleanupConnection(connInfo *ConnInfo) {
	network.PeerDelete(connInfo.PeerId)
	network.ConnDelete(connInfo.Conn)
}

/* SYNCHRONIZED OPERATIONS ON MAPS */

func (network *Network) ConnSet(conn net.Conn, connInfo *ConnInfo) {
	network.ConnsLock.Lock()
	network.Conns[conn] = connInfo
	network.ConnsLock.Unlock()
}

func (network *Network) ConnDelete(conn net.Conn) {
	network.ConnsLock.Lock()
	if _, ok := network.Conns[conn]; ok {
		delete(network.Conns, conn)
	}
	network.ConnsLock.Unlock()
}

func (network *Network) PeerGet(peerId string) (Peer, bool) {
	network.PeersLock.RLock()
	peer, ok := network.Peers[peerId]
	network.PeersLock.RUnlock()
	return peer, ok
}

func (network *Network) PeerSet(peerId string, peer Peer) {
	network.PeersLock.Lock()
	network.Peers[peerId] = peer
	network.PeersLock.Unlock()
	fmt.Printf("Peer connected %s\n", peerId)
}

func (network *Network) PeerDelete(peerId string) {
	network.PeersLock.Lock()
	if _, ok := network.Peers[peerId]; ok {
		delete(network.Peers, peerId)
		fmt.Printf("Peer disconnected %s\n", peerId)
	}
	network.PeersLock.Unlock()
}

/* NODE CONTROL */

// may return a new connection that must be handled
func (network *Network) JoinNetwork(peerAddr string) (net.Conn, error) {
	// the current peer will request to join the network of the target peer

	conn, err := net.Dial("tcp", peerAddr)
	if err != nil {
		return nil, &ProtocolError{"FailConnect", "Failed to connect to peer"}
	} else {
		fmt.Fprintf(conn, "REQUEST %s %s\n", network.NodeId, network.NodeAddr)
		fmt.Fprintf(conn, "LIST\n")
		return conn, nil
	}
}

func (network *Network) GetPeers() []Peer {
	network.PeersLock.RLock()
	peers := make([]Peer, len(network.Peers))
	i := 0
	for _, peer := range network.Peers {
		peers[i] = peer
		i++
	}
	network.PeersLock.RUnlock()
	return peers
}

func (network *Network) AddHandler(messageType string, handler func(connInfo *ConnInfo, args []string)) {
	network.Handlers[messageType] = handler
}