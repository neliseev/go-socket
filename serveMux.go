package socket

import (
	"sync"
)

var DefaultServeMux = NewServeMux()

type ServeMux struct {
	h map[string]Handler
	m *sync.RWMutex
}

func NewServeMux() *ServeMux {
	return &ServeMux{
		h: make(map[string]Handler),
		m: new(sync.RWMutex),
	}
}

func (mux *ServeMux) Handle(pattern string, handler Handler) {
	if pattern == "" {
		log.Crit("socket mux invalid pattern " + pattern)
	}

	mux.m.Lock()
	mux.h[pattern] = handler
	mux.m.Unlock()
}

func (mux *ServeMux) HandleFunc(pattern string, handler func(Response, *Msg)) {
	mux.Handle(pattern, HandlerFunc(handler))
}

func (mux *ServeMux) HandleRemove(pattern string) {
	if pattern == "" {
		log.Crit("socket mux invalid pattern " + pattern)
	}
	mux.m.Lock()
	delete(mux.h, pattern)
	mux.m.Unlock()
}

func (mux *ServeMux) Serve(w Response, m *Msg) {
	var h Handler

	if m.Req == "" {
		h = failedHandler()
	} else {
		if h = mux.match(m.Req); h == nil {
			h = failedHandler()
		}
	}
	h.Serve(w, m)
}

func (mux *ServeMux) match(q string) Handler {
	mux.m.RLock()
	defer mux.m.RUnlock()

	var handler Handler

	if h, ok := mux.h[q]; ok {
		return h
	}

	return handler
}
