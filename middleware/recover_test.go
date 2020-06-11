package middleware

import (
	"bytes"
	"testing"

	"github.com/pitr/gig"
	"github.com/pitr/gig/gigtest"
	"github.com/stretchr/testify/assert"
)

func TestRecover(t *testing.T) {
	e := gig.New()
	buf := new(bytes.Buffer)
	e.Logger.SetOutput(buf)
	c, rec := gigtest.NewContext(e, "/", nil)
	h := Recover()(gig.HandlerFunc(func(c gig.Context) error {
		panic("test")
	}))
	h(c)
	assert.Equal(t, "50 test\r\n", rec.Written)
	assert.Contains(t, buf.String(), "PANIC RECOVER")
}

func TestRecover_Defaults(t *testing.T) {
	e := gig.New()
	buf := new(bytes.Buffer)
	e.Logger.SetOutput(buf)
	c, rec := gigtest.NewContext(e, "/", nil)
	h := RecoverWithConfig(RecoverConfig{})(gig.HandlerFunc(func(c gig.Context) error {
		panic("test")
	}))
	h(c)
	assert.Equal(t, "50 test\r\n", rec.Written)
	assert.Contains(t, buf.String(), "PANIC RECOVER")
}
