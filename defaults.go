package socket

import (
	"time"
	"github.com/neliseev/logger"
)

// Defaults
const Rtimeout   time.Duration = 2 * time.Second // Socket read timeout
const UdpMsgSize int           = 508             // RFC 791 (Min IP Size - Max IP Header Size - UDP Header Size)
const MaxMsgSize int           = 128             // ToDo Set configurable?

// Init logger subsystem
var log logger.Log

func init() {
	log.New()
}