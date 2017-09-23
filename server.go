package socket

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/neliseev/logger"
	"net"
	"sync"
	"time"
)

var log *logger.Log // Using log subsystem

func init() {
	// Initialization log system
	var err error

	if log, err = logger.NewFileLogger("", 8); err != nil {
		panic(err)
	}
}

const sockReadTimeout time.Duration = 60 * time.Second       // Socket read timeout
const defaultProtocol string = "tcp"                         // Default protocol for server
const defaultTCPIdleTimeout time.Duration = 60 * time.Second // How much keep tcp socket open
const defaultTCPMaxQueries int = 256                         // Max tcp queries
const defaultTCPMaxPacketSize int = 65535                    // Max TCP Packet size, limited by uint16
const defaultUDPMaxPacketSize int = 508                      // RFC 791 (Min IP Size - Max IP Header Size - UDP Header Size)

var tcpMaxPacketSize int

type Params struct {
	Addr             string
	Proto            string
	TCPIdleTimeout   time.Duration
	TCPMaxQueries    int
	TCPMaxPacketSize int
	UDPMaxPacketSize int
}

func NewServer(p *Params) *Server {
	address := p.Addr
	if address == "" {
		address = ":2017"
	}

	protocol := p.Proto
	if protocol == "" {
		protocol = defaultProtocol
	}

	tcpIdleTimeout := p.TCPIdleTimeout
	if tcpIdleTimeout == 0 {
		tcpIdleTimeout = defaultTCPIdleTimeout
	}

	tcpMaxQueries := p.TCPMaxQueries
	if tcpMaxQueries == 0 {
		tcpMaxQueries = defaultTCPMaxQueries
	}

	tcpMaxPacketSize = p.TCPMaxPacketSize
	if tcpMaxPacketSize == 0 || tcpMaxPacketSize > defaultTCPMaxPacketSize {
		tcpMaxPacketSize = defaultTCPMaxPacketSize
	}

	udpMaxPacketSize := p.UDPMaxPacketSize
	if udpMaxPacketSize == 0 || udpMaxPacketSize > defaultUDPMaxPacketSize {
		udpMaxPacketSize = defaultUDPMaxPacketSize
	}

	return &Server{
		addr:           address,
		proto:          protocol,
		tcpIdleTimeout: tcpIdleTimeout,
		tcpMaxQueries:  tcpMaxQueries,
		tcpPacketSize:  tcpMaxPacketSize,
		udpPacketSize:  udpMaxPacketSize,
	}
}

type Server struct {
	Handler        Handler
	addr           string
	proto          string
	tcpIdleTimeout time.Duration
	tcpMaxQueries  int
	tcpPacketSize  int
	udpPacketSize  int
	listenerTCP    net.Listener
	listenerUDP    net.PacketConn
	inFlight       sync.WaitGroup
	started        bool
	sync.RWMutex
}

