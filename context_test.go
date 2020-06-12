package gig

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"text/template"

	"github.com/matryer/is"
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

	is := is.New(t)

	// Gig
	is.True(c.Gig() != nil)

	// Conn
	is.True(c.conn != nil)

	// Response
	is.True(c.Response() != nil)

	//--------
	// Render
	//--------

	c.gig.Renderer = &Template{
		templates: template.Must(template.New("hello").Parse("Hello, {{.}}!")),
	}
	err := c.Render(StatusSuccess, "hello", "Jon Snow")
	is.NoErr(err)
	is.Equal(fmt.Sprintf("%d %s\r\nHello, Jon Snow!", StatusSuccess, MIMETextGeminiCharsetUTF8), c.conn.(*fakeConn).Written)

	c.gig.Renderer = &TemplateFail{}
	err = c.Render(StatusSuccess, "hello", "Jon Snow")
	is.True(err != nil)

	c.gig.Renderer = nil
	err = c.Render(StatusSuccess, "hello", "Jon Snow")
	is.True(err != nil)

	// Text
	c = newContext("/").(*context)

	err = c.Text(StatusSuccess, "Hello, World!")
	is.NoErr(err)
	is.Equal(fmt.Sprintf("%d %s\r\nHello, World!", StatusSuccess, MIMETextPlainCharsetUTF8), c.conn.(*fakeConn).Written)

	// Gemini
	c = newContext("/").(*context)

	err = c.Gemini(StatusSuccess, "Hello, World!")
	is.NoErr(err)
	is.Equal(fmt.Sprintf("%d %s\r\nHello, World!", StatusSuccess, MIMETextGeminiCharsetUTF8), c.conn.(*fakeConn).Written)

	// Stream
	c = newContext("/").(*context)

	r := strings.NewReader("response from a stream")
	err = c.Stream(StatusSuccess, "application/octet-stream", r)
	is.NoErr(err)
	is.Equal(fmt.Sprintf("%d application/octet-stream\r\nresponse from a stream", StatusSuccess), c.conn.(*fakeConn).Written)

	c = newContext("/").(*context)
	c.conn.(*fakeConn).failAfter = 1
	is.True(c.Stream(StatusSuccess, "application/octet-stream", r) != nil)

	// NoContentSuccess
	c = newContext("/").(*context)

	_ = c.NoContentSuccess()
	is.Equal(fmt.Sprintf("%d text/gemini\r\n", StatusSuccess), c.conn.(*fakeConn).Written)

	// Error
	c = newContext("/").(*context)

	c.Error(errors.New("error"))
	is.Equal(fmt.Sprintf("%d error\r\n", StatusPermanentFailure), c.conn.(*fakeConn).Written)

	// Reset
	c.Set("foe", "ban")
	c.Reset(nil, nil, "", nil)
	is.Equal(0, len(c.store))
	is.Equal("", c.Path())
}

func TestContextPath(t *testing.T) {
	g := New()
	r := g.Router()

	r.Add("/users/:id", nil)
	c := g.NewContext(nil, nil, "", nil)
	r.Find("/users/1", c)

	is := is.New(t)

	is.Equal("/users/:id", c.Path())

	r.Add("/users/:uid/files/:fid", nil)
	c = g.NewContext(nil, nil, "", nil)
	r.Find("/users/1/files/1", c)
	is.Equal("/users/:uid/files/:fid", c.Path())
}

func TestContextRequestURI(t *testing.T) {
	g := New()

	c := g.NewContext(nil, nil, "/my-uri", nil)

	is := is.New(t)

	is.Equal("/my-uri", c.RequestURI())
}

func TestContextGetParam(t *testing.T) {
	g := New()
	r := g.Router()
	r.Add("/:foo", func(Context) error { return nil })
	c := newContext("/bar")

	is := is.New(t)

	// round-trip param values with modification
	is.Equal("", c.Param("bar"))

	// shouldn't explode during Reset() afterwards!
	c.Reset(nil, nil, "", nil)
}

func TestContextRedirect(t *testing.T) {
	c := newContext("/").(*context)
	is := is.New(t)

	is.Equal(nil, c.Redirect(StatusRedirectPermanent, "gemini://gus.guru/"))
	is.Equal("31 gemini://gus.guru/\r\n", c.conn.(*fakeConn).Written)
	is.True(c.Redirect(StatusSuccess, "gemini://gus.guru/") != nil)
}

func TestContextStore(t *testing.T) {
	c := new(context)
	c.Set("name", "Jon Snow")
	is := is.New(t)
	is.Equal("Jon Snow", c.Get("name"))
}

func BenchmarkContext_Store(b *testing.B) {
	g := &Gig{}

	c := &context{
		gig: g,
	}

	for n := 0; n < b.N; n++ {
		c.Set("name", "Jon Snow")
		if c.Get("name") != "Jon Snow" {
			b.Fail()
		}
	}
}

func TestContextHandler(t *testing.T) {
	g := New()
	r := g.Router()
	b := new(bytes.Buffer)

	r.Add("/handler", func(Context) error {
		_, err := b.Write([]byte("handler"))
		return err
	})
	c := g.NewContext(nil, nil, "", nil)
	r.Find("/handler", c)
	err := c.Handler()(c)

	is := is.New(t)
	is.Equal("handler", b.String())
	is.NoErr(err)
}

func TestContext_SetHandler(t *testing.T) {
	c := new(context)
	is := is.New(t)

	is.Equal(c.Handler(), nil)

	c.SetHandler(func(c Context) error {
		return nil
	})
	is.True(c.Handler() != nil)
}

func TestContext_Path(t *testing.T) {
	path := "/pa/th"

	c := new(context)
	is := is.New(t)

	c.SetPath(path)
	is.Equal(path, c.Path())
}

func TestContext_QueryString(t *testing.T) {
	queryString := "some+val"

	c := newContext("/?" + queryString)
	is := is.New(t)

	is.Equal(queryString, c.QueryString())
}

func TestContext_IP(t *testing.T) {
	c := newContext("/").(*context)

	is := is.New(t)
	is.Equal("192.0.2.1", c.IP())
}

func TestContext_Certificate(t *testing.T) {
	c := newContext("/").(*context)
	is := is.New(t)

	is.Equal(c.Certificate(), nil)

	cert := &x509.Certificate{}
	c.TLS = &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{cert},
	}

	is.Equal(cert, c.Certificate())
}
