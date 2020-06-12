package gig

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type (
	// Context represents the context of the current request. It holds connection
	// reference, path, path parameters, data and registered handler.
	Context interface {
		// Response returns `*Response`.
		Response() *Response

		// IP returns the client's network address.
		IP() string

		// Certificate returns client's certificate or nil if none provided
		Certificate() *x509.Certificate

		// URL returns the URL for the context.
		URL() *url.URL

		// Path returns the registered path for the handler.
		Path() string

		// RequestURI is the unmodified URL string as sent by the client
		// to a server. Usually the URL() or Path() should be used instead.
		RequestURI() string

		// SetPath sets the registered path for the handler.
		SetPath(p string)

		// Param returns path parameter by name.
		Param(name string) string

		// QueryString returns the URL query string.
		QueryString() string

		// Get retrieves data from the context.
		Get(key string) interface{}

		// Set saves data in the context.
		Set(key string, val interface{})

		// Render renders a template with data and sends a text/gemini response with status
		// code. Renderer must be registered using `Gig.Renderer`.
		Render(code Status, name string, data interface{}) error

		// Gemini sends a text/gemini response with status code.
		Gemini(code Status, text string) error

		// GeminiBlob sends a text/gemini blob response with status code.
		GeminiBlob(code Status, b []byte) error

		// String sends a string response with status code.
		Text(code Status, s string) error

		// Blob sends a blob response with status code and content type.
		Blob(code Status, contentType string, b []byte) error

		// Stream sends a streaming response with status code and content type.
		Stream(code Status, contentType string, r io.Reader) error

		// File sends a response with the content of the file.
		File(file string) error

		// NoContent sends a response with no body, and a status code and meta field.
		NoContent(code Status, meta string) error

		// NoContentSuccess sends a StatutSuccess response with no body.
		NoContentSuccess() error

		// Redirect redirects the request to a provided URL with status code.
		Redirect(code Status, url string) error

		// Error invokes the registered error handler. Generally used by middleware.
		Error(err error)

		// Handler returns the matched handler by router.
		Handler() HandlerFunc

		// SetHandler sets the matched handler by router.
		SetHandler(h HandlerFunc)

		// Gig returns the `Gig` instance.
		Gig() *Gig

		// Reset resets the context after request completes. It must be called along
		// with `Gig#AcquireContext()` and `Gig#ReleaseContext()`.
		// See `Gig#ServeGemini()`
		Reset(c net.Conn, u *url.URL, requestURI string, tls *tls.ConnectionState)
	}

	context struct {
		conn       net.Conn
		TLS        *tls.ConnectionState
		u          *url.URL
		response   *Response
		path       string
		requestURI string
		pnames     []string
		pvalues    []string
		handler    HandlerFunc
		store      Map
		gig        *Gig
		lock       sync.RWMutex
	}
)

const (
	indexPage     = "index.gmi"
)

func (c *context) Response() *Response {
	return c.response
}

func (c *context) IP() string {
	ra, _, _ := net.SplitHostPort(c.conn.RemoteAddr().String())
	return ra
}

func (c *context) Certificate() *x509.Certificate {
	if c.TLS == nil || len(c.TLS.PeerCertificates) == 0 {
		return nil
	}
	return c.TLS.PeerCertificates[0]
}

func (c *context) URL() *url.URL {
	return c.u
}

func (c *context) Path() string {
	return c.path
}

func (c *context) RequestURI() string {
	return c.requestURI
}

func (c *context) SetPath(p string) {
	c.path = p
}

func (c *context) Param(name string) string {
	for i, n := range c.pnames {
		if i < len(c.pvalues) {
			if n == name {
				return c.pvalues[i]
			}
		}
	}
	return ""
}

func (c *context) QueryString() string {
	return c.u.RawQuery
}

func (c *context) Get(key string) interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.store[key]
}

func (c *context) Set(key string, val interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.store == nil {
		c.store = make(Map)
	}
	c.store[key] = val
}

func (c *context) Render(code Status, name string, data interface{}) (err error) {
	if c.gig.Renderer == nil {
		return ErrRendererNotRegistered
	}
	if err = c.response.WriteHeader(code, MIMETextGeminiCharsetUTF8); err != nil {
		return
	}
	return c.gig.Renderer.Render(c.response, name, data, c)
}

func (c *context) Gemini(code Status, text string) (err error) {
	return c.GeminiBlob(code, []byte(text))
}

