package socket

import (
	"errors"
	"net"
	"sync"
	"fmt"
	"time"
	"bytes"
	"encoding/binary"
)

// A Server defines parameters for running an EDGE server.
type Server struct {
	Addr        string
	Proto       string
	ListenerTCP net.Listener
	ListenerUDP net.PacketConn
	UDPSize     int            // Default 508 byte by RFC 791 (minimal IP length is are 576 byte)
	running     sync.WaitGroup
	lock        sync.RWMutex
	started     bool
}

// ToDo doc it
func (srv *Server) ListenAndServe() error {
	srv.lock.Lock()
	defer srv.lock.Unlock()

	if srv.started {
		return errors.New("EDGE server already started")
	}

	addr := srv.Addr
	if addr == "" {
		addr = ":2017"
	}

	switch srv.Proto {
	case "tcp", "tcp4", "tcp6":
		a, err := net.ResolveTCPAddr(srv.Proto, addr)
		if err != nil {
			return err
		}

		l, err := net.ListenTCP(srv.Proto, a)
		if err != nil {
			return fmt.Errorf("EDGE server ListenTCP: %s", err)
		}

		srv.ListenerTCP = l
		srv.started = true
		srv.lock.Unlock()
		if err := srv.serveTCP(l); err != nil {
			return err
		}
		srv.lock.Lock()

		return nil
	case "udp", "udp4", "udp6":
		a, err := net.ResolveUDPAddr(srv.Proto, addr)
		if err != nil {
			return err
		}

		l, err := net.ListenUDP(srv.Proto, a)
		if err != nil {
			return fmt.Errorf("EDGE server ListenUDP: %s", err)
		}

		srv.ListenerUDP = l
		srv.started = true
		srv.lock.Unlock()
		if err := srv.serveUDP(l); err != nil {
			return err
		}
		srv.lock.Lock()

		return err
	default:
		return fmt.Errorf("EDGE can't start server, incorrect proto: %s", srv.Proto)
	}
}

// Shutdown gracefully shuts down a server
func (srv *Server) Shutdown() error {
	srv.lock.Lock()
	if !srv.started {
		srv.lock.Unlock()

		return errors.New("EDGE server not started")
	}
	srv.started = false
	srv.lock.Unlock()

	// Close UDP
	if srv.ListenerUDP != nil {
		srv.ListenerUDP.Close()
	}

	// Close TCP
	if srv.ListenerTCP != nil {
		srv.ListenerTCP.Close()
	}

	// Finalizing all active connections
	f := make(chan bool)
	go func() {
		srv.running.Wait()
		f <- true
	}()

	select {
	case <-time.After(Rtimeout):
		// ToDO: try kill it?
		return errors.New("Can't stop server")
	case <-f:
		return nil
	}
}

// serveTCP
// Each request is handled in a separate goroutine.
func (srv *Server) serveTCP(l net.Listener) error {
	defer l.Close()

	reader := Reader(&defaultReader{srv})

	for {
		rw, err := l.Accept()
		if err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Temporary() {
				log.Errf("TCP Net ERR: %s", err)

				continue
			}

			return err
		}

		m, err := reader.ReadTCP(rw, Rtimeout)
		srv.lock.RLock()
		if !srv.started {
			srv.lock.RUnlock()

			return nil
		}
		srv.lock.RUnlock()
		if err != nil {
			log.Errf("EDGE Serve TCP: %s", err)

			continue
		}

		srv.running.Add(1)
		go srv.serve(rw.RemoteAddr(), m, nil, nil, rw)
	}
}

// serveUDP
// Each request is handled in a separate goroutine.
func (srv *Server) serveUDP(l *net.UDPConn) error {
	defer l.Close()

	reader := Reader(&defaultReader{srv})

	for {
		m, s, err := reader.ReadUDP(l, Rtimeout)
		srv.lock.RLock()
		if !srv.started {
			srv.lock.RUnlock()

			return nil
		}
		srv.lock.RUnlock()
		if err != nil {
			log.Errf("EDGE Serve UDP: %s", err)

			continue
		}

		srv.running.Add(1)
		go srv.serve(s.RemoteAddr(), m, l, nil, nil)
	}
}

// todo doc it || not implemented like dispatcher
func (srv *Server) serve(a net.Addr, m []byte, u *net.UDPConn, s *SessionUDP, t net.Conn) {
	defer srv.running.Done()

	req := string(bytes.Split(m, []byte("::"))[0])
	log.Debugf("REQ is: %s", req)
	msgRaw := bytes.Split(m, []byte("::"))[1]

	log.Debugf("MSG: %s", msgRaw)

}

// todo doc it
func (srv *Server) readTCP(conn net.Conn, timeout time.Duration) ([]byte, error) {
	conn.SetReadDeadline(time.Now().Add(timeout))

	l := make([]byte, 2)
	n, err := conn.Read(l)
	if err != nil || n != 2 {
		if err != nil {
			return nil, err
		}

		return nil, errFlagSize
	}

	length := binary.BigEndian.Uint16(l)
	if length == 0 {
		return nil, errFlagLen
	}

	m := make([]byte, int(length))
	n, err = conn.Read(m[:int(length)])
	if err != nil || n == 0 {
		if err != nil {
			return nil, err
		}

		return nil, errDataRead
	}

	i := n
	for i < int(length) {
		j, err := conn.Read(m[i:int(length)])
		if err != nil {
			return nil, err
		}

		i += j
	}

	n = i
	m = m[:n]

	return m, nil
}

// ToDo Doc IT || not tested
func (srv *Server) readUDP(conn *net.UDPConn, timeout time.Duration) ([]byte, *SessionUDP, error) {
	m := make([]byte, 508)
	n, s, err := ReadFromSessionUDP(conn, m)
	if err != nil || n == 0 {
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, errDataRead
	}
	m = m[:n]

	return m, s, nil
}
