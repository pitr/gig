// +build !race

package gig

import (
    "crypto/tls"
    "io"
    "net"
    "syscall"
    "testing"
    "time"

    "github.com/matryer/is"
)

type errorListener struct {
    errs []error
}

func (l *errorListener) Accept() (c net.Conn, err error) {
    if len(l.errs) == 0 {
        return nil, io.EOF
    }
    err = l.errs[0]
    l.errs = l.errs[1:]
    return
}

func (l *errorListener) Close() error {
    return nil
}

func (l *errorListener) Addr() net.Addr {
    return &fakeAddr{}
}

func TestServe_NetError(t *testing.T) {
    is := is.New(t)

    ln := &errorListener{[]error{
        &net.OpError{
            Op:  "accept",
            Err: syscall.EMFILE,
        }}}
    g := New()
    g.Listener = ln
    err := g.serve()
    is.Equal(io.EOF, err)
}

func TestServe(t *testing.T) {
    is := is.New(t)

    g := New()
    go func() {
        _ = g.StartTLS("127.0.0.1:0", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
    }()
    time.Sleep(200 * time.Millisecond)

    addr := g.Listener.Addr().String()
    conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
    is.NoErr(err)
    _, err = conn.Write([]byte("/test\r\n"))
    is.NoErr(err)

    buf := make([]byte, 15)
    n, err := conn.Read(buf)
    is.NoErr(err)

    is.Equal("51 Not Found\r\n\x00", string(buf))
    is.Equal(14, n)

    g.Close()
}

func TestServe_SlowClient_Read(t *testing.T) {
    is := is.New(t)

    g := New()
    g.ReadTimeout = 1 * time.Millisecond

    go func() {
        _ = g.StartTLS("127.0.0.1:0", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
    }()
    time.Sleep(200 * time.Millisecond)

    addr := g.Listener.Addr().String()
    conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
    is.NoErr(err)

    time.Sleep(200 * time.Millisecond) // client sleeps before sending request

    _, err = conn.Write([]byte("/test\r\n"))

    is.True(err != nil)

    g.Close()
}

func TestServe_SlowClient_Write(t *testing.T) {
    is := is.New(t)

    g := New()
    g.WriteTimeout = 1 * time.Millisecond

    go func() {
        err := g.StartTLS("127.0.0.1:0", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
        if err != ErrServerClosed { // Prevent the test to fail after closing the servers
            is.NoErr(err)
        }
    }()
    time.Sleep(200 * time.Millisecond)

    addr := g.Listener.Addr().String()
    conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
    is.NoErr(err)
    _, err = conn.Write([]byte("/test\r\n"))
    is.NoErr(err)

    conn.Close() // client closes connection before reading response

    g.Close()
}

func TestServe_Overflow(t *testing.T) {
    is := is.New(t)

    g := New()
    go func() {
        _ = g.StartTLS("127.0.0.1:0", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
    }()
    time.Sleep(200 * time.Millisecond)

    addr := g.Listener.Addr().String()
    conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
    is.NoErr(err)

    request := make([]byte, 2000)
    _, _ = conn.Write(request)

    buf := make([]byte, 23)
    n, err := conn.Read(buf)
    is.NoErr(err)

    is.Equal("59 Request too long!\r\n\x00", string(buf))
    is.Equal(22, n)

    g.Close()
}

func TestServe_NotGemini(t *testing.T) {
    is := is.New(t)

    g := New()
    go func() {
        _ = g.StartTLS("127.0.0.1:0", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
    }()
    time.Sleep(200 * time.Millisecond)

    addr := g.Listener.Addr().String()
    conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
    is.NoErr(err)

    _, err = conn.Write([]byte("http://google.com\r\n"))
    is.NoErr(err)

    buf := make([]byte, 40)
    n, err := conn.Read(buf)
    is.NoErr(err)

    is.Equal("59 No proxying to non-Gemini content!\r\n\x00", string(buf))
    is.Equal(39, n)

    g.Close()
}

func TestServe_NotURL(t *testing.T) {
    is := is.New(t)

    g := New()
    go func() {
        _ = g.StartTLS("127.0.0.1:0", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
    }()
    time.Sleep(200 * time.Millisecond)

    addr := g.Listener.Addr().String()
    conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
    is.NoErr(err)

    _, err = conn.Write([]byte("::::::\r\n"))
    is.NoErr(err)

    buf := make([]byte, 24)
    n, err := conn.Read(buf)
    is.NoErr(err)

    is.Equal("59 Error parsing URL!\r\n\x00", string(buf))
    is.Equal(23, n)

    g.Close()
}
