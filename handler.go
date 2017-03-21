package socket

type Handler interface {
	Serve(w Response, r *Msg)
}

type HandlerFunc func(Response, *Msg)

func (f HandlerFunc) Serve(w Response, r *Msg) {
	f(w, r)
}

func Handle(pattern string, handler Handler) {
	DefaultServeMux.Handle(pattern, handler)
}

func HandleRemove(pattern string) {
	DefaultServeMux.HandleRemove(pattern)
}

func HandleFunc(pattern string, handler func(Response, *Msg)) {
	DefaultServeMux.HandleFunc(pattern, handler)
}

func HandleFailed(w Response, r *Msg) {
	m := new(Msg)
	// ToDo
	//m.SetRcode(r, RcodeServerFailure)
	// does not matter if this write fails
	w.WriteMsg(m)
}

func failedHandler() Handler {
	return HandlerFunc(HandleFailed)
}
