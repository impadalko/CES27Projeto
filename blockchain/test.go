package blockchain

import (
	"bytes"
	"errors"
)

func TestBlockToStringAndFromString() error {
	hash := HashVal{}
	for i := range hash {
		hash[i] = byte(i + 1)
	}

	block := Block{
		123123123,
		hash,
		321321321,
		5,
		[]byte{11, 22, 33, 44, 55},
	}

	blockToString := block.String()

	blockFromString, err := BlockFromString(blockToString)
	if err != nil {
		return err
	}

	if block.Index != blockFromString.Index ||
		block.PreviousHash != blockFromString.PreviousHash ||
		block.Timestamp != blockFromString.Timestamp ||
		block.DataLen != blockFromString.DataLen ||
		!bytes.Equal(block.Data, blockFromString.Data) {
		return errors.New("block mismatch")
	}

	return nil
}