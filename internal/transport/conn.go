package transport

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	maxBodySize = 99999 // 5-digit decimal max
	prefixLen   = 5
)

// Conn wraps a net.Conn with 5-digit decimal length-prefix framing.
// Both half-packet and sticky-packet are handled by buffering reads and
// parsing by length prefix.
type Conn struct {
	c  net.Conn
	br *bufio.Reader
	bw *bufio.Writer
}

// NewConn wraps an existing net.Conn.
func NewConn(c net.Conn) *Conn {
	return &Conn{
		c:  c,
		br: bufio.NewReaderSize(c, 64<<10),
		bw: bufio.NewWriterSize(c, 64<<10),
	}
}

// ReadFrame reads one full JSON body. It blocks until the entire body
// (per the 5-digit prefix) is available, transparently handling partial
// reads and multiple frames in one TCP segment.
func (f *Conn) ReadFrame() ([]byte, error) {
	var prefix [prefixLen]byte
	if _, err := io.ReadFull(f.br, prefix[:]); err != nil {
		return nil, err
	}
	n := 0
	for i, b := range prefix {
		if b < '0' || b > '9' {
			return nil, fmt.Errorf("invalid length prefix byte %d: %q (0x%02x)", i, b, b)
		}
		n = n*10 + int(b-'0')
	}
	if n <= 0 || n > maxBodySize {
		return nil, fmt.Errorf("frame body length out of range: %d", n)
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(f.br, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// WriteFrame writes a 5-digit decimal length prefix followed by the body.
func (f *Conn) WriteFrame(body []byte) error {
	n := len(body)
	if n > maxBodySize {
		return fmt.Errorf("frame body too large: %d > %d", n, maxBodySize)
	}
	// 5-digit decimal prefix, zero-padded.
	var prefix [prefixLen]byte
	for i := prefixLen - 1; i >= 0; i-- {
		prefix[i] = byte('0' + n%10)
		n /= 10
	}
	if _, err := f.bw.Write(prefix[:]); err != nil {
		return err
	}
	if _, err := f.bw.Write(body); err != nil {
		return err
	}
	return f.bw.Flush()
}

// SetDeadline sets read/write deadlines on the underlying connection.
func (f *Conn) SetDeadline(t time.Time) error { return f.c.SetDeadline(t) }

// Close closes the underlying connection.
func (f *Conn) Close() error { return f.c.Close() }

// RemoteAddr returns the peer address.
func (f *Conn) RemoteAddr() net.Addr { return f.c.RemoteAddr() }
