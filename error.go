package socket

type Error string

func (e Error) Error() string { return string(e) }

// Errs codes
const errHeaderLen = Error("Zero length in header")
const errPktFlag = Error("Can't read in packet size flag")
const errPktLen = Error("Zero length in packet size flag")
const errDataRead = Error("Can't read data")
const errMsgLarge = Error("Message too large")
