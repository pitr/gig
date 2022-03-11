package gig

import (
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/url"
	"time"
)

type (
	// FakeAddr ia a fake net.Addr implementation.
	FakeAddr struct{}
	// FakeConn ia a fake net.Conn that can record what is written and can fail
	// after FailAfter bytes were written.
	FakeConn struct {
		FailAfter int
		Written   string
		Reader    io.Reader
	}
)

// Network returns dummy data.
func (a *FakeAddr) Network() string { return "tcp" }

// String returns dummy data.
func (a *FakeAddr) String() string { return "192.0.2.1:25" }

// Read always returns success if Reader is nil, otherwise Read is delegated
// to the FakeConn reader.
func (c *FakeConn) Read(b []byte) (n int, err error) {
	if c.Reader != nil {
		return c.Reader.Read(b)
	}

	return len(b), nil
}

// Write records bytes written and fails after FailAfter bytes.
func (c *FakeConn) Write(b []byte) (n int, err error) {
	if c.FailAfter > 0 && len(c.Written)+len(b) > c.FailAfter {
		cut := c.FailAfter - len(c.Written)
		c.Written += string(b[:cut])

		return cut, errors.New("cannot write")
	}

	c.Written += string(b)

	return len(b), nil
}

// Close always returns nil.
func (c *FakeConn) Close() error { return nil }

// LocalAddr returns fake address.
func (c *FakeConn) LocalAddr() net.Addr { return &FakeAddr{} }

// RemoteAddr returns fake address.
func (c *FakeConn) RemoteAddr() net.Addr { return &FakeAddr{} }

// SetDeadline always returns nil.
func (c *FakeConn) SetDeadline(t time.Time) error { return nil }

// SetReadDeadline always returns nil.
func (c *FakeConn) SetReadDeadline(t time.Time) error { return nil }

// SetWriteDeadline always returns nil.
func (c *FakeConn) SetWriteDeadline(t time.Time) error { return nil }

// ConnectionState always returns nil.
func (c *FakeConn) ConnectionState() tls.ConnectionState { return tls.ConnectionState{} }

type FakeOpt func(Context, *FakeConn)

// WithFakeReader adds reader to the context and connection.
func WithFakeReader(r io.Reader) FakeOpt {
	return func(c Context, f *FakeConn) {
		f.Reader = r
		c.(*context).reader = r
	}
}

// NewFakeContext returns Context that writes to FakeConn.
func (g *Gig) NewFakeContext(uri string, tlsState *tls.ConnectionState, opts ...FakeOpt) (Context, *FakeConn) {
	u, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}

	conn := &FakeConn{}
	ctx := g.newContext(conn, u, uri, tlsState)

	for _, o := range opts {
		o(ctx, conn)
	}

	return ctx, conn
}
