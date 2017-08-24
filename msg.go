package socket

import (
	"encoding/binary"
	"errors"
)

// Msg is are message structure
type Msg struct {
	Req  string // Req - used like pattern for registered HandlerFunc
	Data []byte // Data - any data
}

// Packet - raw packet
type Packet struct {
	headerSize uint16 // header size
	header     []byte // Header
	data       []byte // Data
}

// Unpack method - unpacking binary packet to Msq structure
func (m *Msg) Unpack(data []byte) error {
	headerLength := binary.BigEndian.Uint16(data[:2])
	if headerLength == 0 {
		return errHeaderLen
	}

	// Remove header size from data
	data = data[2:]

	// Unpack message
	m.Req = string(data[:headerLength])
	m.Data = data[headerLength:]

	return nil
}

// Pack method - packing data to binary packet
func (m *Msg) Pack() ([]byte, error) {
	// Preparing packet
	hdr := []byte(m.Req)
	hdrSize := len(hdr)
	pktSize := hdrSize + len(m.Data)
	pkt := &Packet{
		headerSize: uint16(hdrSize),
		header:     hdr,
		data:       m.Data,
	}

	// Creating packet with size 2 bytes total size + 2 bytes header size + size header + data
	var buf []byte = make([]byte, 4+pktSize)
	// Put total packet size
	offset := 0
	binary.BigEndian.PutUint16(buf[offset:], pkt.headerSize)
	// Put header
	offset += 2
	if n := copy(buf[offset:], hdr); n == 0 {
		return nil, errors.New("Can't pack header to packet")
	}
	// Put data
	offset += hdrSize
	if n := copy(buf[offset:], pkt.data); n == 0 {
		return nil, errors.New("Can't pack data to packet")
	}

	return buf, nil
}
