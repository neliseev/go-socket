package socket

import (
	"errors"
	"encoding/binary"
)

type Msg struct {
	Req  string
	Data []byte
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

	for i, b := range data {
		if b == s {
			m.Req  = string(data[:i])
			m.Data = data[i + 1:]

			break
		}

		i++
	}

	if m.Req == "" {
		return errors.New("Undefined request in message")
	}

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
