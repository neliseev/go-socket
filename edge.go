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

// Config type is used by core/cfg to construct config from json
type Config struct {
	ListenUDP string `json:"listenUDP"`  // ip:port for UDP
	ListenTCP string `json:"listenTCP"`  // ip:port for TCP
}