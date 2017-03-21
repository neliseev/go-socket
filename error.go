package socket

type Error string

func (e Error) Error() string { return string(e) }

// Errs codes
const errFlagSize = Error("Can't read in packet size flag")
const errFlagLen  = Error("Zero lenght in packet size flag")
const errDataRead = Error("Can't read data")
const errMsgLarge = Error("Message too large")
