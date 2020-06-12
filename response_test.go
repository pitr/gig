package gig

import (
	"testing"

	"github.com/matryer/is"
)

func TestResponse(t *testing.T) {
	conn := &fakeConn{}
	res := &Response{Writer: conn}

	is := is.New(t)

	_, err := res.Write([]byte("test"))
	is.NoErr(err)
	is.Equal("20 text/gemini\r\ntest", conn.Written)
}

func TestResponse_Write_FallsBackToDefaultStatus(t *testing.T) {
	conn := &fakeConn{}
	res := &Response{Writer: conn}

	is := is.New(t)

	_, err := res.Write([]byte("test"))
	is.NoErr(err)
	is.Equal("20 text/gemini\r\ntest", conn.Written)
}

func TestResponse_Write_Fails(t *testing.T) {
	conn := &fakeConn{failAfter: 1}
	res := &Response{Writer: conn}

	is := is.New(t)

	_, err := res.Write([]byte("test"))
	is.True(err != nil)
	is.Equal("2", conn.Written)
}

func TestResponse_Double_WriteHeader(t *testing.T) {
	conn := &fakeConn{}
	res := &Response{Writer: conn}

	is := is.New(t)

	is.NoErr(res.WriteHeader(StatusSuccess, "text/gemini"))
	is.NoErr(res.WriteHeader(StatusGone, "oops"))
	is.Equal("20 text/gemini\r\n", conn.Written)
}

func TestResponse_Write_UsesSetResponseCode(t *testing.T) {
	conn := &fakeConn{}
	res := &Response{Writer: conn}

	is := is.New(t)

	res.Status = StatusCGIError
	_, err := res.Write([]byte("test"))
	is.NoErr(err)
	is.Equal("42 text/gemini\r\ntest", conn.Written)
}
