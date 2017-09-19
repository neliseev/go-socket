package socket

import (
	"github.com/neliseev/logger"
	"time"
)

const maxTCPQueries int = 256                         // maximum tcp queries
const tcpIdleTimeout time.Duration = 60 * time.Second // How much keep tcp socket open

const udpMsgSize int = 508   // RFC 791 (Min IP Size - Max IP Header Size - UDP Header Size)
const maxMsgSize int = 65535 // Max Packet size, limited by uint16

const rtimeout time.Duration = 2 * time.Second // Socket read timeout

var log *logger.Log // Using log subsystem

func init() {
	// Initialization log system
	var err error

	if log, err = logger.NewFileLogger("", 8); err != nil {
		panic(err)
	}
}
