package socket

// Handler interface - implementing serve method for each new message from packet
type Handler interface {
	Serve(w Response, r *Msg)
}

// HandlerFunc type - is are adapter type for use ordinary functions as Socket handlers
type HandlerFunc func(Response, *Msg)

// Server method - implementing call f(w, r)
func (f HandlerFunc) Serve(w Response, r *Msg) {
	f(w, r)
}

// Handle func - registers handler with given pattern in default serve multiplexer
func Handle(pattern string, handler Handler) {
	DefaultServeMux.Handle(pattern, handler)
}

// HandleFunc func - registers handler functions with given pattern in default serve multiplexer
func HandleFunc(pattern string, handler func(Response, *Msg)) {
	DefaultServeMux.HandleFunc(pattern, handler)
}

// HandleRemove func - de-registers handler with given patter in default serve multiplexer
func HandleRemove(pattern string) {
	DefaultServeMux.HandleRemove(pattern)
}

// HandleFailed func - return SERVEFAIL for every requests
func HandleFailed(w Response, r *Msg) {
	m := new(Msg)
	m.Req = ""
	m.Data = []byte("SERVFAIL")

	w.WriteMsg(m)
}

// failedHandler func return Handler with HandleFailed func
func failedHandler() Handler {
	return HandlerFunc(HandleFailed)
}
