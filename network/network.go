package network

import (
	"net"
	"strings"
)

// IsErrClosedConn checks if the error is about closed connection
// Ugly, but probably no better way to detect this
// TODO
func IsErrClosedConn(err error) bool {
	return strings.Contains(err.Error(), "use of closed network connection")
}

// IsErrTimeout is true if the error is timeout
func IsErrTimeout(err error) bool {
	e, ok := err.(net.Error)
	return ok && e.Timeout()
}
