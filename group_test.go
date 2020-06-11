package gig

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO: Fix me
func TestGroup(t *testing.T) {
	g := New().Group("/group")
	h := func(Context) error { return nil }
	g.Handle("/", h)
	g.Static("/static", "/tmp")
	g.File("/walle", "_fixture/images//walle.png")
}

func TestGroupFile(t *testing.T) {
	gig := New()
	g := gig.Group("/group")
	g.File("/walle", "_fixture/images/walle.png")
	expectedData, err := ioutil.ReadFile("_fixture/images/walle.png")
	assert.Nil(t, err)
	c := newContext("/group/walle")
	gig.ServeGemini(c)
	assert.Equal(t, "20 image/png\r\n"+string(expectedData), c.(*context).conn.(*fakeConn).Written)
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

	b := request("/group/40_1", gig)
	assert.Equal(t, "40 oops\r\n", b)
	b = request("/group/40_2", gig)
	assert.Equal(t, "40 another\r\n", b)
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
			return c.Text(StatusSuccess, c.Path())
		}
	}
	h := func(c Context) error {
		return c.Text(StatusSuccess, c.Path())
	}
	g.Use(m1)
	g.Handle("/help", h, m2)
	g.Handle("/*", h, m2)
	g.Handle("", h, m2)
	gig.Handle("unrelated", h, m2)
	gig.Handle("*", h, m2)

	b := request("/group/help", gig)
	assert.Equal(t, "20 text/plain; charset=UTF-8\r\n/group/help", b)
	b = request("/group/help/other", gig)
	assert.Equal(t, "20 text/plain; charset=UTF-8\r\n/group/*", b)
	b = request("/group/404", gig)
	assert.Equal(t, "20 text/plain; charset=UTF-8\r\n/group/*", b)
	b = request("/group", gig)
	assert.Equal(t, "20 text/plain; charset=UTF-8\r\n/group", b)
	b = request("/other", gig)
	assert.Equal(t, "20 text/plain; charset=UTF-8\r\n/*", b)
	b = request("/", gig)
	assert.Equal(t, "20 text/plain; charset=UTF-8\r\n", b)

}
