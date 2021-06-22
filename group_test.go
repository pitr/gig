package gig

import (
	"io/ioutil"
	"strings"
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

func TestGroupStatic(t *testing.T) {
	is := is.New(t)
	gig := New()
	g := gig.Group("/group")

	// OK
	g.Static("/images", "_fixture/images")

	b := request("/group/images/walle.png", gig)
	is.True(strings.HasPrefix(b, "20 image/png\r\n"))

	// No file
	g.Static("/images", "_fixture/scripts")

	b = request("/group/images/bolt.png", gig)
	is.Equal("51 Not Found\r\n", b)

	// Directory
	g.Static("/images", "_fixture/images")

	b = request("/group/images", gig)
	is.Equal("51 Not Found\r\n", b)

	b = request("/group/images/", gig)
	is.Equal("20 text/gemini\r\n# Listing /group/images/\n\n=> /group/images/walle.png walle.png [ 219.9kB ]\n", b)

	// Directory with index.gmi
	g.Static("/d", "_fixture")

	b = request("/group/d/", gig)
	is.Equal("20 text/gemini\r\n# Hello from gig\n\n=> / 🏠 Home\n", b)

	// Sub-directory with index.gmi
	b = request("/group/d/folder", gig)
	is.Equal("20 text/gemini\r\n# Listing /group/d/folder\n\n=> /group/d/folder/about.gmi about.gmi [ 29B ]\n=> /group/d/folder/another.blah another.blah [ 14B ]\n", b)

	// File without known mime
	b = request("/group/d/folder/another.blah", gig)
	is.Equal("20 octet/stream\r\n# Another page", b)

	// Escape
	b = request("/d/../../../../../../../../etc/profile", gig)
	is.Equal(b, "51 Not Found\r\n")
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
