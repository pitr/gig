package gig

import (
	"crypto/md5"
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
	// DO NOT retain Context instance, as it will be reused by other connections.
	Context interface {
		// Response returns `*Response`.
		Response() *Response

		// IP returns the client's network address.
		IP() string

		// Certificate returns client's leaf certificate or nil if none provided
		Certificate() *x509.Certificate

		// CertHash returns a hash of client's leaf certificate or empty string is none
		CertHash() string

		// URL returns the URL for the context.
		URL() *url.URL

		// Path returns the registered path for the handler.
		Path() string

		// QueryString returns unescaped URL query string or error if the raw query
		// could not be unescaped. Use Context#URL().RawQuery to get raw query string.
		QueryString() (string, error)

		// RequestURI is the unmodified URL string as sent by the client
		// to a server. Usually the URL() or Path() should be used instead.
		RequestURI() string

		// Reader returns request connection reader that can be
		// used to read data that client sent after Gemini request.
		// This feature is necessary to implement Gemini extensions
		// like Titan protocol. Spec compliant Gemini server should
		// ignore all data after the request.
		Reader() io.Reader

		// Param returns path parameter by name.
		Param(name string) string

		// Get retrieves data from the context.
		Get(key string) interface{}

		// Set saves data in the context.
		Set(key string, val interface{})

		// Render renders a template with data and sends a text/gemini response with status
		// code Success. Renderer must be registered using `Gig.Renderer`.
		Render(name string, data interface{}) error

		// Gemini sends a text/gemini response with status code Success.
		Gemini(text string, args ...interface{}) error

		// GeminiBlob sends a text/gemini blob response with status code Success.
		GeminiBlob(b []byte) error

		// Text sends a text/plain response with status code Success.
		Text(format string, values ...interface{}) error

		// Blob sends a blob response with status code Success and content type.
		Blob(contentType string, b []byte) error

		// Stream sends a streaming response with status code Success and content type.
		Stream(contentType string, r io.Reader) error

		// File sends a response with the content of the file.
		File(file string) error

		// NoContent sends a response with no body, and a status code and meta field.
		// Use for any non-2x status codes
		NoContent(code Status, meta string, values ...interface{}) error

		// Error invokes the registered error handler. Generally used by middleware.
		Error(err error)

		// Handler returns the matched handler by router.
		Handler() HandlerFunc

		// Gig returns the `Gig` instance.
		Gig() *Gig
	}

	context struct {
		conn       tlsconn
		TLS        *tls.ConnectionState
		u          *url.URL
		reader     io.Reader
		response   *Response
		path       string
		requestURI string
		pnames     []string
		pvalues    []string
		handler    HandlerFunc
		store      storeMap
		gig        *Gig
		lock       sync.RWMutex
	}
)

const (
	indexPage = "index.gmi"
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

func (c *context) CertHash() string {
	cert := c.Certificate()
	if cert == nil {
		return ""
	}

	return fmt.Sprintf("%x", md5.Sum(cert.Raw))
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

func (c *context) QueryString() (string, error) {
	return url.QueryUnescape(c.u.RawQuery)
}

func (c *context) Reader() io.Reader {
	return c.reader
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
		c.store = make(storeMap)
	}

	c.store[key] = val
}

func (c *context) Render(name string, data interface{}) (err error) {
	if c.gig.Renderer == nil {
		return ErrRendererNotRegistered
	}

	if err = c.response.WriteHeader(StatusSuccess, MIMETextGemini); err != nil {
		return
	}

	return c.gig.Renderer.Render(c.response, name, data, c)
}

func (c *context) Gemini(format string, values ...interface{}) error {
	return c.GeminiBlob([]byte(fmt.Sprintf(format, values...)))
}

func (c *context) GeminiBlob(b []byte) (err error) {
	return c.Blob(MIMETextGemini, b)
}

func (c *context) Text(format string, values ...interface{}) (err error) {
	return c.Blob(MIMETextPlain, []byte(fmt.Sprintf(format, values...)))
}

func (c *context) Blob(contentType string, b []byte) (err error) {
	err = c.response.WriteHeader(StatusSuccess, contentType)
	if err != nil {
		return
	}

	_, err = c.response.Write(b)

	return
}

func (c *context) Stream(contentType string, r io.Reader) (err error) {
	err = c.response.WriteHeader(StatusSuccess, contentType)
	if err != nil {
		return
	}

	_, err = io.Copy(c.response, r)

	return
}

func (c *context) File(file string) (err error) {
	if containsDotDot(file) {
		c.Error(ErrBadRequest)
		return
	}

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

		_, _ = c.response.Write([]byte(fmt.Sprintf("# Listing %s\n\n", c.u.Path)))

		sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })

		for _, file := range files {
			if strings.HasPrefix(file.Name(), ".") {
				continue
			}

			if uint64(file.Mode().Perm())&0444 != 0444 {
				continue
			}

			_, _ = c.response.Write([]byte(fmt.Sprintf("=> %s %s [ %v ]\n", filepath.Clean(path.Join(c.u.Path, file.Name())), file.Name(), bytefmt(file.Size()))))
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

func containsDotDot(v string) bool {
	if !strings.Contains(v, "..") {
		return false
	}

	for _, ent := range strings.FieldsFunc(v, isSlashRune) {
		if ent == ".." {
			return true
		}
	}

	return false
}

func isSlashRune(r rune) bool { return r == '/' || r == '\\' }

func (c *context) NoContent(code Status, meta string, values ...interface{}) error {
	return c.response.WriteHeader(code, fmt.Sprintf(meta, values...))
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

func (c *context) reset(conn tlsconn, u *url.URL, requestURI string, reader io.Reader, tls *tls.ConnectionState) {
	c.conn = conn
	c.TLS = tls
	c.u = u
	c.requestURI = requestURI
	c.response.reset(conn)
	c.reader = reader
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
