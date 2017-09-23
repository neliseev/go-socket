package socket

import (
	"net"
)

type SessionUDP struct {
	raddr   *net.UDPAddr
	context []byte
}

func (s *SessionUDP) RemoteAddr() net.Addr { return s.raddr }

func ReadFromSessionUDP(conn *net.UDPConn, b []byte) (int, *SessionUDP, error) {
	oob := make([]byte, 40)
	n, oobn, _, raddr, err := conn.ReadMsgUDP(b, oob)
	if err != nil {
		return n, nil, err
	}

	return n, &SessionUDP{raddr, oob[:oobn]}, err
}

func WriteToSessionUDP(conn *net.UDPConn, b []byte, session *SessionUDP) (int, error) {
	n, _, err := conn.WriteMsgUDP(b, session.context, session.raddr)

	return n, err
}
