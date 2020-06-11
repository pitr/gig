package gig

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	user struct {
		ID   int    `json:"id" xml:"id" form:"id" query:"id" param:"id"`
		Name string `json:"name" xml:"name" form:"name" query:"name" param:"name"`
	}
)

const (
	userJSON = `{"id":1,"name":"Jon Snow"}`
	userXML  = `<user><id>1</id><name>Jon Snow</name></user>`
)

const userJSONPretty = `{
  "id": 1,
  "name": "Jon Snow"
}`

const userXMLPretty = `<user>
  <id>1</id>
  <name>Jon Snow</name>
</user>`

func TestGig(t *testing.T) {
	g := New()
	c := newContext("/").(*context)

	// Router
	assert.NotNil(t, g.Router())

	// DefaultGeminiErrorHandler
	g.DefaultGeminiErrorHandler(errors.New("error"), c)
	assert.Equal(t, "50 error\r\n", c.conn.(*fakeConn).Written)
}

func TestGigStatic(t *testing.T) {
	g := New()

	assert := assert.New(t)

	// OK
	g.Static("/images", "_fixture/images")
	b := request("/images/walle.png", g)
	assert.Equal(true, strings.HasPrefix(b, "20 image/png\r\n"))

	// No file
	g.Static("/images", "_fixture/scripts")
	b = request("/images/bolt.png", g)
	assert.Equal("51 Not Found\r\n", b)

	// Directory
	g.Static("/images", "_fixture/images")
	b = request("/images", g)
	assert.Equal("51 Not Found\r\n", b)

	// Directory with index.gmi
	g.Static("/", "_fixture")
	b = request("/", g)
	assert.Equal("20 text/gemini\r\n# Hello from gig\n\n=> / 🏠 Home\n", b)

	// Sub-directory with index.gmi
	b = request("/folder", g)
	assert.Equal("20 text/gemini\r\n# Listing _fixture/folder\n\n=> /*/about.gmi about.gmi [ 29B ]\n=> /*/another.blah another.blah [ 14B ]\n", b)

	// File without known mime
	b = request("/folder/another.blah", g)
	assert.Equal("20 octet/stream\r\n# Another page", b)

	// Escape
	g.Static("/escape", "")
	b = request("/escape/../../", g)
	assert.Equal(true, strings.Contains(b, "/escape/*/gig.go"))
}

func TestGigFile(t *testing.T) {
	g := New()
	g.File("/walle", "_fixture/images/walle.png")
	b := request("/walle", g)
	assert.Equal(t, true, strings.HasPrefix(b, "20 "))
}

func TestGigMiddleware(t *testing.T) {
	g := New()
	buf := new(bytes.Buffer)

	g.Pre(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			assert.Empty(t, c.Path())
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
	assert.Equal(t, "-1123", buf.String())
	assert.Equal(t, "20 text/plain; charset=UTF-8\r\nOK", b)
}

func TestGigMiddlewareError(t *testing.T) {
	g := New()
	g.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return NewGeminiErrorFrom(ErrPermanentFailure, "oops")
		}
	})
	g.Handle("/", NotFoundHandler)
	b := request("/", g)
	assert.Equal(t, "50 oops\r\n", b)
}

func TestGigHandler(t *testing.T) {
	g := New()

	// HandlerFunc
	g.Handle("/ok", func(c Context) error {
		return c.Text(StatusSuccess, "OK")
	})

	b := request("/ok", g)
	assert.Equal(t, "20 text/plain; charset=UTF-8\r\nOK", b)
}

func TestGigHandle(t *testing.T) {
	g := New()
	g.Handle("/", func(c Context) error {
		return c.Text(StatusSuccess, "hello")
	})
	b := request("/", g)
	assert.Equal(t, "20 text/plain; charset=UTF-8\r\nhello", b)
}

func TestGigURL(t *testing.T) {
	g := New()
	static := func(Context) error { return nil }
	getUser := func(Context) error { return nil }
	getFile := func(Context) error { return nil }

	g.Handle("/static/file", static)
	g.Handle("/users/:id", getUser)
	gr := g.Group("/group")
	gr.Handle("/users/:uid/files/:fid", getFile)

	assert := assert.New(t)

	assert.Equal("/static/file", g.URL(static))
	assert.Equal("/users/:id", g.URL(getUser))
	assert.Equal("/users/1", g.URL(getUser, "1"))
	assert.Equal("/group/users/1/files/:fid", g.URL(getFile, "1"))
	assert.Equal("/group/users/1/files/1", g.URL(getFile, "1", "1"))
}

