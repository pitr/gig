package gig

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestGig(t *testing.T) {
	is := is.New(t)

	g := New()
	c := newContext("/").(*context)

	// Router
	is.True(g.router != nil)

	// DefaultGeminiErrorHandler
	DefaultGeminiErrorHandler(errors.New("error"), c)
	is.Equal("50 error\r\n", c.conn.(*fakeConn).Written)
}

func TestGigStatic(t *testing.T) {
	is := is.New(t)

	g := New()

	// OK
	g.Static("/images", "_fixture/images")
	b := request("/images/walle.png", g)
	is.True(strings.HasPrefix(b, "20 image/png\r\n"))

	// No file
	g.Static("/images", "_fixture/scripts")
	b = request("/images/bolt.png", g)
	is.Equal("51 Not Found\r\n", b)

	// Directory
	g.Static("/images", "_fixture/images")
	b = request("/images", g)
	is.Equal("51 Not Found\r\n", b)

	// Directory with index.gmi
	g.Static("/", "_fixture")
	b = request("/", g)
	is.Equal("20 text/gemini\r\n# Hello from gig\n\n=> / 🏠 Home\n", b)

	// Sub-directory with index.gmi
	b = request("/folder", g)
	is.Equal("20 text/gemini\r\n# Listing _fixture/folder\n\n=> /*/about.gmi about.gmi [ 29B ]\n=> /*/another.blah another.blah [ 14B ]\n", b)

	// File without known mime
	b = request("/folder/another.blah", g)
	is.Equal("20 octet/stream\r\n# Another page", b)

	// Escape
	g.Static("/escape", "")
	b = request("/escape/../../", g)
	is.True(strings.Contains(b, "/escape/*/gig.go"))
}

func TestGigFile(t *testing.T) {
	is := is.New(t)

	g := New()
	g.File("/walle", "_fixture/images/walle.png")
	b := request("/walle", g)
	is.True(strings.HasPrefix(b, "20 "))
}

func TestGigMiddleware(t *testing.T) {
	is := is.New(t)

	g := New()
	buf := new(bytes.Buffer)

	g.Pre(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			is.True(c.Path() == "")
			buf.WriteString("-1")
			return next(c)
		}
	})

	g.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("1")
			return next(c)
		}
	})

	g.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("2")
			return next(c)
		}
	})

	g.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("3")
			return next(c)
		}
	})

	// Route
	g.Handle("/", func(c Context) error {
		return c.Text(StatusSuccess, "OK")
	})

	b := request("/", g)

	is.Equal("-1123", buf.String())
	is.Equal("20 text/plain; charset=UTF-8\r\nOK", b)
}

func TestGigMiddlewareError(t *testing.T) {
	is := is.New(t)

	g := New()
	g.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return NewErrorFrom(ErrPermanentFailure, "oops")
		}
	})
	g.Handle("/", NotFoundHandler)
	b := request("/", g)
	is.Equal("50 oops\r\n", b)
}

func TestGigHandler(t *testing.T) {
	is := is.New(t)

	g := New()

	// HandlerFunc
	g.Handle("/ok", func(c Context) error {
		return c.Text(StatusSuccess, "OK")
	})

	b := request("/ok", g)
	is.Equal("20 text/plain; charset=UTF-8\r\nOK", b)
}

func TestGigHandle(t *testing.T) {
	is := is.New(t)

	g := New()
	g.Handle("/", func(c Context) error {
		return c.Text(StatusSuccess, "hello")
	})

	b := request("/", g)
	is.Equal("20 text/plain; charset=UTF-8\r\nhello", b)
}

func TestGigURL(t *testing.T) {
	is := is.New(t)

	g := New()
	static := func(Context) error { return nil }
	getUser := func(Context) error { return nil }
	getFile := func(Context) error { return nil }

	g.Handle("/static/file", static)
	g.Handle("/users/:id", getUser)
	gr := g.Group("/group")
	gr.Handle("/users/:uid/files/:fid", getFile)

	is.Equal("/static/file", g.URL(static))
	is.Equal("/users/:id", g.URL(getUser))
	is.Equal("/users/1", g.URL(getUser, "1"))
	is.Equal("/group/users/1/files/:fid", g.URL(getFile, "1"))
	is.Equal("/group/users/1/files/1", g.URL(getFile, "1", "1"))
}

