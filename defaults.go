package socket

import (
	"time"
	"github.com/neliseev/logger"
)

// Defaults
const maxTCPQueries  int           = 256
const tcpIdleTimeout time.Duration = 60 * time.Second
const rtimeout       time.Duration = 2 * time.Second // Socket read timeout
const msgSep         byte          = byte(':')
const udpMsgSize     int           = 508             // RFC 791 (Min IP Size - Max IP Header Size - UDP Header Size)
const maxMsgSize     int           = 128             // ToDo Set configurable?

// Init logger subsystem
var log logger.Log

func init() {
	log.New()
}