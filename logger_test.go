package gig

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/matryer/is"
)

func TestLogger(t *testing.T) {
	is := is.New(t)

	// Note: Just for the test coverage, not a real test.
	g := New()
	c, _ := g.NewFakeContext("/", nil)

	h := Logger()(func(c Context) error {
		return c.Gemini("test")
	})

	// Status 2x
	is.NoErr(h(c))

	// Status 3x
	c, _ = g.NewFakeContext("/", nil)
	h = Logger()(func(c Context) error {
		return c.NoContent(StatusRedirectTemporary, "test")
	})
	is.NoErr(h(c))

	// Status 4x
	c, _ = g.NewFakeContext("/", nil)
	h = Logger()(func(c Context) error {
		return c.NoContent(StatusSlowDown, "test")
	})
	is.NoErr(h(c))

	// Status 5x with empty path
	c, _ = g.NewFakeContext("/", nil)
	h = Logger()(func(c Context) error {
		return errors.New("error")
	})
	is.NoErr(h(c))

	// Status 6x with empty path
	c, _ = g.NewFakeContext("/", nil)
	h = Logger()(func(c Context) error {
		return c.NoContent(StatusClientCertificateRequired, "test")
	})
	is.NoErr(h(c))
}

func TestLoggerTemplate(t *testing.T) {
	buf := new(bytes.Buffer)
	oldWriter := DefaultWriter
	DefaultWriter = buf

	defer func() {
		DefaultWriter = oldWriter
	}()

	g := New()
	g.Use(LoggerWithConfig(LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","time_unix":"${time_unix}",` +
			`"time_unix_nano":"${time_unix_nano}","time_rfc3339":"${time_rfc3339}",` +
			`"id":"${id}","remote_ip":"${remote_ip}","host":"${host}",` +
			`""uri":"${uri}","status":${status}, "latency":${latency},` +
			`"latency_human":"${latency_human}","bytes_in":${bytes_in}, "path":"${path}", ` +
			`"bytes_out":${bytes_out},"us":"${query}","meta":"${meta}"}` + "\n",
	}))

	g.Handle("/login", func(c Context) error {
		return c.Gemini("Header Logged")
	})

	c, _ := g.NewFakeContext("/login?username=apagano-param&password=secret", nil)
	g.ServeGemini(c)

	cases := []string{
		"apagano-param",
		"\"path\":\"/login\"",
		"\"uri\":\"/login?user",
		"\"remote_ip\":\"192.0.2.1\"",
		"\"status\":20",
		"\"bytes_in\":45,",
		"\"meta\":\"text/gemini",
	}

	for _, token := range cases {
		is := is.New(t)
		t.Run(token, func(t *testing.T) {
			is.True(strings.Contains(buf.String(), token))
		})
	}
}

func TestLoggerCustomTimestamp(t *testing.T) {
	is := is.New(t)
	buf := new(bytes.Buffer)
	oldWriter := DefaultWriter
	DefaultWriter = buf

	defer func() {
		DefaultWriter = oldWriter
	}()

	customTimeFormat := "2006-01-02 15:04:05.00000"
	g := New()
	g.Use(LoggerWithConfig(LoggerConfig{
		Format: `{"time":"${time_custom}","id":"${id}","remote_ip":"${remote_ip}","host":"${host}","user_agent":"${user_agent}",` +
			`"method":"${method}","uri":"${uri}","status":${status}, "latency":${latency},` +
			`"latency_human":"${latency_human}","bytes_in":${bytes_in}, "path":"${path}", "referer":"${referer}",` +
			`"bytes_out":${bytes_out},"ch":"${header:X-Custom-Header}",` +
			`"us":"${query:username}", "cf":"${form:username}", "session":"${cookie:session}"}` + "\n",
		CustomTimeFormat: customTimeFormat,
	}))

	g.Handle("/", func(c Context) error {
		return c.Gemini("custom time stamp test")
	})

	c, _ := g.NewFakeContext("/", nil)
	g.ServeGemini(c)

	var objs map[string]*json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &objs); err != nil {
		is.Fail()
	}

	loggedTime := *(*string)(unsafe.Pointer(objs["time"]))
	_, err := time.Parse(customTimeFormat, loggedTime)
	is.True(err != nil)
}
