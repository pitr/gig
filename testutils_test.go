package gig

import (
	"errors"
	"net"
	"net/url"
	"time"
)

type (
	fakeAddr struct{}
	fakeConn struct {
		failAfter int
		Written   string
	}
)

func (a *fakeAddr) Network() string { return "tcp" }
func (a *fakeAddr) String() string  { return "192.0.2.1:25" }

func (c *fakeConn) Read(b []byte) (n int, err error) { return len(b), nil }
func (c *fakeConn) Write(b []byte) (n int, err error) {
	if c.failAfter > 0 && len(c.Written)+len(b) > c.failAfter {
		c.Written = c.Written + string(b[:c.failAfter-len(c.Written)])
		return 0, errors.New("cannot write")
	}
	c.Written = c.Written + string(b)
	return len(b), nil
}
func (c *fakeConn) Close() error                       { return nil }
func (c *fakeConn) LocalAddr() net.Addr                { return &fakeAddr{} }
func (c *fakeConn) RemoteAddr() net.Addr               { return &fakeAddr{} }
func (c *fakeConn) SetDeadline(t time.Time) error      { return nil }
func (c *fakeConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *fakeConn) SetWriteDeadline(t time.Time) error { return nil }

func newContext(uri string) Context {
	u, _ := url.Parse(uri)
	e := New()
	maxParam := 20
	e.maxParam = &maxParam
	return e.NewContext(&fakeConn{}, u, uri, nil)
}