func TestGigRoutes(t *testing.T) {
	is := is.New(t)

	g := New()
	routes := []*Route{
		{"/users/:user/events", ""},
		{"/users/:user/events/public", ""},
		{"/repos/:owner/:repo/git/refs", ""},
		{"/repos/:owner/:repo/git/tags", ""},
	}

	for _, r := range routes {
		g.Handle(r.Path, func(c Context) error {
			return c.Text(StatusSuccess, "OK")
		})
	}

	is.Equal(len(routes), len(g.Routes()))

	for _, r := range g.Routes() {
		found := false

		for _, rr := range routes {
			if r.Path == rr.Path {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Route %s not found", r.Path)
		}
	}
}

func TestGigEncodedPath(t *testing.T) {
	is := is.New(t)

	g := New()
	g.Handle("/:id", func(c Context) error {
		return c.NoContent(StatusInput, "please enter name")
	})

	c := newContext("/with%2Fslash")
	g.ServeGemini(c)
	is.Equal("10 please enter name\r\n", c.(*context).conn.(*fakeConn).Written)
}

func TestGigGroup(t *testing.T) {
	is := is.New(t)

	g := New()
	buf := new(bytes.Buffer)

	g.Use(MiddlewareFunc(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("0")
			return next(c)
		}
	}))

	h := func(c Context) error {
		return c.NoContent(StatusInput, "please enter name")
	}

	//--------
	// Routes
	//--------

	g.Handle("/users", h)

	// Group
	g1 := g.Group("/group1")
	g1.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("1")
			return next(c)
		}
	})
	g1.Handle("", h)

	// Nested groups with middleware
	g2 := g.Group("/group2")
	g2.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("2")
			return next(c)
		}
	})

	g3 := g2.Group("/group3")
	g3.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("3")
			return next(c)
		}
	})
	g3.Handle("", h)

	request("/users", g)
	is.Equal("0", buf.String())

	buf.Reset()
	request("/group1", g)
	is.Equal("01", buf.String())

	buf.Reset()
	request("/group2/group3", g)
	is.Equal("023", buf.String())
}

func TestGigNotFound(t *testing.T) {
	is := is.New(t)

	g := New()
	c := newContext("/files").(*context)
	g.ServeGemini(c)
	is.Equal("51 Not Found\r\n", c.conn.(*fakeConn).Written)
}

func TestGigRun(t *testing.T) {
	g := New()

	go func() {
		_ = g.Run("127.0.0.1:0", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
	}()
	time.Sleep(200 * time.Millisecond)

	g.Close()
}

func TestGigRun_BadAddress(t *testing.T) {
	is := is.New(t)

	g := New()
	err := g.Run("garbage address", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
	is.True(err != nil)
	is.True(strings.Contains(err.Error(), "address garbage address: missing port in address"))
}

func TestGigRunByteString(t *testing.T) {
	is := is.New(t)

	cert, err := ioutil.ReadFile("_fixture/certs/cert.pem")
	is.NoErr(err)
	key, err := ioutil.ReadFile("_fixture/certs/key.pem")
	is.NoErr(err)

	switchedCertError := errors.New("tls: failed to find certificate PEM data in certificate input, but did find a private key; PEM inputs may have been switched")

	testCases := []struct {
		cert        interface{}
		key         interface{}
		expectedErr error
		name        string
	}{
		{
			cert:        "_fixture/certs/cert.pem",
			key:         "_fixture/certs/key.pem",
			expectedErr: nil,
			name:        `ValidCertAndKeyFilePath`,
		},
		{
			cert:        cert,
			key:         key,
			expectedErr: nil,
			name:        `ValidCertAndKeyByteString`,
		},
		{
			cert:        cert,
			key:         1,
			expectedErr: ErrInvalidCertOrKeyType,
			name:        `InvalidKeyType`,
		},
		{
			cert:        0,
			key:         key,
			expectedErr: ErrInvalidCertOrKeyType,
			name:        `InvalidCertType`,
		},
		{
			cert:        0,
			key:         1,
			expectedErr: ErrInvalidCertOrKeyType,
			name:        `InvalidCertAndKeyTypes`,
		},
		{
			cert:        "_fixture/certs/key.pem",
			key:         "_fixture/certs/cert.pem",
			expectedErr: switchedCertError,
			name:        `BadCertAndKey`,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			is := is.New(t)

			g := New()
			g.HideBanner = true

			go func() {
				err := g.Run("127.0.0.1:0", test.cert, test.key)
				if test.expectedErr != nil {
					is.Equal(err.Error(), test.expectedErr.Error())
				} else if err != ErrServerClosed { // Prevent the test to fail after closing the servers
					is.NoErr(err)
				}
			}()
			time.Sleep(200 * time.Millisecond)

			g.Close()
		})
	}
}

func request(path string, g *Gig) string {
	c := newContext(path).(*context)
	g.ServeGemini(c)

	return c.conn.(*fakeConn).Written
}

func TestGeminiError(t *testing.T) {
	is := is.New(t)

	t.Run("manual", func(t *testing.T) {
		err := NewError(StatusSlowDown, "oops")
		is.Equal("error=oops", err.Error())
	})
	t.Run("existing", func(t *testing.T) {
		err := ErrSlowDown
		is.Equal("error=Slow Down", err.Error())
	})
	t.Run("inherited", func(t *testing.T) {
		err := NewErrorFrom(ErrSlowDown, "oops")
		is.Equal("error=oops", err.Error())
	})
}

func TestGigClose(t *testing.T) {
	is := is.New(t)

	g := New()
	errCh := make(chan error)

	go func() {
		errCh <- g.Run("127.0.0.1:0", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
	}()

	time.Sleep(200 * time.Millisecond)

	if err := g.Close(); err != nil {
		t.Fatal(err)
	}

	is.True(strings.Contains(g.Close().Error(), "use of closed network connection"))

	err := <-errCh
	is.Equal(err.Error(), "gemini: Server closed")
}
