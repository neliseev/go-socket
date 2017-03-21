package socket

import (
	"io"
	"net"
	"encoding/binary"
	"bytes"
)

// Writer writes raw protoBuf reply messages.
// Each call to Write should send an entire message.
type Writer interface {
	io.Writer
}

// A ResponseWriter interface
type ResponseWriter interface {
	LocalAddr()    net.Addr     // LocalAddr returns net.Addr of the server
	RemoteAddr()   net.Addr     // RemoteAddr returns net.Addr of client that sent the current request.
	WriteMsg(*Msg) error        // WriteMsg writes a reply back to the client.
	Write([]byte)  (int, error) // Write writes a raw buffer back to the client.
	Close() error               // Close closes the connection.
}

type response struct {
	udp            *net.UDPConn      // i/o connection if UDP was used
	tcp            net.Conn          // i/o connection if TCP was used
	udpSession     *SessionUDP       // oob data to get egress interface right
	remoteAddr     net.Addr          // address of the client
	writer         Writer            // writer to output
}


// Write implements the ResponseWriter.Write method.
func (w *response) Write(m []byte) (int, error) {
	switch {
	case w.udp != nil:
		n, err := WriteToSessionUDP(w.udp, m, w.udpSession)

		return n, err
	case w.tcp != nil:
		lm := len(m)
		if lm < 2 {
			return 0, io.ErrShortBuffer
		}
		if lm > MaxMsgSize {
			return 0, errMsgLarge
		}

		l := make([]byte, 2, 2+lm)
		binary.BigEndian.PutUint16(l, uint16(lm))
		m = append(l, m...)

		n, err := io.Copy(w.tcp, bytes.NewReader(m))

		return int(n), err
	default:
		panic("Write switch fatal")
	}
}

// WriteMsg implements the ResponseWriter.WriteMsg method.
func (w *response) WriteMsg(m *Msg) (err error) {
	var data []byte

	// ToDo implement pack/unpack
	//data, err = m.Pack()
	if err != nil {
		return err
	}

	_, err = w.writer.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// LocalAddr implements the ResponseWriter.LocalAddr method.
func (w *response) LocalAddr() net.Addr {
	if w.tcp != nil {
		return w.tcp.LocalAddr()
	}
	return w.udp.LocalAddr()
}

// RemoteAddr implements the ResponseWriter.RemoteAddr method.
func (w *response) RemoteAddr() net.Addr { return w.remoteAddr }

// Close implements the ResponseWriter.Close method
func (w *response) Close() error {
	// Can't close the udp conn, as that is actually the listener.
	if w.tcp != nil {
		e := w.tcp.Close()
		w.tcp = nil
		return e
	}
	return nil
}
