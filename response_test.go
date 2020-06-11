package gig

import (
	"testing"

	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
)

func TestResponse(t *testing.T) {
	conn := &fakeConn{}
	res := &Response{Writer: conn}

	_, err := res.Write([]byte("test"))
	assert.NoError(t, err)
	assert.Equal(t, "20 text/gemini\r\ntest", conn.Written)
}

func TestResponse_Write_FallsBackToDefaultStatus(t *testing.T) {
	conn := &fakeConn{}
	res := &Response{Writer: conn}

	_, err := res.Write([]byte("test"))
	assert.NoError(t, err)
	assert.Equal(t, "20 text/gemini\r\ntest", conn.Written)
}

func TestResponse_Write_Fails(t *testing.T) {
	conn := &fakeConn{failAfter: 1}
	res := &Response{Writer: conn}

	_, err := res.Write([]byte("test"))
	assert.Error(t, err)
	assert.Equal(t, "2", conn.Written)
}

func TestResponse_Double_WriteHeader(t *testing.T) {
	conn := &fakeConn{}
	res := &Response{Writer: conn, logger: log.New("-")}

	assert.NoError(t, res.WriteHeader(StatusSuccess, "text/gemini"))
	assert.NoError(t, res.WriteHeader(StatusGone, "oops"))
	assert.Equal(t, "20 text/gemini\r\n", conn.Written)
}

func TestResponse_Write_UsesSetResponseCode(t *testing.T) {
	conn := &fakeConn{}
	res := &Response{Writer: conn}

	res.Status = StatusCGIError
	_, err := res.Write([]byte("test"))
	assert.NoError(t, err)
	assert.Equal(t, "42 text/gemini\r\ntest", conn.Written)
}
