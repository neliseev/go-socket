package socket

import (
	"time"
	"github.com/neliseev/logger"
)

// Defaults TCP
const maxTCPQueries  int           = 256
const tcpIdleTimeout time.Duration = 60 * time.Second //

// Defaults UDP
const udpMsgSize     int           = 508              // RFC 791 (Min IP Size - Max IP Header Size - UDP Header Size)
const maxMsgSize     int           = 65535            // Max Packet size, limited by uint16

// Defaults Socket
const rtimeout       time.Duration = 2 * time.Second  // Socket read timeout

// Init logger subsystem
var log logger.Log

func init() {
	log.New()
}