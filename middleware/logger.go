package middleware

import (
	"bytes"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/gommon/color"
	"github.com/pitr/gig"
	"github.com/valyala/fasttemplate"
)

type (
	// LoggerConfig defines the config for Logger middleware.
	LoggerConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Tags to construct the logger format.
		//
		// - time_unix
		// - time_unix_nano
		// - time_rfc3339
		// - time_rfc3339_nano
		// - time_custom
		// - remote_ip
		// - uri
		// - host
		// - path
		// - status
		// - error
		// - latency (In nanoseconds)
		// - latency_human (Human readable)
		// - bytes_in (Bytes received)
		// - bytes_out (Bytes sent)
		// - meta
		// - query
		//
		// Example "${remote_ip} ${status}"
		//
		// Optional. Default value DefaultLoggerConfig.Format.
		Format string `yaml:"format"`

		// Optional. Default value DefaultLoggerConfig.CustomTimeFormat.
		CustomTimeFormat string `yaml:"custom_time_format"`

		// Output is a writer where logs in JSON format are written.
		// Optional. Default value os.Stdout.
		Output io.Writer

		template *fasttemplate.Template
		colorer  *color.Color
		pool     *sync.Pool
	}
)

var (
	// DefaultLoggerConfig is the default Logger middleware config.
	DefaultLoggerConfig = LoggerConfig{
		Skipper:          DefaultSkipper,
		Format:           `path=${path} status=${status} duration=${latency} ${error}\n`,
		CustomTimeFormat: "2006-01-02 15:04:05.00000",
		colorer:          color.New(),
	}
)

// Logger returns a middleware that logs Gemini requests.
func Logger() gig.MiddlewareFunc {
	return LoggerWithConfig(DefaultLoggerConfig)
}

// LoggerWithConfig returns a Logger middleware with config.
// See: `Logger()`.
func LoggerWithConfig(config LoggerConfig) gig.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultLoggerConfig.Skipper
	}
	if config.Format == "" {
		config.Format = DefaultLoggerConfig.Format
	}
	if config.Output == nil {
		config.Output = DefaultLoggerConfig.Output
	}

	config.template = fasttemplate.New(config.Format, "${", "}")
	config.colorer = color.New()
	config.colorer.SetOutput(config.Output)
	config.pool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 256))
		},
	}

	return func(next gig.HandlerFunc) gig.HandlerFunc {
		return func(c gig.Context) (err error) {
			if config.Skipper(c) {
				return next(c)
			}

			start := time.Now()
			if err = next(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()
			buf := config.pool.Get().(*bytes.Buffer)
			buf.Reset()
			defer config.pool.Put(buf)

			res := c.Response()

			if _, err = config.template.ExecuteFunc(buf, func(w io.Writer, tag string) (int, error) {
				switch tag {
				case "time_unix":
					return buf.WriteString(strconv.FormatInt(time.Now().Unix(), 10))
				case "time_unix_nano":
					return buf.WriteString(strconv.FormatInt(time.Now().UnixNano(), 10))
				case "time_rfc3339":
					return buf.WriteString(time.Now().Format(time.RFC3339))
				case "time_rfc3339_nano":
					return buf.WriteString(time.Now().Format(time.RFC3339Nano))
				case "time_custom":
					return buf.WriteString(time.Now().Format(config.CustomTimeFormat))
				case "remote_ip":
					return buf.WriteString(c.IP())
				case "host":
					return buf.WriteString(c.URL().Host)
				case "uri":
					return buf.WriteString(c.RequestURI())
				case "path":
					p := c.Path()
					if p == "" {
						p = "/"
					}
					return buf.WriteString(p)
				case "status":
					n := res.Status
					s := config.colorer.Green(n)
					switch {
					case n >= 60:
						s = config.colorer.Magenta(n)
					case n >= 50:
						s = config.colorer.Red(n)
					case n >= 40:
						s = config.colorer.Yellow(n)
					case n >= 30:
						s = config.colorer.Cyan(n)
					}
					return buf.WriteString(s)
				case "error":
					if err != nil {
						return buf.WriteString(err.Error())
					}
				case "latency":
					ms := float64(stop.Sub(start)) / float64(time.Millisecond)
					return buf.WriteString(strconv.FormatFloat(ms, 'f', -1, 64))
				case "latency_human":
					return buf.WriteString(stop.Sub(start).String())
				case "bytes_in":
					i := len(c.RequestURI())
					return buf.WriteString(strconv.FormatInt(int64(i), 10))
				case "bytes_out":
					return buf.WriteString(strconv.FormatInt(res.Size, 10))
				case "meta":
					return buf.WriteString(res.Meta)
				case "query":
					return buf.Write([]byte(c.QueryString()))
				}
				return 0, nil
			}); err != nil {
				return
			}

			if config.Output == nil {
				_, err = c.Logger().Output().Write(buf.Bytes())
				return
			}
			_, err = config.Output.Write(buf.Bytes())
			return
		}
	}
}
