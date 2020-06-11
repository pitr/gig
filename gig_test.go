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
	e := New()
	c := newContext("/").(*context)

	// Router
	assert.NotNil(t, e.Router())

	// DefaultGeminiErrorHandler
	e.DefaultGeminiErrorHandler(errors.New("error"), c)
	assert.Equal(t, "50 error\r\n", c.conn.(*fakeConn).Written)
}

func TestGigStatic(t *testing.T) {
	e := New()

	assert := assert.New(t)

	// OK
	e.Static("/images", "_fixture/images")
	b := request("/images/walle.png", e)
	assert.Equal(true, strings.HasPrefix(b, "20 image/png\r\n"))

	// No file
	e.Static("/images", "_fixture/scripts")
	b = request("/images/bolt.png", e)
	assert.Equal("51 Not Found\r\n", b)

	// Directory
	e.Static("/images", "_fixture/images")
	b = request("/images", e)
	assert.Equal("51 Not Found\r\n", b)

	// Directory with index.gmi
	e.Static("/", "_fixture")
	b = request("/", e)
	assert.Equal("20 text/gemini\r\n# Hello from gig\n\n=> / 🏠 Home\n", b)

	// Sub-directory with index.gmi
	b = request("/folder", e)
	assert.Equal("20 text/gemini\r\n# Listing _fixture/folder\n\n=> /*/about.gmi about.gmi [ 29B ]\n=> /*/another.blah another.blah [ 14B ]\n", b)

	// File without known mime
	b = request("/folder/another.blah", e)
	assert.Equal("20 octet/stream\r\n# Another page", b)

	// Escape
	e.Static("/escape", "")
	b = request("/escape/../../", e)
	assert.Equal(true, strings.Contains(b, "/escape/*/gig.go"))
}

func TestGigFile(t *testing.T) {
	e := New()
	e.File("/walle", "_fixture/images/walle.png")
	b := request("/walle", e)
	assert.Equal(t, true, strings.HasPrefix(b, "20 "))
}

func TestGigMiddleware(t *testing.T) {
	e := New()
	buf := new(bytes.Buffer)

	e.Pre(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			assert.Empty(t, c.Path())
			buf.WriteString("-1")
			return next(c)
		}
	})

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("1")
			return next(c)
		}
	})

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("2")
			return next(c)
		}
	})

	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("3")
			return next(c)
		}
	})

	// Route
	e.Handle("/", func(c Context) error {
		return c.Text(StatusSuccess, "OK")
	})

	b := request("/", e)
	assert.Equal(t, "-1123", buf.String())
	assert.Equal(t, "20 text/plain; charset=UTF-8\r\nOK", b)
}

func TestGigMiddlewareError(t *testing.T) {
	e := New()
	e.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return NewGeminiErrorFrom(ErrPermanentFailure, "oops")
		}
	})
	e.Handle("/", NotFoundHandler)
	b := request("/", e)
	assert.Equal(t, "50 oops\r\n", b)
}

func TestGigHandler(t *testing.T) {
	e := New()

	// HandlerFunc
	e.Handle("/ok", func(c Context) error {
		return c.Text(StatusSuccess, "OK")
	})

	b := request("/ok", e)
	assert.Equal(t, "20 text/plain; charset=UTF-8\r\nOK", b)
}

func TestGigHandle(t *testing.T) {
	e := New()
	e.Handle("/", func(c Context) error {
		return c.Text(StatusSuccess, "hello")
	})
	b := request("/", e)
	assert.Equal(t, "20 text/plain; charset=UTF-8\r\nhello", b)
}

func TestGigURL(t *testing.T) {
	e := New()
	static := func(Context) error { return nil }
	getUser := func(Context) error { return nil }
	getFile := func(Context) error { return nil }

	e.Handle("/static/file", static)
	e.Handle("/users/:id", getUser)
	g := e.Group("/group")
	g.Handle("/users/:uid/files/:fid", getFile)

	assert := assert.New(t)

	assert.Equal("/static/file", e.URL(static))
	assert.Equal("/users/:id", e.URL(getUser))
	assert.Equal("/users/1", e.URL(getUser, "1"))
	assert.Equal("/group/users/1/files/:fid", e.URL(getFile, "1"))
	assert.Equal("/group/users/1/files/1", e.URL(getFile, "1", "1"))
}

