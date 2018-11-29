package network

import (
	"fmt"
	"net"
)

func (network *Network) SetConn(conn net.Conn, connInfo *ConnInfo) {
	network.ConnsLock.Lock()
	network.Conns[conn] = connInfo
	network.ConnsLock.Unlock()
}

func (network *Network) DeleteConn(conn net.Conn) {
	network.ConnsLock.Lock()
	if _, ok := network.Conns[conn]; ok {
		delete(network.Conns, conn)
	}
	network.ConnsLock.Unlock()
}

func (network *Network) GetPeer(peerId string) (Peer, bool) {
	network.PeersLock.RLock()
	peer, ok := network.Peers[peerId]
	network.PeersLock.RUnlock()
	return peer, ok
}

func (network *Network) SetPeer(peerId string, peer Peer) {
	network.PeersLock.Lock()
	network.Peers[peerId] = peer
	network.PeersLock.Unlock()
	fmt.Printf("Peer connected: %s\n\n", peerId)
}

func (network *Network) DeletePeer(peerId string) {
	network.PeersLock.Lock()
	if _, ok := network.Peers[peerId]; ok {
		delete(network.Peers, peerId)
		fmt.Printf("Peer disconnected: %s\n\n", peerId)
	}
	network.PeersLock.Unlock()
}

func (network *Network) GetHandler(messageType string) (func(connInfo *ConnInfo, args []string), bool) {
	network.HandlersLock.RLock()
	handler, ok := network.Handlers[messageType]
	network.HandlersLock.RUnlock()
	return handler, ok
}

func (network *Network) AddHandler(messageType string, handler func(connInfo *ConnInfo, args []string)) {
	network.HandlersLock.Lock()
	network.Handlers[messageType] = handler
	network.HandlersLock.Unlock()
}