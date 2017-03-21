package socket

import (
	"net"
	"time"
)

// Reader reads raw messages.
// Each call to ReadTCP or ReadUDP should return an raw message.
type Reader interface {
	// ReadTCP reads a raw message from a TCP connection.
	ReadTCP(conn net.Conn, timeout time.Duration) ([]byte, error)
	// ReadUDP reads a raw message from a UDP connection.
	ReadUDP(conn *net.UDPConn, timeout time.Duration) ([]byte, *SessionUDP, error)
}

// defaultReader is an adapter
// implementing Reader interface
type defaultReader struct {
	*Server
}

func (dr *defaultReader) ReadTCP(conn net.Conn, timeout time.Duration) ([]byte, error) {
	return dr.readTCP(conn, timeout)
}

func (dr *defaultReader) ReadUDP(conn *net.UDPConn, timeout time.Duration) ([]byte, *SessionUDP, error) {
	return dr.readUDP(conn, timeout)
}
