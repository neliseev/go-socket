package socket

import (
	"github.com/neliseev/logger"
	"time"
)

var log logger.Log // logger subsystem

// Defaults
const rtimeout   time.Duration = 2 * time.Second // Socket read timeout
const udpMsgSize int           = 508             // RFC 791 (Min IP Size - Max IP Header Size - UDP Header Size)
const maxMsgSize int           = 128             // ToDo Set configurable?
