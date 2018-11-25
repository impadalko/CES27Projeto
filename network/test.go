package network

import (
	//"fmt"
	"strings"
)

type TestError struct {
    msg string
}

func (err *TestError) Error() string {
    return err.msg
}

func TestNodeJoinNetwork() error {
	nodeA := NewNode("A")
	nodeB := NewNode("B")

	err := nodeA.Start()
	if err != nil {
		return err
	}

	err = nodeB.Start()
	if err != nil {
		return err
	}

	nodeBConn, err := nodeB.JoinNetwork(nodeA.Addr)
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
	if !strings.HasPrefix(nodeAMsg, "REQUEST ") {
		return &TestError{"Expected REQUEST message"}
	}
	
	_, err = nodeA.HandleMessage(nodeAConnInfo, nodeAMsg)
	if err != nil {
		return err
	}

	nodeBMsg, err := nodeB.ReadNextMessage(nodeBConnInfo)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(nodeBMsg, "ACCEPTED ") {
		return &TestError{"Expected ACCEPTED message"}
	}

	_, err = nodeB.HandleMessage(nodeBConnInfo, nodeBMsg)
	if err != nil {
		return err
	}

	return nil
}