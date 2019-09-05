package web

import (
	"net/http"
)

// LogStatusReponseWriter wraps http.ResponseWriter and logs status code.
// Logged status code can be retrieved later by calling Status() method.
type LogStatusReponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

// NewLogStatusReponseWriter creates an instance of LogStatusReponseWriter
func NewLogStatusReponseWriter(w http.ResponseWriter) *LogStatusReponseWriter {
	return &LogStatusReponseWriter{ResponseWriter: w, status: http.StatusOK}
}

// Status returns logged http status code
func (w *LogStatusReponseWriter) Status() int {
	return w.status
}

// Write implements http.ResponseWriter interface
func (w *LogStatusReponseWriter) Write(p []byte) (n int, err error) {
	if !w.wroteHeader {
		w.WriteHeader(w.status)
	}
	return w.ResponseWriter.Write(p)
}

// WriteHeader implements http.ResponseWriter interface
func (w *LogStatusReponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	// Check after in case there's error handling in the wrapped ResponseWriter.
	if w.wroteHeader {
		return
	}
	w.status = code
	w.wroteHeader = true
}

// Push implements http.Pusher interface
func (w *LogStatusReponseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := w.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}