func TestGigRoutes(t *testing.T) {
	e := New()
	routes := []*Route{
		{"/users/:user/events", ""},
		{"/users/:user/events/public", ""},
		{"/repos/:owner/:repo/git/refs", ""},
		{"/repos/:owner/:repo/git/tags", ""},
	}
	for _, r := range routes {
		e.Handle(r.Path, func(c Context) error {
			return c.Text(StatusSuccess, "OK")
		})
	}

	if assert.Equal(t, len(routes), len(e.Routes())) {
		for _, r := range e.Routes() {
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
	e := New()
	e.Handle("/:id", func(c Context) error {
		return c.NoContentSuccess()
	})
	c := newContext("/with%2Fslash")
	e.ServeGemini(c)
	assert.Equal(t, "20 text/gemini\r\n", c.(*context).conn.(*fakeConn).Written)
}

func TestGigGroup(t *testing.T) {
	e := New()
	buf := new(bytes.Buffer)
	e.Use(MiddlewareFunc(func(next HandlerFunc) HandlerFunc {
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

	e.Handle("/users", h)

	// Group
	g1 := e.Group("/group1")
	g1.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("1")
			return next(c)
		}
	})
	g1.Handle("", h)

	// Nested groups with middleware
	g2 := e.Group("/group2")
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

	request("/users", e)
	assert.Equal(t, "0", buf.String())

	buf.Reset()
	request("/group1", e)
	assert.Equal(t, "01", buf.String())

	buf.Reset()
	request("/group2/group3", e)
	assert.Equal(t, "023", buf.String())
}

func TestGigNotFound(t *testing.T) {
	e := New()
	c := newContext("/files").(*context)
	e.ServeGemini(c)
	assert.Equal(t, "51 Not Found\r\n", c.conn.(*fakeConn).Written)
}

func TestGigContext(t *testing.T) {
	e := New()
	c := e.AcquireContext()
	assert.IsType(t, new(context), c)
	e.ReleaseContext(c)
}

func TestGigStartTLS(t *testing.T) {
	e := New()
	go func() {
		_ = e.StartTLS(":0", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
	}()
	time.Sleep(200 * time.Millisecond)

	e.Close()
}

func TestGigStartTLS_BadAddress(t *testing.T) {
	e := New()
	err := e.StartTLS("garbage address", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
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
			e := New()
			e.HideBanner = true

			go func() {
				err := e.StartTLS(":0", test.cert, test.key)
				if test.expectedErr != nil {
					require.EqualError(t, err, test.expectedErr.Error())
				} else if err != ErrServerClosed { // Prevent the test to fail after closing the servers
					require.NoError(t, err)
				}
			}()
			time.Sleep(200 * time.Millisecond)

			e.Close()
		})
	}
}

func TestGigStartAutoTLS(t *testing.T) {
	e := New()
	errChan := make(chan error)

	go func() {
		errChan <- e.StartAutoTLS(":0")
	}()
	time.Sleep(200 * time.Millisecond)

	select {
	case err := <-errChan:
		assert.NoError(t, err)
	default:
		assert.NoError(t, e.Close())
	}
}

func request(path string, e *Gig) string {
	c := newContext(path).(*context)
	e.ServeGemini(c)
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
	e := New()
	errCh := make(chan error)

	go func() {
		errCh <- e.StartTLS(":0", "_fixture/certs/cert.pem", "_fixture/certs/key.pem")
	}()

	time.Sleep(200 * time.Millisecond)

	if err := e.Close(); err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, e.Close().Error(), "use of closed network connection")

	err := <-errCh
	assert.Equal(t, err.Error(), "gemini: Server closed")
}
