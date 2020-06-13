package gigtest

import (
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/pitr/gig"
)

func TestNewContext(t *testing.T) {
	is := is.New(t)
	g := gig.New()

	c, conn := NewContext(g, "/login", nil)

	is.NoErr(c.Response().WriteHeader(gig.StatusGone, "oops"))
	is.Equal("52 oops\r\n", conn.Written)

	n, err := conn.Read(make([]byte, 1))
	is.Equal(1, n)
	is.NoErr(err)

	n, err = conn.Write([]byte("test"))
	is.Equal(4, n)
	is.NoErr(err)

	is.Equal(nil, conn.Close())
	is.Equal(conn.LocalAddr().String(), "192.0.2.1:25")
	is.Equal(conn.RemoteAddr().String(), "192.0.2.1:25")
	is.Equal(nil, conn.SetDeadline(time.Now()))
	is.Equal(nil, conn.SetReadDeadline(time.Now()))
	is.Equal(nil, conn.SetWriteDeadline(time.Now()))
}

func TestFakeAddr(t *testing.T) {
	is := is.New(t)
	addr := &FakeAddr{}

	is.Equal("tcp", addr.Network())
	is.Equal("192.0.2.1:25", addr.String())
}

func TestFakeConn(t *testing.T) {
	is := is.New(t)
	conn := &FakeConn{FailAfter: 5}

	n, err := conn.Write([]byte("test"))
	is.Equal(4, n)
	is.NoErr(err)

	n, err = conn.Write([]byte("more"))
	is.Equal(1, n)

	if err == nil {
		is.Fail()
	}
}
