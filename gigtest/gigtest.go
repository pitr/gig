package gigtest

import (
	"crypto/tls"
	"errors"
	"net"
	"net/url"
	"time"

	"github.com/pitr/gig"
)

type (
	FakeAddr struct{}
	FakeConn struct {
		FailAfter int
		Written   string
	}
)

func (a *FakeAddr) Network() string { return "tcp" }
func (a *FakeAddr) String() string  { return "192.0.2.1:25" }

func (c *FakeConn) Read(b []byte) (n int, err error) { return len(b), nil }
func (c *FakeConn) Write(b []byte) (n int, err error) {
	if c.FailAfter > 0 && len(c.Written)+len(b) > c.FailAfter {
		cut := c.FailAfter - len(c.Written)
		c.Written = c.Written + string(b[:cut])
		return cut, errors.New("cannot write")
	}
	c.Written = c.Written + string(b)
	return len(b), nil
}
func (c *FakeConn) Close() error                       { return nil }
func (c *FakeConn) LocalAddr() net.Addr                { return &FakeAddr{} }
func (c *FakeConn) RemoteAddr() net.Addr               { return &FakeAddr{} }
func (c *FakeConn) SetDeadline(t time.Time) error      { return nil }
func (c *FakeConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *FakeConn) SetWriteDeadline(t time.Time) error { return nil }

func NewContext(g *gig.Gig, uri string, tlsState *tls.ConnectionState) (gig.Context, *FakeConn) {
	u, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}
	conn := &FakeConn{}
	return g.NewContext(conn, u, uri, tlsState), conn
}
