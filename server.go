package socket

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

type Server struct {
	Addr          string
	Proto         string
	ListenerTCP   net.Listener
	IdleTimeout   time.Duration
	MaxTCPQueries int //
	ListenerUDP   net.PacketConn
	UDPSize       int // UDPSize is default 508 byte by RFC 791 (minimal IP length is are 576 byte)
	Handler       Handler
	running       sync.WaitGroup //
	lock          sync.RWMutex
	started       bool
}

func (srv *Server) ListenAndServe() error {
	srv.lock.Lock()
	defer srv.lock.Unlock()

	if srv.started {
		return errors.New("Socket server already started")
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
			return fmt.Errorf("Socket server ListenTCP: %s", err)
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
			return fmt.Errorf("Socket server ListenUDP: %s", err)
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
		return fmt.Errorf("Socket can't start server, incorrect proto: %s", srv.Proto)
	}
}

func (srv *Server) serveTCP(l net.Listener) error {
	defer l.Close()

	reader := Reader(&defaultReader{srv})
	handler := srv.Handler
	if handler == nil {
		handler = DefaultServeMux
	}

	for {
		rw, err := l.Accept()
		log.Debugf("incomming TCP connection from: %s", rw.RemoteAddr())
		if err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Temporary() {
				log.Errf("TCP Network: %s", err)

				continue
			}

			return err
		}

		m, err := reader.ReadTCP(rw, rtimeout)
		srv.lock.RLock()
		if !srv.started {
			srv.lock.RUnlock()

			return nil
		}
		srv.lock.RUnlock()
		if err != nil {
			log.Errf("Socket Serve TCP: %s", err)

			continue
		}

		srv.running.Add(1)
		go srv.serve(rw.RemoteAddr(), handler, m, nil, nil, rw)
	}
}

func (srv *Server) serveUDP(l *net.UDPConn) error {
	defer l.Close()

	reader := Reader(&defaultReader{srv})
	handler := srv.Handler
	if handler == nil {
		handler = DefaultServeMux
	}

	for {
		m, s, err := reader.ReadUDP(l, rtimeout)
		log.Debugf("Incomming UDP connection from: %s", s.RemoteAddr())
		srv.lock.RLock()
		if !srv.started {
			srv.lock.RUnlock()

			return nil
		}
		srv.lock.RUnlock()
		if err != nil {
			log.Errf("Socket Serve UDP: %s", err)

			continue
		}

		srv.running.Add(1)
		go srv.serve(s.RemoteAddr(), handler, m, l, nil, nil)
	}
}

func (srv *Server) serve(a net.Addr, h Handler, m []byte, u *net.UDPConn, s *SessionUDP, t net.Conn) {
	defer srv.running.Done()

	reader := Reader(&defaultReader{srv})
	writer := &response{
		udp:        u,
		tcp:        t,
		remoteAddr: a,
	}

	// Init TCP queue
	q := 0

	////
	// Redo Label
Redo:
	req := new(Msg)
	err := req.Unpack(m)
	if err != nil {
		log.Err(err)

		goto Exit
	}

	h.Serve(writer, req)

	////
	// Exit Label
Exit:
	if writer.tcp == nil {
		return
	}

	// close socket after this many queries
	maxQueries := maxTCPQueries
	if srv.MaxTCPQueries != 0 {
		maxQueries = srv.MaxTCPQueries
	}
	if q > maxQueries {
		writer.Close()
		return
	}

	// UDP, "close" and return
	if u != nil {
		writer.Close()
		return
	}

	idleTimeout := tcpIdleTimeout
	if srv.IdleTimeout != 0 {
		idleTimeout = srv.IdleTimeout
	}

	m, err = reader.ReadTCP(writer.tcp, idleTimeout)
	if err == nil {
		q++

		goto Redo
	}

	writer.Close()

	return
}

func (srv *Server) readTCP(conn net.Conn, timeout time.Duration) ([]byte, error) {
	conn.SetReadDeadline(time.Now().Add(timeout))

	l := make([]byte, 2)
	n, err := conn.Read(l)
	if err != nil || n != 2 {
		if err != nil {
			return nil, fmt.Errorf("Can't read packet size flag: %s", err)
		}

		return nil, errPktFlag
	}

	length := binary.BigEndian.Uint16(l)
	log.Debugf("Incomming packet from: %v, length: %v", conn.RemoteAddr(), length)
	if length == 0 {
		return nil, errPktLen
	}

	m := make([]byte, int(length))
	n, err = conn.Read(m[:int(length)])
	if err != nil || n == 0 {
		if err != nil {
			return nil, fmt.Errorf("Can't read packet: %s", err)
		}

		return nil, errDataRead
	}

	i := n
	for i < int(length) {
		j, err := conn.Read(m[i:int(length)])
		if err != nil {
			return nil, fmt.Errorf("Can't sort packet: %s", err)
		}

		i += j
	}

	n = i
	m = m[:n]
	log.Debugf("Packet readed from %v, size: %v, data: %v", conn.RemoteAddr(), length, m)

	return m, nil
}

// ToDo add more debug
func (srv *Server) readUDP(conn *net.UDPConn, timeout time.Duration) ([]byte, *SessionUDP, error) {
	m := make([]byte, udpMsgSize)
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

func (srv *Server) Shutdown() error {
	srv.lock.Lock()
	if !srv.started {
		srv.lock.Unlock()

		return errors.New("Socket server not started")
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
	case <-time.After(rtimeout):
		// ToDO: try kill it?
		return errors.New("Can't stop server")
	case <-f:
		return nil
	}
}
