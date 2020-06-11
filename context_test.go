package gig

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"testing"
	"text/template"

	"github.com/labstack/gommon/log"
	testify "github.com/stretchr/testify/assert"
)

type (
	Template struct {
		templates *template.Template
	}
	TemplateFail struct{}
)

func (t *Template) Render(w io.Writer, name string, data interface{}, c Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func (t *TemplateFail) Render(w io.Writer, name string, data interface{}, c Context) error {
	return errors.New("could not render")
}

type responseWriterErr struct{}

func (responseWriterErr) Write([]byte) (int, error) {
	return 0, errors.New("err")
}
func (responseWriterErr) WriteHeader(statusCode Status, mime string) {}

func TestContext(t *testing.T) {
	c := newContext("/").(*context)

	assert := testify.New(t)

	// Gig
	assert.NotNil(c.Gig())

	// Conn
	assert.NotNil(c.conn)

	// Response
	assert.NotNil(c.Response())

	//--------
	// Render
	//--------

	c.gig.Renderer = &Template{
		templates: template.Must(template.New("hello").Parse("Hello, {{.}}!")),
	}
	err := c.Render(StatusSuccess, "hello", "Jon Snow")
	if assert.NoError(err) {
		assert.Equal(fmt.Sprintf("%d %s\r\nHello, Jon Snow!", StatusSuccess, MIMETextGeminiCharsetUTF8), c.conn.(*fakeConn).Written)
	}

	c.gig.Renderer = &TemplateFail{}
	err = c.Render(StatusSuccess, "hello", "Jon Snow")
	assert.Error(err)

	c.gig.Renderer = nil
	err = c.Render(StatusSuccess, "hello", "Jon Snow")
	assert.Error(err)

	// JSON
	c = newContext("/").(*context)

	err = c.JSON(StatusSuccess, user{1, "Jon Snow"})
	if assert.NoError(err) {
		assert.Equal(fmt.Sprintf("%d %s\r\n%s\n", StatusSuccess, MIMEApplicationJSONCharsetUTF8, userJSON), c.conn.(*fakeConn).Written)
	}

	// JSONPretty
	c = newContext("/").(*context)

	err = c.JSONPretty(StatusSuccess, user{1, "Jon Snow"}, "  ")
	if assert.NoError(err) {
		assert.Equal(fmt.Sprintf("%d %s\r\n%s\n", StatusSuccess, MIMEApplicationJSONCharsetUTF8, userJSONPretty), c.conn.(*fakeConn).Written)
	}

	// JSON (error)
	c = newContext("/").(*context)

	err = c.JSON(StatusSuccess, make(chan bool))
	assert.Error(err)

	// XML
	c = newContext("/").(*context)

	err = c.XML(StatusSuccess, user{1, "Jon Snow"})
	if assert.NoError(err) {
		assert.Equal(fmt.Sprintf("%d %s\r\n%s", StatusSuccess, MIMEApplicationXMLCharsetUTF8, xml.Header+userXML), c.conn.(*fakeConn).Written)
	}

	// XML (error)
	c = newContext("/").(*context)

	err = c.XML(StatusSuccess, make(chan bool))
	assert.Error(err)

	// XML response write error
	c = newContext("/").(*context)

	c.response.Writer = responseWriterErr{}
	err = c.XML(0, 0)
	testify.Error(t, err)

	// XMLPretty
	c = newContext("/").(*context)

	err = c.XMLPretty(StatusSuccess, user{1, "Jon Snow"}, "  ")
	if assert.NoError(err) {
		assert.Equal(fmt.Sprintf("%d %s\r\n%s", StatusSuccess, MIMEApplicationXMLCharsetUTF8, xml.Header+userXMLPretty), c.conn.(*fakeConn).Written)
	}

	t.Run("empty indent", func(t *testing.T) {
		var (
			u           = user{1, "Jon Snow"}
			buf         = new(bytes.Buffer)
			emptyIndent = ""
		)

		t.Run("json", func(t *testing.T) {
			buf.Reset()
			assert := testify.New(t)

			// JSONBlob with empty indent
			c := newContext("/").(*context)

			enc := json.NewEncoder(buf)
			enc.SetIndent(emptyIndent, emptyIndent)
			err = enc.Encode(u)
			err = c.json(StatusSuccess, user{1, "Jon Snow"}, emptyIndent)
			if assert.NoError(err) {
				assert.Equal(fmt.Sprintf("%d %s\r\n%s", StatusSuccess, MIMEApplicationJSONCharsetUTF8, buf.String()), c.conn.(*fakeConn).Written)
			}

			c = newContext("/").(*context)
			c.conn.(*fakeConn).failAfter = 1
			assert.Error(c.json(StatusSuccess, user{1, "Jon Snow"}, emptyIndent))
		})

		t.Run("xml", func(t *testing.T) {
			buf.Reset()
			assert := testify.New(t)

			// XMLBlob with empty indent
			c := newContext("/").(*context)

			enc := xml.NewEncoder(buf)
			enc.Indent(emptyIndent, emptyIndent)
			err = enc.Encode(u)
			err = c.xml(StatusSuccess, user{1, "Jon Snow"}, emptyIndent)
			if assert.NoError(err) {
				assert.Equal(fmt.Sprintf("%d %s\r\n%s", StatusSuccess, MIMEApplicationXMLCharsetUTF8, xml.Header+buf.String()), c.conn.(*fakeConn).Written)
			}

			c = newContext("/").(*context)
			c.conn.(*fakeConn).failAfter = 1
			assert.Error(c.xml(StatusSuccess, user{1, "Jon Snow"}, emptyIndent))

			c = newContext("/").(*context)
			c.conn.(*fakeConn).failAfter = 40
			assert.Error(c.xml(StatusSuccess, user{1, "Jon Snow"}, emptyIndent))
		})
	})

	// JSONBlob
	c = newContext("/").(*context)

	data, err := json.Marshal(user{1, "Jon Snow"})
	assert.NoError(err)
	err = c.JSONBlob(StatusSuccess, data)
	if assert.NoError(err) {
		assert.Equal(fmt.Sprintf("%d %s\r\n%s", StatusSuccess, MIMEApplicationJSONCharsetUTF8, userJSON), c.conn.(*fakeConn).Written)
	}

	c = newContext("/").(*context)
	c.conn.(*fakeConn).failAfter = 1
	assert.Error(c.JSONBlob(StatusSuccess, data))

	// XMLBlob
	c = newContext("/").(*context)

	data, err = xml.Marshal(user{1, "Jon Snow"})
	assert.NoError(err)
	err = c.XMLBlob(StatusSuccess, data)
	if assert.NoError(err) {
		assert.Equal(fmt.Sprintf("%d %s\r\n%s", StatusSuccess, MIMEApplicationXMLCharsetUTF8, xml.Header+userXML), c.conn.(*fakeConn).Written)
	}

	c = newContext("/").(*context)
	c.conn.(*fakeConn).failAfter = 1
	assert.Error(c.XMLBlob(StatusSuccess, data))

	c = newContext("/").(*context)
	c.conn.(*fakeConn).failAfter = 40
	assert.Error(c.XMLBlob(StatusSuccess, data))

	// Text
	c = newContext("/").(*context)

	err = c.Text(StatusSuccess, "Hello, World!")
	if assert.NoError(err) {
		assert.Equal(fmt.Sprintf("%d %s\r\nHello, World!", StatusSuccess, MIMETextPlainCharsetUTF8), c.conn.(*fakeConn).Written)
	}

	// Gemini
	c = newContext("/").(*context)

	err = c.Gemini(StatusSuccess, "Hello, World!")
	if assert.NoError(err) {
		assert.Equal(fmt.Sprintf("%d %s\r\nHello, World!", StatusSuccess, MIMETextGeminiCharsetUTF8), c.conn.(*fakeConn).Written)
	}

	// Stream
	c = newContext("/").(*context)

	r := strings.NewReader("response from a stream")
	err = c.Stream(StatusSuccess, "application/octet-stream", r)
	if assert.NoError(err) {
		assert.Equal(fmt.Sprintf("%d application/octet-stream\r\nresponse from a stream", StatusSuccess), c.conn.(*fakeConn).Written)
	}

	c = newContext("/").(*context)
	c.conn.(*fakeConn).failAfter = 1
	assert.Error(c.Stream(StatusSuccess, "application/octet-stream", r))

	// NoContentSuccess
	c = newContext("/").(*context)

	_ = c.NoContentSuccess()
	assert.Equal(fmt.Sprintf("%d text/gemini\r\n", StatusSuccess), c.conn.(*fakeConn).Written)

	// Error
	c = newContext("/").(*context)

	c.Error(errors.New("error"))
	assert.Equal(fmt.Sprintf("%d error\r\n", StatusPermanentFailure), c.conn.(*fakeConn).Written)

	// Reset
	c.Set("foe", "ban")
	c.Reset(nil, nil, "", nil)
	assert.Equal(0, len(c.store))
	assert.Equal("", c.Path())
}

func TestContext_JSON_CommitsCustomResponseCode(t *testing.T) {
	c := newContext("/").(*context)

	err := c.JSON(StatusSuccess, user{1, "Jon Snow"})

	assert := testify.New(t)
	if assert.NoError(err) {
		assert.Equal(fmt.Sprintf("%d %s\r\n%s\n", StatusSuccess, MIMEApplicationJSONCharsetUTF8, userJSON), c.conn.(*fakeConn).Written)
	}
}

func TestContext_JSON_HandlesBadJSON(t *testing.T) {
	c := newContext("/").(*context)

	err := c.JSON(StatusSuccess, map[string]float64{"a": math.NaN()})

	assert := testify.New(t)
	assert.Error(err)
}

func TestContextPath(t *testing.T) {
	e := New()
	r := e.Router()

	r.Add("/users/:id", nil)
	c := e.NewContext(nil, nil, "", nil)
	r.Find("/users/1", c)

	assert := testify.New(t)

	assert.Equal("/users/:id", c.Path())

	r.Add("/users/:uid/files/:fid", nil)
	c = e.NewContext(nil, nil, "", nil)
	r.Find("/users/1/files/1", c)
	assert.Equal("/users/:uid/files/:fid", c.Path())
}

func TestContextRequestURI(t *testing.T) {
	e := New()

	c := e.NewContext(nil, nil, "/my-uri", nil)

	assert := testify.New(t)

	assert.Equal("/my-uri", c.RequestURI())
}

func TestContextGetParam(t *testing.T) {
	e := New()
	r := e.Router()
	r.Add("/:foo", func(Context) error { return nil })
	c := newContext("/bar")

	// round-trip param values with modification
	testify.Equal(t, "", c.Param("bar"))

	// shouldn't explode during Reset() afterwards!
	testify.NotPanics(t, func() {
		c.Reset(nil, nil, "", nil)
	})
}

func TestContextRedirect(t *testing.T) {
	c := newContext("/").(*context)

	testify.Equal(t, nil, c.Redirect(StatusRedirectPermanent, "gemini://gus.guru/"))
	testify.Equal(t, "31 gemini://gus.guru/\r\n", c.conn.(*fakeConn).Written)
	testify.Error(t, c.Redirect(StatusSuccess, "gemini://gus.guru/"))
}

func TestContextStore(t *testing.T) {
	c := new(context)
	c.Set("name", "Jon Snow")
	testify.Equal(t, "Jon Snow", c.Get("name"))
}

func BenchmarkContext_Store(b *testing.B) {
	e := &Gig{}

	c := &context{
		gig: e,
	}

	for n := 0; n < b.N; n++ {
		c.Set("name", "Jon Snow")
		if c.Get("name") != "Jon Snow" {
			b.Fail()
		}
	}
}

func TestContextHandler(t *testing.T) {
	e := New()
	r := e.Router()
	b := new(bytes.Buffer)

	r.Add("/handler", func(Context) error {
		_, err := b.Write([]byte("handler"))
		return err
	})
	c := e.NewContext(nil, nil, "", nil)
	r.Find("/handler", c)
	err := c.Handler()(c)
	testify.Equal(t, "handler", b.String())
	testify.NoError(t, err)
}

func TestContext_SetHandler(t *testing.T) {
	c := new(context)

	testify.Nil(t, c.Handler())

	c.SetHandler(func(c Context) error {
		return nil
	})
	testify.NotNil(t, c.Handler())
}

func TestContext_Path(t *testing.T) {
	path := "/pa/th"

	c := new(context)

	c.SetPath(path)
	testify.Equal(t, path, c.Path())
}

func TestContext_QueryString(t *testing.T) {
	queryString := "some+val"

	c := newContext("/?" + queryString)

	testify.Equal(t, queryString, c.QueryString())
}

func TestContext_Logger(t *testing.T) {
	c := newContext("/")

	log1 := c.Logger()
	testify.NotNil(t, log1)

	log2 := log.New("gig2")
	c.SetLogger(log2)
	testify.Equal(t, log2, c.Logger())

	// Resetting the context returns the initial logger
	c.Reset(nil, nil, "", nil)
	testify.Equal(t, log1, c.Logger())
}

func TestContext_IP(t *testing.T) {
	c := newContext("/").(*context)

	testify.Equal(t, "192.0.2.1", c.IP())
}

func TestContext_Certificate(t *testing.T) {
	c := newContext("/").(*context)
	testify.Nil(t, c.Certificate())

	cert := &x509.Certificate{}
	c.TLS = &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{cert},
	}

	testify.Equal(t, cert, c.Certificate())
}
