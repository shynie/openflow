package openflow

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

type dummyAddr string

func (a dummyAddr) Network() string {
	return "dummy"
}

func (a dummyAddr) String() string {
	return string(a)
}

type dummyConn struct {
	r bytes.Buffer
	w bytes.Buffer
}

func (c *dummyConn) Read(b []byte) (int, error) {
	return c.r.Read(b)
}

func (c *dummyConn) Write(b []byte) (int, error) {
	return c.w.Write(b)
}

func (c *dummyConn) Close() error {
	return nil
}

func (c *dummyConn) LocalAddr() net.Addr {
	return dummyAddr("local-addr")
}

func (c *dummyConn) RemoteAddr() net.Addr {
	return dummyAddr("remote-addr")
}

func (c *dummyConn) SetDeadline(_ time.Time) error {
	return nil
}

func (c *dummyConn) SetReadDeadline(_ time.Time) error {
	return nil
}

func (c *dummyConn) SetWriteDeadline(_ time.Time) error {
	return nil
}

type dummyListener struct {
	conn net.Conn
}

func (l *dummyListener) Accept() (c net.Conn, e error) {
	c, l.conn = l.conn, nil
	if c == nil {
		e = io.EOF
	}

	return
}

func (l *dummyListener) Close() error {
	if l.conn != nil {
		return l.conn.Close()
	}

	return nil
}

func (l *dummyListener) Addr() net.Addr {
	return dummyAddr("dummy-address")
}

func TestServerMux(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	mux := NewServeMux()
	mux.HandleFunc(T_HELLO, func(rw ResponseWriter, r *Request) {
		rw.Header().Set(VersionHeaderKey, uint8(4))
		rw.Write([]byte{0, 0, 0, 0})
		wg.Done()
	})

	reader := bytes.NewBuffer([]byte{4, 0, 0, 8, 0, 0, 0, 0})
	conn := &dummyConn{r: *reader}

	s := Server{Addr: "0.0.0.0:6633", Handler: mux}
	err := s.Serve(&dummyListener{conn})

	if err != io.EOF {
		t.Fatal("Serve failed:", err)
	}

	wg.Wait()

	returned := fmt.Sprintf("%x", conn.w.Bytes())
	if returned != "0400000c0000000000000000" {
		t.Fatal("Invalid data returned: ", returned)
	}
}
