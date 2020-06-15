package gig

import (
	"io/ioutil"
	"testing"

	"github.com/matryer/is"
)

func TestGroupFile(t *testing.T) {
	gig := New()
	g := gig.Group("/group")
	g.File("/walle", "_fixture/images/walle.png")

	expectedData, err := ioutil.ReadFile("_fixture/images/walle.png")

	is := is.New(t)

	is.NoErr(err)

	c, conn := gig.NewFakeContext("/group/walle", nil)
	gig.ServeGemini(c)
	is.Equal("20 image/png\r\n"+string(expectedData), conn.Written)
}

func TestGroupRouteMiddleware(t *testing.T) {
	// Ensure middleware slices are not re-used
	gig := New()
	g := gig.Group("/group")
	h := func(Context) error { return nil }
	m1 := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return next(c)
		}
	}
	m2 := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return next(c)
		}
	}
	m3 := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return next(c)
		}
	}
	m4 := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return c.NoContent(40, "oops")
		}
	}
	m5 := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return c.NoContent(40, "another")
		}
	}

	g.Use(m1, m2, m3)
	g.Handle("/40_1", h, m4)
	g.Handle("/40_2", h, m5)

	is := is.New(t)

	b := request("/group/40_1", gig)
	is.Equal("40 oops\r\n", b)
	b = request("/group/40_2", gig)
	is.Equal("40 another\r\n", b)
}

func TestGroupRouteMiddlewareWithMatchAny(t *testing.T) {
	// Ensure middleware and match any routes do not conflict
	gig := New()
	g := gig.Group("/group")
	m1 := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return next(c)
		}
	}
	m2 := func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			return c.Text(c.Path())
		}
	}
	h := func(c Context) error {
		return c.Text(c.Path())
	}

	g.Use(m1)
	g.Handle("/help", h, m2)
	g.Handle("/*", h, m2)
	g.Handle("", h, m2)
	gig.Handle("unrelated", h, m2)
	gig.Handle("*", h, m2)

	is := is.New(t)

	b := request("/group/help", gig)
	is.Equal("20 text/plain\r\n/group/help", b)
	b = request("/group/help/other", gig)
	is.Equal("20 text/plain\r\n/group/*", b)
	b = request("/group/404", gig)
	is.Equal("20 text/plain\r\n/group/*", b)
	b = request("/group", gig)
	is.Equal("20 text/plain\r\n/group", b)
	b = request("/other", gig)
	is.Equal("20 text/plain\r\n/*", b)
	b = request("/", gig)
	is.Equal("20 text/plain\r\n", b)
}
