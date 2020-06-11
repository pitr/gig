package gigtest

import (
	"testing"
	"time"

	"github.com/pitr/gig"
	"github.com/stretchr/testify/assert"
)

func TestNewContext(t *testing.T) {
	assert := assert.New(t)
	e := gig.New()

	c, conn := NewContext(e, "/login", nil)

	assert.NoError(c.Response().WriteHeader(gig.StatusGone, "oops"))
	assert.Equal("52 oops\r\n", conn.Written)

	n, err := conn.Read(make([]byte, 1))
	assert.Equal(1, n)
	assert.Nil(err)

	n, err = conn.Write([]byte("test"))
	assert.Equal(4, n)
	assert.Nil(err)

	assert.Equal(nil, conn.Close())
	assert.IsType(&FakeAddr{}, conn.LocalAddr())
	assert.IsType(&FakeAddr{}, conn.RemoteAddr())
	assert.Equal(nil, conn.SetDeadline(time.Now()))
	assert.Equal(nil, conn.SetReadDeadline(time.Now()))
	assert.Equal(nil, conn.SetWriteDeadline(time.Now()))
}

func TestFakeAddr(t *testing.T) {
	assert := assert.New(t)
	addr := &FakeAddr{}

	assert.Equal("tcp", addr.Network())
	assert.Equal("192.0.2.1:25", addr.String())
}

func TestFakeConn(t *testing.T) {
	assert := assert.New(t)
	conn := &FakeConn{FailAfter: 5}

	n, err := conn.Write([]byte("test"))
	assert.Equal(4, n)
	assert.Nil(err)

	n, err = conn.Write([]byte("more"))
	assert.Equal(1, n)
	assert.NotNil(err)
}
