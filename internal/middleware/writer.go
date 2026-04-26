package middleware

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(status int) {
	sw.status = status
	sw.ResponseWriter.WriteHeader(status)
}

func (sw *statusWriter) Write(b []byte) (int, error) {
	if sw.status == 0 {
		sw.status = http.StatusOK
	}
	return sw.ResponseWriter.Write(b)
}

func (sw *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := sw.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, errors.New("ResponseWriter does not implement Hijacker")
}