func TestGigRoutes(t *testing.T) {
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

	if assert.Equal(t, len(routes), len(g.Routes())) {
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
}

func TestGigEncodedPath(t *testing.T) {
	g := New()
	g.Handle("/:id", func(c Context) error {
		return c.NoContentSuccess()
	})
	c := newContext("/with%2Fslash")
	g.ServeGemini(c)
	assert.Equal(t, "20 text/gemini\r\n", c.(*context).conn.(*fakeConn).Written)
}

func TestGigGroup(t *testing.T) {
	g := New()
	buf := new(bytes.Buffer)
	g.Use(MiddlewareFunc(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("0")
			return next(c)
		}
	}))
	h := func(c Context) error {
		return c.NoContentSuccess()
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
	assert.Equal(t, "0", buf.String())

	buf.Reset()
	request("/group1", g)
	assert.Equal(t, "01", buf.String())

	buf.Reset()
	request("/group2/group3", g)
	assert.Equal(t, "023", buf.String())
}

func TestGigNotFound(t *testing.T) {
	g := New()
	c := newContext("/files").(*context)
	g.ServeGemini(c)
	assert.Equal(t, "51 Not Found\r\n", c.conn.(*fakeConn).Written)
}

func TestGigContext(t *testing.T) {
	g := New()
	c := g.AcquireContext()
	assert.IsType(t, new(context), c)
	g.ReleaseContext(c)
}

func TestGigStartTLS(t *testing.T) {
	g := New()
	go func() {
		_ = g.StartTLS(":0", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
	}()
	time.Sleep(200 * time.Millisecond)

	g.Close()
}

func TestGigStartTLS_BadAddress(t *testing.T) {
	g := New()
	err := g.StartTLS("garbage address", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
	require.Error(t, err)
	require.Contains(t, err.Error(), "address garbage address: missing port in address")
}

func TestGigStartTLSByteString(t *testing.T) {
	cert, err := ioutil.ReadFile("_fixture/certs/cert.pem")
	require.NoError(t, err)
	key, err := ioutil.ReadFile("_fixture/certs/key.pem")
	require.NoError(t, err)

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
			g := New()
			g.HideBanner = true

			go func() {
				err := g.StartTLS(":0", test.cert, test.key)
				if test.expectedErr != nil {
					require.EqualError(t, err, test.expectedErr.Error())
				} else if err != ErrServerClosed { // Prevent the test to fail after closing the servers
					require.NoError(t, err)
				}
			}()
			time.Sleep(200 * time.Millisecond)

			g.Close()
		})
	}
}

func TestGigStartAutoTLS(t *testing.T) {
	g := New()
	errChan := make(chan error)

	go func() {
		errChan <- g.StartAutoTLS(":0")
	}()
	time.Sleep(200 * time.Millisecond)

	select {
	case err := <-errChan:
		assert.NoError(t, err)
	default:
		assert.NoError(t, g.Close())
	}
}

func request(path string, g *Gig) string {
	c := newContext(path).(*context)
	g.ServeGemini(c)
	return c.conn.(*fakeConn).Written
}

func TestGeminiError(t *testing.T) {
	t.Run("manual", func(t *testing.T) {
		err := NewGeminiError(StatusSlowDown, "oops")
		assert.Equal(t, "code=44, message=oops", err.Error())

	})
	t.Run("existing", func(t *testing.T) {
		err := ErrSlowDown
		assert.Equal(t, "code=44, message=Slow Down", err.Error())
	})
	t.Run("inherited", func(t *testing.T) {
		err := NewGeminiErrorFrom(ErrSlowDown, "oops")
		assert.Equal(t, "code=44, message=oops", err.Error())
	})
}

func TestGigClose(t *testing.T) {
	g := New()
	errCh := make(chan error)

	go func() {
		errCh <- g.StartTLS(":0", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
	}()

	time.Sleep(200 * time.Millisecond)

	if err := g.Close(); err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, g.Close().Error(), "use of closed network connection")

	err := <-errCh
	assert.Equal(t, err.Error(), "gemini: Server closed")
}
