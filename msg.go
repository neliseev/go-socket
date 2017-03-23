package socket

import (
	"errors"
	"encoding/binary"
)

type Msg struct {
	Req  string
	Data []byte
}

type Packet struct {
	packetSize uint16 // full packet size
	headerSize uint16 // header size
	header     []byte // Header
	data       []byte // Data
}

type packet struct {
	size uint16 // packet size
	data []byte // Operation
}

// Unpack binary message to Msq structure
// s is are separator
func (m *Msg) Unpack(data []byte, s byte) error {
	if s == 0 {
		return errors.New("Separator size oferflow, should be 1")
	}

	headerLength := binary.BigEndian.Uint16(data[:2])
	if headerLength == 0 {
		return errHeaderLen
	}

	data = data[2:] // Remove header size from data

	// Unpack message
	m.Req  = string(data[:headerLength])
	// ToDo Test it
	m.Data = data[headerLength:]

	return nil
}

func (m *Msg) Pack() ([]byte, error) {
	rawData := append([]byte(""), []byte(m.Req)...)
	rawData = append(rawData, []byte("::")...)
	rawData = append(rawData, m.Data...)
	l := len(rawData)
	pkt := &packet{size: uint16(l), data: rawData}

	var buf []byte = make([]byte, 2 + l)
	offset := 0
	binary.BigEndian.PutUint16(buf[offset:], pkt.size)
	offset += 2
	copy(buf[offset:], pkt.data)

	return buf, nil
}
