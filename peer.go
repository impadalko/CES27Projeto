// Implementation of a network of symmetrical peers that can start, join or leave a network.
// The network is designed to be fully connected, that is, all peer connected to each other.

package main

import (
	"fmt"
	"net"
	"os"
	"bufio"
	"strings"
	"sync"
)

/* RANDOM STRING GENERATOR */
import (
	"math/rand"
	"time"
)
func init() {
	rand.Seed(time.Now().UnixNano())
}
var alphabet []rune = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_")
func RandomString(length int) string {
	res := make([]rune, length)
	for i := range res {
		res[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(res)
}

/* GLOBAL VARIABLES */

var MyId string
var MyAddr string

type Peer struct {
	id string
	addr string
	conn net.Conn
}

// map: peerId string => peer Peer
var Peers = map[string]Peer{}

// map: conn net.Conn => peerId string | ""
var Conns = map[net.Conn]string{}

// the native map implementation is not thread-safe.
// we are using multiple goroutines that are able to access and modify the maps,
// so we need locks for synchronization
var PeersLock = sync.RWMutex{}
var ConnsLock = sync.RWMutex{}

/* MAIN PROGRAM */

func main() {
	MyId = RandomString(8)
	fmt.Printf("MyId=%s\n", MyId)

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Println("CRASH: Could not start server")
		fmt.Println(err)
		os.Exit(1)
	}
	MyAddr = listener.Addr().String()
	fmt.Println("Server started at", MyAddr)

	go AcceptConnections(listener)

	// accept user commands from terminal
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		split := strings.Split(line, " ")

		if len(split) == 2 && split[0] == "join" {
			// the current peer will request to join the network of the target peer

			peerAddr := split[1]
			conn, err := net.Dial("tcp", peerAddr)
			if err != nil {
				fmt.Printf("ERROR: Failed to connect to peer %s\n", peerAddr)
				fmt.Println(err)
			} else {
				go HandleConnection(conn)
				fmt.Fprintf(conn, "REQUEST %s %s\n", MyId, MyAddr)
				fmt.Fprintf(conn, "LIST\n")
			}

		} else if len(split) == 1 && split[0] == "peers" {
			// print a list of all peers

			if len(Peers) == 0 {
				fmt.Println("No peers")
			} else {
				fmt.Println("Peers:")
				PeersLock.RLock()
				for _, peer := range Peers {
					fmt.Printf("  id=%s addr=%s\n", peer.id, peer.addr)
				}
				PeersLock.RUnlock()
			}

		} else if len(split) == 1 && split[0] == "conns" {
			// print a list of all connections

			ConnsLock.RLock()
			if len(Conns) == 0 {
				fmt.Println("No connections")
			} else {
				fmt.Println("Connections:")
				for conn, peerId := range Conns {
					fmt.Printf("  local=%s remote=%s peerId=%s\n",
						conn.LocalAddr(), conn.RemoteAddr(), peerId)
				}
			}
			ConnsLock.RUnlock()

		} else {
			fmt.Println("Invalid command")
		}
	}
}

func AcceptConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("ERROR: Could not accept connection")
			fmt.Println(err)
			continue
		}
		go HandleConnection(conn)
	}
}

// connections handled concurrently to each other and stay open during the lifetime of the newtork
func HandleConnection(conn net.Conn) {
	ConnSet(conn, "") // allows to print a list of connections
	connLabel := conn.RemoteAddr().String() // allows to more helpfully label events in the connection

	// Id and Addr of the peer on the other side of the connection, if it is known
	connPeerId := ""
	connPeerAddr := ""

	fmt.Println("Connected:", connLabel)

	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		fmt.Printf("Message from %s: %s\n", connLabel, line)
		split := strings.Split(line, " ")

		if len(split) == 3 && split[0] == "REQUEST" {
			// the other peer is requesting the current peer to add as peer

			connPeerId = split[1]
			connPeerAddr = split[2]
			ConnSet(conn, connPeerId)
			connLabel = connPeerId

			if connPeerId == MyId {
				fmt.Println("SKIP: Can't add itself as peer")
				conn.Close()
			} else {
				_, ok := PeerGet(connPeerId)
				if ok {
					fmt.Println("ERROR: Peer %s is already a peer", connPeerId)
					conn.Close()
				} else {
					// accept requesting peer as peer
					PeerSet(connPeerId, Peer{connPeerId, connPeerAddr, conn})
					fmt.Fprintf(conn, "ACCEPTED %s %s\n", MyId, MyAddr)
				}
			}
		} else if len(split) == 3 && split[0] == "ACCEPTED" {
			// the other peer accepted the current peer as a peer
			
			connPeerId = split[1]
			connPeerAddr = split[2]
			ConnSet(conn, connPeerId)
			connLabel = connPeerId

			if connPeerId == MyId {
				fmt.Println("SKIP: Can't add itself as peer")
				conn.Close()
			} else {
				_, ok := PeerGet(connPeerId)
				if ok {
					fmt.Println("ERROR: Peer %s is already a peer", connPeerId)
					conn.Close()
				} else {
					PeerSet(connPeerId, Peer{connPeerId, connPeerAddr, conn})
				}
			}
		} else if len(split) == 1 && split[0] == "LIST" {
			// the other peer is requesting a list of all the other peers of the current peer
			PeersLock.RLock()
			for _, peer := range Peers {
				if peer.id == connPeerId {
					continue
				}
				fmt.Fprintf(conn, "PEER %s %s\n", peer.id, peer.addr)
			}
			PeersLock.RUnlock()
		} else if len(split) == 3 && split[0] == "PEER" {
			// the other peer sent information about one of his peers, as requested by
			// the current peer with the LIST message
			
			peerId := split[1]
			peerAddr := split[2]

			if peerId == MyId {
				fmt.Println("SKIP: Can't add itself as peer")
			} else {
				_, ok := PeerGet(peerId)
				if ok {
					fmt.Println("SKIP: Peer %s is already a peer", peerId)
				} else {
					conn, err := net.Dial("tcp", peerAddr)
					if err != nil {
						fmt.Printf("ERROR: Failed to connect to peer %s\n", peerAddr)
						fmt.Println(err)
					} else {
						go HandleConnection(conn)
						fmt.Fprintf(conn, "REQUEST %s %s\n", MyId, MyAddr)
					}
				}
			}
		} else {
			fmt.Println("Invalid message")
		}
	}
	fmt.Println("Disconnected:", connLabel)
	ConnDelete(conn)
	PeerDelete(connPeerId)
}

/* SYNCHRONIZED OPERATIONS ON MAPS */

func ConnSet(conn net.Conn, peerId string) {
	ConnsLock.Lock()
	Conns[conn] = peerId
	ConnsLock.Unlock()
}

func ConnDelete(conn net.Conn) {
	ConnsLock.Lock()
	delete(Conns, conn)
	ConnsLock.Unlock()
}

func PeerGet(peerId string) (Peer, bool) {
	PeersLock.RLock()
	peer, ok := Peers[peerId]
	PeersLock.RUnlock()
	return peer, ok
}

func PeerSet(peerId string, peer Peer) {
	PeersLock.Lock()
	Peers[peerId] = peer
	PeersLock.Unlock()
}

func PeerDelete(peerId string) {
	PeersLock.Lock()
	delete(Peers, peerId)
	PeersLock.Unlock()
}