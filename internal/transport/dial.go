package transport

import (
	"fmt"
	"net"
	"time"
)

// Dial connects to the server and returns a framed Conn.
func Dial(host, port string) (*Conn, error) {
	addr := net.JoinHostPort(host, port)
	c, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}
	return NewConn(c), nil
}
