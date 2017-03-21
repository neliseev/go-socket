package socket

import (
	"errors"
)

type Msg struct {
	Req  string
	Data []byte
}

// Unpack binary message to Msq structure
// s is are separator
func (m *Msg) Unpack(data []byte, s byte) error {
	if len(s) != 1 {
		return errors.New("Separator size oferflow, should be 1")
	}

	for i, b := range data {
		if b == []byte(s) {
			m.Req  = string(data[:i])
			m.Data = data[i + 1:]

			break
		}

		i++
	}

	if m.Req == "" {
		return errors.New("Undefined request in message")
	}

	return nil
}
