package gig

import (
	"bytes"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestRecover(t *testing.T) {
	g := New()
	buf := new(bytes.Buffer)
	oldWriter := DefaultWriter
	DefaultWriter = buf

	defer func() {
		DefaultWriter = oldWriter
	}()

	c, conn := g.NewFakeContext("/", nil)
	h := Recover()(HandlerFunc(func(c Context) error {
		panic("test")
	}))

	is := is.New(t)

	is.NoErr(h(c))
	is.Equal("50 test\r\n", conn.Written)
	is.True(strings.Contains(buf.String(), "PANIC RECOVER"))
}

func TestRecover_Defaults(t *testing.T) {
	g := New()
	buf := new(bytes.Buffer)
	oldWriter := DefaultWriter
	DefaultWriter = buf

	defer func() {
		DefaultWriter = oldWriter
	}()

	c, conn := g.NewFakeContext("/", nil)
	h := RecoverWithConfig(RecoverConfig{})(HandlerFunc(func(c Context) error {
		panic("test")
	}))

	is := is.New(t)

	is.NoErr(h(c))
	is.Equal("50 test\r\n", conn.Written)
	is.True(strings.Contains(buf.String(), "PANIC RECOVER"))
}
