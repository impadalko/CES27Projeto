package block

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/binary"
	"bytes"
)

// SHA-256 outputs 256 bits (32 bytes)
type HashVal [32]byte

type Block struct {
	Index        int64   // 8 bytes
	PreviousHash HashVal // 32 bytes
	Timestamp    int64   // 8 bytes
	DataLen      int32   // 4 bytes
	Data         []byte  // DataLen bytes
}

// We will be using SHA-256 for hashing blocks
func (block Block) Hash() HashVal {
	buffer := bytes.Buffer{}
	binary.Write(&buffer, binary.LittleEndian, block.Data)
	return sha256.Sum256(buffer.Bytes())
}

func (hashVal HashVal) ToString() string {
	return hex.EncodeToString(hashVal[:])
}

func (block Block) ToString() string {
	buffer := bytes.Buffer{}
	binary.Write(&buffer, binary.LittleEndian, block.Index)
	binary.Write(&buffer, binary.LittleEndian, block.PreviousHash)
	binary.Write(&buffer, binary.LittleEndian, block.Timestamp)
	binary.Write(&buffer, binary.LittleEndian, block.DataLen)
	binary.Write(&buffer, binary.LittleEndian, block.Data)
	return hex.EncodeToString(buffer.Bytes())
}

func FromString(str string) (Block, error) {
	block := Block{}

	bin, err := hex.DecodeString(str)
	if err != nil {
		return block, err
	}

	reader := bytes.NewReader(bin)
	binary.Read(reader, binary.LittleEndian, &block.Index)
	binary.Read(reader, binary.LittleEndian, &block.PreviousHash)
	binary.Read(reader, binary.LittleEndian, &block.Timestamp)
	binary.Read(reader, binary.LittleEndian, &block.DataLen)

	remainingBytes := reader.Len()
	block.Data = make([]byte, remainingBytes)
	headerSize := len(bin) - remainingBytes
	copy(block.Data, bin[headerSize:]) // copy(dst, src)

	return block, nil
}