package network

type TestError struct {
    msg string
}

func (err *TestError) Error() string {
    return err.msg
}

func TestNode() error {
	node := NewNode("a")

	err := node.Start()
	if err != nil {
		return err
	}

	return nil
}