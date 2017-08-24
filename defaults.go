package socket

import (
	"github.com/neliseev/logger"
	"time"
)

// Defaults TCP
const maxTCPQueries int = 256
const tcpIdleTimeout time.Duration = 60 * time.Second //

// Defaults UDP
const udpMsgSize int = 508   // RFC 791 (Min IP Size - Max IP Header Size - UDP Header Size)
const maxMsgSize int = 65535 // Max Packet size, limited by uint16

// Defaults Socket
const rtimeout time.Duration = 2 * time.Second // Socket read timeout

// Init logger subsystem
var log *logger.Params

func init() {
	// Initialization log system
	if err := logger.NewLogger(&logger.Params{}); err != nil {
		log.Critf("Can't init log subsystem: %s", err)
	}
}
