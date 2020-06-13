package middleware

import (
	"bytes"
	"strings"
	"testing"

	"github.com/matryer/is"
	"github.com/pitr/gig"
	"github.com/pitr/gig/gigtest"
)

func TestRecover(t *testing.T) {
	g := gig.New()
	buf := new(bytes.Buffer)
	oldWriter := gig.DefaultWriter
	gig.DefaultWriter = buf

	defer func() {
		gig.DefaultWriter = oldWriter
	}()

	c, rec := gigtest.NewContext(g, "/", nil)
	h := Recover()(gig.HandlerFunc(func(c gig.Context) error {
		panic("test")
	}))

	is := is.New(t)

	is.NoErr(h(c))
	is.Equal("50 test\r\n", rec.Written)
	is.True(strings.Contains(buf.String(), "PANIC RECOVER"))
}

func TestRecover_Defaults(t *testing.T) {
	g := gig.New()
	buf := new(bytes.Buffer)
	oldWriter := gig.DefaultWriter
	gig.DefaultWriter = buf

	defer func() {
		gig.DefaultWriter = oldWriter
	}()

	c, rec := gigtest.NewContext(g, "/", nil)
	h := RecoverWithConfig(RecoverConfig{})(gig.HandlerFunc(func(c gig.Context) error {
		panic("test")
	}))

	is := is.New(t)

	is.NoErr(h(c))
	is.Equal("50 test\r\n", rec.Written)
	is.True(strings.Contains(buf.String(), "PANIC RECOVER"))
}