func (c *context) GeminiBlob(code Status, b []byte) (err error) {
	return c.Blob(code, MIMETextGeminiCharsetUTF8, b)
}

func (c *context) Text(code Status, s string) (err error) {
	return c.Blob(code, MIMETextPlainCharsetUTF8, []byte(s))
}

func (c *context) Blob(code Status, contentType string, b []byte) (err error) {
	err = c.response.WriteHeader(code, contentType)
	if err != nil {
		return
	}
	_, err = c.response.Write(b)
	return
}

func (c *context) Stream(code Status, contentType string, r io.Reader) (err error) {
	err = c.response.WriteHeader(code, contentType)
	if err != nil {
		return
	}
	_, err = io.Copy(c.response, r)
	return
}

func (c *context) File(file string) (err error) {
	s, err := os.Stat(file)
	if err != nil {
		c.Error(ErrNotFound)
		return
	}
	if uint64(s.Mode().Perm())&0444 != 0444 {
		c.Error(ErrGone)
		return
	}
	if s.IsDir() {
		files, err := ioutil.ReadDir(file)
		if err != nil {
			c.Error(ErrTemporaryFailure)
			return err
		}

		for _, f := range files {
			if f.Name() == indexPage {
				return c.File(path.Join(file, indexPage))
			}
		}
		err = c.response.WriteHeader(StatusSuccess, "text/gemini")
		if err != nil {
			return err
		}
		_, _ = c.response.Write([]byte(fmt.Sprintf("# Listing %s\n\n", file)))

		sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })

		for _, file := range files {
			if strings.HasPrefix(file.Name(), ".") {
				continue
			}

			if uint64(file.Mode().Perm())&0444 != 0444 {
				continue
			}

			_, _ = c.response.Write([]byte(fmt.Sprintf("=> %s %s [ %v ]\n", filepath.Clean(path.Join(c.path, file.Name())), file.Name(), bytefmt(file.Size()))))
		}
		return nil
	}

	ext := filepath.Ext(file)
	var mimeType string
	if ext == ".gmi" {
		mimeType = "text/gemini"
	} else {
		mimeType = mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = "octet/stream"
		}
	}

	f, err := os.OpenFile(file, os.O_RDONLY, 0600)
	if err != nil {
		c.Error(ErrTemporaryFailure)
		return
	}
	defer f.Close()

	err = c.response.WriteHeader(StatusSuccess, mimeType)
	if err != nil {
		return
	}
	_, err = io.Copy(c.response, f)
	if err != nil {
		// .. remote closed the connection, nothing we can do besides log
		// or io error, but status is already sent, everything is broken!
		c.Error(ErrTemporaryFailure)
	}
	return
}

func (c *context) NoContent(code Status, meta string) error {
	return c.response.WriteHeader(code, meta)
}

func (c *context) NoContentSuccess() error {
	return c.response.WriteHeader(StatusSuccess, "text/gemini")
}

func (c *context) Redirect(code Status, url string) error {
	if code < 30 || code >= 40 {
		return ErrInvalidRedirectCode
	}
	return c.response.WriteHeader(code, url)
}

func (c *context) Error(err error) {
	c.gig.GeminiErrorHandler(err, c)
}

func (c *context) Gig() *Gig {
	return c.gig
}

func (c *context) Handler() HandlerFunc {
	return c.handler
}

func (c *context) SetHandler(h HandlerFunc) {
	c.handler = h
}

func (c *context) Reset(conn net.Conn, u *url.URL, requestURI string, tls *tls.ConnectionState) {
	c.conn = conn
	c.TLS = tls
	c.u = u
	c.requestURI = requestURI
	c.response.reset(conn)
	c.handler = NotFoundHandler
	c.store = nil
	c.path = ""
	c.pnames = nil
	// NOTE: Don't reset because it has to have length c.gig.maxParam at all times
	for i := 0; i < *c.gig.maxParam; i++ {
		c.pvalues[i] = ""
	}
}

func bytefmt(b int64) string {
        const unit = 1000
        if b < unit {
                return fmt.Sprintf("%dB", b)
        }
        div, exp := int64(unit), 0
        for n := b / unit; n >= unit; n /= unit {
                div *= unit
                exp++
        }
        return fmt.Sprintf("%.1f%cB", float64(b)/float64(div), "kMGTPE"[exp])
}
