package network

import (
	"errors"
	"strings"
)

func TestNodeJoinNetwork() error {
	nodeA := NewNode("A")
	nodeB := NewNode("B")

	err := nodeA.Listen()
	if err != nil {
		return err
	}

	err = nodeB.Listen()
	if err != nil {
		return err
	}

	nodeBConn, err := nodeB.JoinNetwork(nodeA.NodeAddr)
	if err != nil {
		return err
	}
	nodeBConnInfo := nodeB.HandleConnection(nodeBConn)

	nodeAConn, err := nodeA.AcceptConnection()
	if err != nil {
		return err
	}
	nodeAConnInfo := nodeA.HandleConnection(nodeAConn)
	nodeAMsg, err := nodeA.ReadNextMessage(nodeAConnInfo)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(nodeAMsg, "PEER-REQUEST ") {
		return errors.New("Expected PEER-REQUEST message")
	}
	
	_, err = nodeA.HandleMessage(nodeAConnInfo, nodeAMsg)
	if err != nil {
		return err
	}

	nodeBMsg, err := nodeB.ReadNextMessage(nodeBConnInfo)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(nodeBMsg, "PEER-ACCEPTED ") {
		return errors.New("Expected PEER-ACCEPTED message")
	}

	_, err = nodeB.HandleMessage(nodeBConnInfo, nodeBMsg)
	if err != nil {
		return err
	}

	return nil
}