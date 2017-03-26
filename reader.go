package socket

import (
	"net"
	"time"
)

// Reader interface - implementing readTCP and ReadUDP methods
type Reader interface {
	ReadTCP(conn net.Conn, timeout time.Duration) ([]byte, error)
	ReadUDP(conn *net.UDPConn, timeout time.Duration) ([]byte, *SessionUDP, error)
}

// Adapter for Server
type defaultReader struct {
	*Server
}

// ReadTCP method - calling readTCP func
func (dr *defaultReader) ReadTCP(conn net.Conn, timeout time.Duration) ([]byte, error) {
	return dr.readTCP(conn, timeout)
}

// ReadUDP method - calling readUDP func
func (dr *defaultReader) ReadUDP(conn *net.UDPConn, timeout time.Duration) ([]byte, *SessionUDP, error) {
	return dr.readUDP(conn, timeout)
}
