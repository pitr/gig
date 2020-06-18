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

func TestContext(t *testing.T) {
	g := New()
	c, conn := g.NewFakeContext("/", nil)

	is := is.New(t)

	// Gig
	is.True(c.Gig() != nil)

	// Conn
	if conn == nil {
		panic("staticcheck SA5011 false positive, conn cannot be nil here")
	}

	// Response
	is.True(c.Response() != nil)

	//--------
	// Render
	//--------

	g.Renderer = &Template{
		templates: template.Must(template.New("hello").Parse("Hello, {{.}}!")),
	}
	err := c.Render("hello", "Jon Snow")
	is.NoErr(err)
	is.Equal("20 text/gemini\r\nHello, Jon Snow!", conn.Written)

	g.Renderer = &TemplateFail{}
	err = c.Render("hello", "Jon Snow")
	is.True(err != nil)

	g.Renderer = nil
	err = c.Render("hello", "Jon Snow")
	is.True(err != nil)

	// Text
	c, conn = g.NewFakeContext("/", nil)

	err = c.Text("Hello, %s!", "World")
	is.NoErr(err)
	is.Equal(fmt.Sprintf("%d %s\r\nHello, World!", StatusSuccess, MIMETextPlain), conn.Written)

	// Gemini
	c, conn = g.NewFakeContext("/", nil)

	err = c.Gemini("Hello, %s!", "World")
	is.NoErr(err)
	is.Equal(fmt.Sprintf("%d %s\r\nHello, World!", StatusSuccess, MIMETextGemini), conn.Written)

	// Stream
	c, conn = g.NewFakeContext("/", nil)

	r := strings.NewReader("response from a stream")
	err = c.Stream("application/octet-stream", r)
	is.NoErr(err)
	is.Equal(fmt.Sprintf("%d application/octet-stream\r\nresponse from a stream", StatusSuccess), conn.Written)

	c, conn = g.NewFakeContext("/", nil)
	conn.FailAfter = 1

	is.True(c.Stream("application/octet-stream", r) != nil)

	// Error
	c, conn = g.NewFakeContext("/", nil)

	c.Error(errors.New("error"))
	is.Equal(fmt.Sprintf("%d error\r\n", StatusPermanentFailure), conn.Written)

	// Reset
	c.Set("foe", "ban")
	c.(*context).reset(nil, nil, "", nil)
	is.Equal(0, len(c.(*context).store))
	is.Equal("", c.Path())
}

func TestContextPath(t *testing.T) {
	g := New()
	r := g.router

	r.add("/users/:id", nil)
	c := g.newContext(nil, nil, "", nil)
	r.find("/users/1", c)

	is := is.New(t)

	is.Equal("/users/:id", c.Path())

	r.add("/users/:uid/files/:fid", nil)
	c = g.newContext(nil, nil, "", nil)
	r.find("/users/1/files/1", c)
	is.Equal("/users/:uid/files/:fid", c.Path())
}

func TestContextRequestURI(t *testing.T) {
	g := New()

	c := g.newContext(nil, nil, "/my-uri", nil)

	is := is.New(t)

	is.Equal("/my-uri", c.RequestURI())
}

func TestContextGetParam(t *testing.T) {
	g := New()
	is := is.New(t)
	r := g.router

	r.add("/:foo", func(Context) error { return nil })

	c, _ := g.NewFakeContext("/bar", nil)

	// round-trip param values with modification
	is.Equal("", c.Param("bar"))

	// shouldn't explode during Reset() afterwards!
	c.(*context).reset(nil, nil, "", nil)
}

func TestContextFile(t *testing.T) {
	g := New()
	is := is.New(t)
	c, conn := g.NewFakeContext("/", nil)

	is.NoErr(c.File("_fixture/folder/about.gmi"))
	is.Equal("20 text/gemini\r\n# About page\n\n=> / 🏠 Home\n", conn.Written)

	c, conn = g.NewFakeContext("/", nil)

	is.NoErr(c.File("../../../../../../../../etc/profile"))
	is.Equal("59 Bad Request\r\n", conn.Written)
}

func TestContextNoContent(t *testing.T) {
	c, conn := New().NewFakeContext("/", nil)
	is := is.New(t)

	is.NoErr(c.NoContent(StatusRedirectPermanent, "gemini://gus.guru/"))
	is.Equal("31 gemini://gus.guru/\r\n", conn.Written)
}

func TestContextStore(t *testing.T) {
	c := new(context)
	c.Set("name", "Jon Snow")

	is := is.New(t)
	is.Equal("Jon Snow", c.Get("name"))
}

func BenchmarkContext_Store(b *testing.B) {
	b.ReportAllocs()

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
	r := g.router
	b := new(bytes.Buffer)

	r.add("/handler", func(Context) error {
		_, err := b.Write([]byte("handler"))
		return err
	})

	c := g.newContext(nil, nil, "", nil)
	r.find("/handler", c)
	err := c.Handler()(c)

	is := is.New(t)
	is.Equal("handler", b.String())
	is.NoErr(err)
}

func TestContext_Path(t *testing.T) {
	path := "/pa/th"

	c := new(context)
	is := is.New(t)

	c.path = path
	is.Equal(path, c.Path())
}

func TestContext_QueryString(t *testing.T) {
	queryString := "some+val"

	c, _ := New().NewFakeContext("/?"+queryString, nil)
	is := is.New(t)

	q, err := c.QueryString()
	is.NoErr(err)
	is.Equal("some val", q)
}

func TestContext_IP(t *testing.T) {
	c, _ := New().NewFakeContext("/", nil)

	is := is.New(t)
	is.Equal("192.0.2.1", c.IP())
}

func TestContext_Certificate(t *testing.T) {
	c, _ := New().NewFakeContext("/", nil)
	is := is.New(t)

	is.Equal(c.Certificate(), nil)

	cert := &x509.Certificate{}
	c, _ = New().NewFakeContext("/", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{cert},
	})

	is.Equal(cert, c.Certificate())
}

func TestContext_bytefmt(t *testing.T) {
	is := is.New(t)

	is.Equal(bytefmt(12345678), "12.3MB")
}
