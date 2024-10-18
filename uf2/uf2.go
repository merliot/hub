package uf2

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

const (
	UF2MagicStart0 = 0x0A324655 // "UF2\n"
	UF2MagicStart1 = 0x9E5D5157 // "Q\u009D"
	UF2BlockSize   = 512
)

// UF2File represents the structure of a UF2 file.
type UF2File struct {
	Blocks []UF2Block
}

// UF2Block represents a UF2 block.
type UF2Block struct {
	MagicStart0 uint32
	MagicStart1 uint32
	Flags       uint32
	TargetAddr  uint32
	PayloadSize uint32
	BlockNo     uint32
	NumBlocks   uint32
	FileSize    uint32 // or familyID;
	Data        [476]byte
	MagicEnd    uint32
}

// Read reads a UF2 file
func Read(filename string) (*UF2File, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var uf2File UF2File
	for {
		var block UF2Block
		err := binary.Read(file, binary.LittleEndian, &block)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		uf2File.Blocks = append(uf2File.Blocks, block)
	}

	return &uf2File, nil
}

// Write writes a new UF2 file
func (u *UF2File) Write(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, block := range u.Blocks {
		err := binary.Write(file, binary.LittleEndian, &block)
		if err != nil {
			return err
		}
	}

	return nil
}

// ReplaceBytes replaces old slice with new slice in the UF2File
func (u *UF2File) ReplaceBytes(oldBytes, newBytes []byte) {
	modifiedData := bytes.Replace(u.Bytes(), oldBytes, newBytes, 1)
	u.updateFile(modifiedData)
}

// Bytes creates a composite slice from all the Data slices in UF2File
func (u *UF2File) Bytes() []byte {
	var bytes []byte

	for _, block := range u.Blocks {
		span := block.PayloadSize
		bytes = append(bytes, block.Data[:span]...)
	}

	return bytes
}

// UpdateUF2File updates the UF2File blocks with the modified data.
func (u *UF2File) updateFile(modifiedData []byte) {
	for i := range u.Blocks {
		span := u.Blocks[i].PayloadSize
		copy(u.Blocks[i].Data[:span], modifiedData[:span])
		modifiedData = modifiedData[span:]
	}
}