func (srv *Server) ListenAndServe() error {
	srv.Lock()
	defer srv.Unlock()

	if srv.started {
		return errors.New("socket server already started")
	}

	switch srv.proto {
	case "tcp", "tcp4", "tcp6":
		var err error

		a, err := net.ResolveTCPAddr(srv.proto, srv.addr)
		if err != nil {
			return err
		}

		l, err := net.ListenTCP(srv.proto, a)
		if err != nil {
			return fmt.Errorf("socket server ListenTCP: %s", err)
		}

		srv.listenerTCP = l
		srv.started = true
		srv.Unlock()
		if err = srv.serveTCP(l); err != nil {
			return err
		}
		srv.Lock()

		return nil
	case "udp", "udp4", "udp6":
		var err error

		a, err := net.ResolveUDPAddr(srv.proto, srv.addr)
		if err != nil {
			return err
		}

		l, err := net.ListenUDP(srv.proto, a)
		if err != nil {
			return fmt.Errorf("socket server ListenUDP: %s", err)
		}

		srv.listenerUDP = l
		srv.started = true
		srv.Unlock()
		if err = srv.serveUDP(l); err != nil {
			return err
		}
		srv.Lock()

		return nil
	default:
		return fmt.Errorf("socket server, can't start, incorrect proto: %s", srv.proto)
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
		srv.RLock()
		if !srv.started {
			return nil
		}
		srv.RUnlock()
		if err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Temporary() {
				log.Errf("Socket server: TCP Network: %s", err)

				continue
			}

			return err
		}

		m, err := reader.ReadTCP(rw, sockReadTimeout)
		if err != nil {
			log.Errf("Socket server, read tcp: %s", err)

			continue
		}

		srv.inFlight.Add(1)
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
		m, s, err := reader.ReadUDP(l, sockReadTimeout)
		srv.RLock()
		if !srv.started {
			return nil
		}
		srv.RUnlock()
		if err != nil {
			log.Errf("Socket server, read udp: %s", err)

			continue
		}

		srv.inFlight.Add(1)
		go srv.serve(s.RemoteAddr(), handler, m, l, s, nil)
	}
}

func (srv *Server) serve(a net.Addr, h Handler, m []byte, u *net.UDPConn, s *SessionUDP, t net.Conn) {
	defer srv.inFlight.Done()

	reader := Reader(&defaultReader{srv})
	writer := &response{
		udp:        u,
		tcp:        t,
		udpSession: s,
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
	if q > srv.tcpMaxQueries {
		writer.Close()

		return
	}

	// UDP, "close" and return
	if u != nil {
		writer.Close()

		return
	}

	m, err = reader.ReadTCP(writer.tcp, srv.tcpIdleTimeout)
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
			return nil, fmt.Errorf("can't read packet size flag: %s", err)
		}

		return nil, errPktFlag
	}

	length := binary.BigEndian.Uint16(l)
	log.Debugf("Incoming tcp packet from: %v, length: %v", conn.RemoteAddr(), length)
	if length == 0 {
		return nil, errPktLen
	}

	m := make([]byte, int(length))
	n, err = conn.Read(m[:int(length)])
	if err != nil || n == 0 {
		if err != nil {
			return nil, fmt.Errorf("can't read packet: %s", err)
		}

		return nil, errDataRead
	}

	i := n
	for i < int(length) {
		j, e := conn.Read(m[i:int(length)])
		if e != nil {
			return nil, fmt.Errorf("can't sort packet: %s", e)
		}

		i += j
	}

	n = i
	m = m[:n]
	defer log.Debugf("tcp packet was read from %v, size: %v", conn.RemoteAddr(), length, m)

	return m, nil
}

func (srv *Server) readUDP(conn *net.UDPConn, timeout time.Duration) ([]byte, *SessionUDP, error) {
	m := make([]byte, srv.udpPacketSize)
	n, s, err := ReadFromSessionUDP(conn, m)
	if err != nil || n == 0 {
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, errDataRead
	}
	log.Debugf("Socket server, incoming udp packet from: %v", s.RemoteAddr())

	m = m[:n]
	defer log.Debugf("Socket server, udp packet was read from %v", s.RemoteAddr())

	return m, s, nil
}

func (srv *Server) Shutdown() (err error) {
	srv.Lock()
	defer srv.Unlock()

	if !srv.started {
		return errors.New("socket server not started")
	}
	srv.started = false

	// Close UDP
	if srv.listenerUDP != nil {
		srv.listenerUDP.Close()
	}

	// Close TCP
	if srv.listenerTCP != nil {
		srv.listenerTCP.Close()
	}

	// Finalizing all active connections
	f := make(chan bool)
	go func() {
		srv.inFlight.Wait()
		f <- true
	}()

	select {
	case <-time.After(sockReadTimeout):
		return errors.New("can't stop socket server")
	case <-f:
		return nil
	}
}
