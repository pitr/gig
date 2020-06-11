// +build !race

package gig

import (
    "crypto/tls"
    "io"
    "net"
    "syscall"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
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
    ln := &errorListener{[]error{
        &net.OpError{
            Op:  "accept",
            Err: syscall.EMFILE,
        }}}
    e := New()
    e.Listener = ln
    err := e.serve()
    assert.Equal(t, io.EOF, err)
}

func TestServe(t *testing.T) {
    e := New()
    go func() {
        _ = e.StartTLS("127.0.0.1:0", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
    }()
    time.Sleep(200 * time.Millisecond)

    addr := e.Listener.Addr().String()
    conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
    require.NoError(t, err)
    _, err = conn.Write([]byte("/test\r\n"))
    require.NoError(t, err)

    buf := make([]byte, 15)
    n, err := conn.Read(buf)
    require.NoError(t, err)

    assert.Equal(t, "51 Not Found\r\n\x00", string(buf))
    assert.Equal(t, 14, n)

    e.Close()
}

func TestServe_SlowClient_Read(t *testing.T) {
    e := New()
    e.ReadTimeout = 1 * time.Millisecond

    go func() {
        _ = e.StartTLS("127.0.0.1:0", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
    }()
    time.Sleep(200 * time.Millisecond)

    addr := e.Listener.Addr().String()
    conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
    require.NoError(t, err)

    time.Sleep(200 * time.Millisecond) // client sleeps before sending request

    _, err = conn.Write([]byte("/test\r\n"))

    require.Error(t, err)

    e.Close()
}

func TestServe_SlowClient_Write(t *testing.T) {
    e := New()
    e.WriteTimeout = 1 * time.Millisecond

    go func() {
        err := e.StartTLS("127.0.0.1:0", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
        if err != ErrServerClosed { // Prevent the test to fail after closing the servers
            require.NoError(t, err)
        }
    }()
    time.Sleep(200 * time.Millisecond)

    addr := e.Listener.Addr().String()
    conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
    require.NoError(t, err)
    _, err = conn.Write([]byte("/test\r\n"))
    require.NoError(t, err)

    conn.Close() // client closes connection before reading response

    e.Close()
}

func TestServe_Overflow(t *testing.T) {
    e := New()
    go func() {
        _ = e.StartTLS("127.0.0.1:0", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
    }()
    time.Sleep(200 * time.Millisecond)

    addr := e.Listener.Addr().String()
    conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
    require.NoError(t, err)

    request := make([]byte, 2000)
    _, _ = conn.Write(request)

    buf := make([]byte, 23)
    n, err := conn.Read(buf)
    require.NoError(t, err)

    assert.Equal(t, "59 Request too long!\r\n\x00", string(buf))
    assert.Equal(t, 22, n)

    e.Close()
}

func TestServe_NotGemini(t *testing.T) {
    e := New()
    go func() {
        _ = e.StartTLS("127.0.0.1:0", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
    }()
    time.Sleep(200 * time.Millisecond)

    addr := e.Listener.Addr().String()
    conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
    require.NoError(t, err)

    _, err = conn.Write([]byte("http://google.com\r\n"))
    require.NoError(t, err)

    buf := make([]byte, 40)
    n, err := conn.Read(buf)
    require.NoError(t, err)

    assert.Equal(t, "59 No proxying to non-Gemini content!\r\n\x00", string(buf))
    assert.Equal(t, 39, n)

    e.Close()
}

func TestServe_NotURL(t *testing.T) {
    e := New()
    go func() {
        _ = e.StartTLS("127.0.0.1:0", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
    }()
    time.Sleep(200 * time.Millisecond)

    addr := e.Listener.Addr().String()
    conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
    require.NoError(t, err)

    _, err = conn.Write([]byte("::::::\r\n"))
    require.NoError(t, err)

    buf := make([]byte, 24)
    n, err := conn.Read(buf)
    require.NoError(t, err)

    assert.Equal(t, "59 Error parsing URL!\r\n\x00", string(buf))
    assert.Equal(t, 23, n)

    e.Close()
}
