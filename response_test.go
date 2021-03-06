package gig

import (
	"testing"

	"github.com/matryer/is"
)

func TestResponse(t *testing.T) {
	conn := &FakeConn{}
	res := &Response{Writer: conn}

	is := is.New(t)

	_, err := res.Write([]byte("test"))
	is.NoErr(err)
	is.Equal("20 text/gemini\r\ntest", conn.Written)
}

func TestResponse_Write_FallsBackToDefaultStatus(t *testing.T) {
	conn := &FakeConn{}
	res := &Response{Writer: conn}

	is := is.New(t)

	_, err := res.Write([]byte("test"))
	is.NoErr(err)
	is.Equal("20 text/gemini\r\ntest", conn.Written)
}

func TestResponse_Write_Fails(t *testing.T) {
	conn := &FakeConn{FailAfter: 1}
	res := &Response{Writer: conn}

	is := is.New(t)

	_, err := res.Write([]byte("test"))
	is.True(err != nil)
	is.Equal("2", conn.Written)
}

func TestResponse_Double_WriteHeader(t *testing.T) {
	conn := &FakeConn{}
	res := &Response{Writer: conn}

	is := is.New(t)

	is.NoErr(res.WriteHeader(StatusSuccess, "text/gemini"))
	is.NoErr(res.WriteHeader(StatusGone, "oops"))
	is.Equal("20 text/gemini\r\n", conn.Written)
}

func TestResponse_Write_UsesSetResponseCode(t *testing.T) {
	conn := &FakeConn{}
	res := &Response{Writer: conn}

	is := is.New(t)

	res.Status = StatusCGIError
	_, err := res.Write([]byte("test"))
	is.NoErr(err)
	is.Equal("42 text/gemini\r\ntest", conn.Written)
}

func TestResponse_Write_FailIfFailedBefore(t *testing.T) {
	conn := &FakeConn{
		FailAfter: 4,
	}
	res := &Response{Writer: conn}

	is := is.New(t)

	n, err := res.Write([]byte("test"))
	is.True(err != nil)
	is.Equal(n, 0)
	is.Equal(res.err, err)
	is.Equal("20 t", conn.Written)

	_, err2 := res.Write([]byte("test"))
	is.True(err2 != nil)
	is.Equal(n, 0)
	is.Equal(res.err, err)
	is.Equal(err, err2)
	is.Equal("20 t", conn.Written)
}
