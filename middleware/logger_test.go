package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"
	"unsafe"

	"github.com/pitr/gig"
	"github.com/pitr/gig/gigtest"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	// Note: Just for the test coverage, not a real test.
	e := gig.New()
	c, _ := gigtest.NewContext(e, "/", nil)

	h := Logger()(func(c gig.Context) error {
		return c.Gemini(gig.StatusSuccess, "test")
	})

	// Status 2x
	assert.Nil(t, h(c))

	// Status 3x
	c, _ = gigtest.NewContext(e, "/", nil)
	h = Logger()(func(c gig.Context) error {
		return c.NoContent(gig.StatusRedirectTemporary, "test")
	})
	assert.Nil(t, h(c))

	// Status 4x
	c, _ = gigtest.NewContext(e, "/", nil)
	h = Logger()(func(c gig.Context) error {
		return c.NoContent(gig.StatusSlowDown, "test")
	})
	assert.Nil(t, h(c))

	// Status 5x with empty path
	c, _ = gigtest.NewContext(e, "/", nil)
	h = Logger()(func(c gig.Context) error {
		return errors.New("error")
	})
	assert.Nil(t, h(c))

	// Status 6x with empty path
	c, _ = gigtest.NewContext(e, "/", nil)
	h = Logger()(func(c gig.Context) error {
		return c.NoContent(gig.StatusTransientCertificateRequested, "test")
	})
	assert.Nil(t, h(c))
}

func TestLoggerTemplate(t *testing.T) {
	buf := new(bytes.Buffer)

	e := gig.New()
	e.Use(LoggerWithConfig(LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","time_unix":"${time_unix}",` +
			`"time_unix_nano":"${time_unix_nano}","time_rfc3339":"${time_rfc3339}",` +
			`"id":"${id}","remote_ip":"${remote_ip}","host":"${host}",` +
			`""uri":"${uri}","status":${status}, "latency":${latency},` +
			`"latency_human":"${latency_human}","bytes_in":${bytes_in}, "path":"${path}", ` +
			`"bytes_out":${bytes_out},"us":"${query}","meta":"${meta}"}` + "\n",
		Output: buf,
	}))

	e.Handle("/login", func(c gig.Context) error {
		return c.Gemini(gig.StatusSuccess, "Header Logged")
	})

	c, _ := gigtest.NewContext(e, "/login?username=apagano-param&password=secret", nil)
	e.ServeGemini(c)

	cases := map[string]bool{
		"apagano-param":               true,
		"\"path\":\"/login\"":         true,
		"\"uri\":\"/login?user":       true,
		"\"remote_ip\":\"192.0.2.1\"": true,
		"\"status\":20":               true,
		"\"bytes_in\":45,":            true,
		"\"meta\":\"text/gemini":      true,
	}

	for token, present := range cases {
		if present {
			assert.Contains(t, buf.String(), token, "Case: "+token)
		} else {
			assert.NotContains(t, buf.String(), token, "Case: "+token)
		}
	}
}

func TestLoggerCustomTimestamp(t *testing.T) {
	buf := new(bytes.Buffer)
	customTimeFormat := "2006-01-02 15:04:05.00000"
	e := gig.New()
	e.Use(LoggerWithConfig(LoggerConfig{
		Format: `{"time":"${time_custom}","id":"${id}","remote_ip":"${remote_ip}","host":"${host}","user_agent":"${user_agent}",` +
			`"method":"${method}","uri":"${uri}","status":${status}, "latency":${latency},` +
			`"latency_human":"${latency_human}","bytes_in":${bytes_in}, "path":"${path}", "referer":"${referer}",` +
			`"bytes_out":${bytes_out},"ch":"${header:X-Custom-Header}",` +
			`"us":"${query:username}", "cf":"${form:username}", "session":"${cookie:session}"}` + "\n",
		CustomTimeFormat: customTimeFormat,
		Output:           buf,
	}))

	e.Handle("/", func(c gig.Context) error {
		return c.Gemini(gig.StatusSuccess, "custom time stamp test")
	})

	c, _ := gigtest.NewContext(e, "/", nil)
	e.ServeGemini(c)

	var objs map[string]*json.RawMessage
	if err := json.Unmarshal([]byte(buf.String()), &objs); err != nil {
		panic(err)
	}
	loggedTime := *(*string)(unsafe.Pointer(objs["time"]))
	_, err := time.Parse(customTimeFormat, loggedTime)
	assert.Error(t, err)
}
