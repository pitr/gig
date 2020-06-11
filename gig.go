/*
Package gig implements high performance, minimalist Go framework for Gemini protocol.

Example:

  package main

  import (
    "github.com/pitr/gig"
    "github.com/pitr/gig/middleware"
  )

  // Handler
  func hello(c gig.Context) error {
    return c.String(gig.StatusSuccess, "Hello, World!")
  }

  func main() {
    // Gig instance
    e := gig.New()

    // Middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())

    // Routes
    e.GET("/", hello)

    // Start server
    e.Logger.Fatal(e.StartAutoTLS(":1323"))
  }
*/
package gig

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/labstack/gommon/color"
	"github.com/labstack/gommon/log"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

type (
	// Gig is the top-level framework instance.
	Gig struct {
		common
		colorer            *color.Color
		premiddleware      []MiddlewareFunc
		middleware         []MiddlewareFunc
		maxParam           *int
		router             *Router
		pool               sync.Pool
		ReadTimeout        time.Duration
		WriteTimeout       time.Duration
		TLSConfig          *tls.Config
		Addr               string
		Listener           net.Listener
		doneChan           chan struct{}
		closeOnce          sync.Once
		mu                 sync.Mutex
		AutoTLSManager     autocert.Manager
		Debug              bool
		HideBanner         bool
		HidePort           bool
		GeminiErrorHandler GeminiErrorHandler
		Validator          Validator
		Renderer           Renderer
		Logger             Logger
	}

	// Route contains a handler and information for matching against requests.
	Route struct {
		Path string `json:"path"`
		Name string `json:"name"`
	}

	// GeminiError represents an error that occurred while handling a request.
	GeminiError struct {
		Code     Status
		Message  string
		Internal error
	}

	// MiddlewareFunc defines a function to process middleware.
	MiddlewareFunc func(HandlerFunc) HandlerFunc

	// HandlerFunc defines a function to serve requests.
	HandlerFunc func(Context) error

	// GeminiErrorHandler is a centralized error handler.
	GeminiErrorHandler func(error, Context)

	// Validator is the interface that wraps the Validate function.
	Validator interface {
		Validate(i interface{}) error
	}

	// Renderer is the interface that wraps the Render function.
	Renderer interface {
		Render(io.Writer, string, interface{}, Context) error
	}

	// Map defines a generic map of type `map[string]interface{}`.
	Map map[string]interface{}

	// Common struct for Gig & Group.
	common struct{}
)

// MIME types
const (
	MIMEApplicationJSON                  = "application/json"
	MIMEApplicationJSONCharsetUTF8       = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEApplicationJavaScript            = "application/javascript"
	MIMEApplicationJavaScriptCharsetUTF8 = MIMEApplicationJavaScript + "; " + charsetUTF8
	MIMEApplicationXML                   = "application/xml"
	MIMEApplicationXMLCharsetUTF8        = MIMEApplicationXML + "; " + charsetUTF8
	MIMETextXML                          = "text/xml"
	MIMETextXMLCharsetUTF8               = MIMETextXML + "; " + charsetUTF8
	MIMETextGemini                       = "text/gemini"
	MIMETextGeminiCharsetUTF8            = MIMETextGemini + "; " + charsetUTF8
	MIMETextPlain                        = "text/plain"
	MIMETextPlainCharsetUTF8             = MIMETextPlain + "; " + charsetUTF8
	MIMEMultipartForm                    = "multipart/form-data"
)

const (
	charsetUTF8 = "charset=UTF-8"
)

const (
	// Version of Gig
	Version = "1.0.0"
	// http://patorjk.com/software/taag/#p=display&f=Small%20Slant&t=gig
	banner = `
        _
  ___ _(_)__ _
 / _  / / _  /
 \_, /_/\_, /
/___/  /___/   %s

`
)

// Errors
var (
	ErrTemporaryFailure              = NewGeminiError(StatusTemporaryFailure, "Temporary Failure")
	ErrServerUnavailable             = NewGeminiError(StatusServerUnavailable, "Server Unavailable")
	ErrCGIError                      = NewGeminiError(StatusCGIError, "CGI Error")
	ErrProxyError                    = NewGeminiError(StatusProxyError, "Proxy Error")
	ErrSlowDown                      = NewGeminiError(StatusSlowDown, "Slow Down")
	ErrPermanentFailure              = NewGeminiError(StatusPermanentFailure, "Permanent Failure")
	ErrNotFound                      = NewGeminiError(StatusNotFound, "Not Found")
	ErrGone                          = NewGeminiError(StatusGone, "Gone")
	ErrProxyRequestRefused           = NewGeminiError(StatusProxyRequestRefused, "Proxy Request Refused")
	ErrBadRequest                    = NewGeminiError(StatusBadRequest, "Bad Request")
	ErrClientCertificateRequired     = NewGeminiError(StatusClientCertificateRequired, "Client Certificate Required")
	ErrTransientCertificateRequested = NewGeminiError(StatusTransientCertificateRequested, "Transient Certificate Requested")
	ErrAuthorisedCertificateRequired = NewGeminiError(StatusAuthorisedCertificateRequired, "Authorised Certificate Required")
	ErrCertificateNotAccepted        = NewGeminiError(StatusCertificateNotAccepted, "Certificate Not Accepted")
	ErrFutureCertificateRejected     = NewGeminiError(StatusFutureCertificateRejected, "Future Certificate Rejected")
	ErrExpiredCertificateRejected    = NewGeminiError(StatusExpiredCertificateRejected, "Expired Certificate Rejected")

	ErrRendererNotRegistered = errors.New("renderer not registered")
	ErrInvalidRedirectCode   = errors.New("invalid redirect status code")
	ErrInvalidCertOrKeyType  = errors.New("invalid cert or key type, must be string or []byte")

	ErrServerClosed = errors.New("gemini: Server closed")
)

// Error handlers
var (
	NotFoundHandler = func(c Context) error {
		return ErrNotFound
	}
)

// New creates an instance of Gig.
func New() (e *Gig) {
	e = &Gig{
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		AutoTLSManager: autocert.Manager{
			Prompt: autocert.AcceptTOS,
		},
		Logger:   log.New("gig"),
		colorer:  color.New(),
		maxParam: new(int),
		doneChan: make(chan struct{}),
	}
	e.GeminiErrorHandler = e.DefaultGeminiErrorHandler
	e.Logger.SetLevel(log.ERROR)
	e.pool.New = func() interface{} {
		return e.NewContext(nil, nil, "", nil)
	}
	e.router = NewRouter(e)
	return
}

// NewContext returns a Context instance.
func (e *Gig) NewContext(c net.Conn, u *url.URL, requestURI string, tls *tls.ConnectionState) Context {
	return &context{
		conn:       c,
		TLS:        tls,
		u:          u,
		requestURI: requestURI,
		response:   NewResponse(c, e.Logger),
		store:      make(Map),
		gig:        e,
		pvalues:    make([]string, *e.maxParam),
		handler:    NotFoundHandler,
	}
}

// Router returns the default router.
func (e *Gig) Router() *Router {
	return e.router
}

// DefaultGeminiErrorHandler is the default HTTP error handler. It sends a JSON response
// with status code.
func (e *Gig) DefaultGeminiErrorHandler(err error, c Context) {
	he, ok := err.(*GeminiError)
	if !ok {
		he = &GeminiError{
			Code:    StatusPermanentFailure,
			Message: err.Error(),
		}
	}

	code := he.Code
	message := he.Message
	if e.Debug {
		message = err.Error()
	}

	// Send response
	if !c.Response().Committed {
		err = c.NoContent(code, message)
		if err != nil {
			e.Logger.Error(err)
		}
	}
}

// Pre adds middleware to the chain which is run before router.
func (e *Gig) Pre(middleware ...MiddlewareFunc) {
	e.premiddleware = append(e.premiddleware, middleware...)
}

// Use adds middleware to the chain which is run after router.
func (e *Gig) Use(middleware ...MiddlewareFunc) {
	e.middleware = append(e.middleware, middleware...)
}

// Handle registers a new route for a path with matching handler in the router
// with optional route-level middleware.
func (e *Gig) Handle(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.add(path, h, m...)
}

// Static registers a new route with path prefix to serve static files from the
// provided root directory.
func (e *Gig) Static(prefix, root string) *Route {
	if root == "" {
		root = "." // For security we want to restrict to CWD.
	}
	return e.static(prefix, root, e.Handle)
}

func (common) static(prefix, root string, get func(string, HandlerFunc, ...MiddlewareFunc) *Route) *Route {
	h := func(c Context) error {
		p, err := url.PathUnescape(c.Param("*"))
		if err != nil {
			return err
		}
		name := filepath.Join(root, path.Clean("/"+p)) // "/"+ for security
		return c.File(name)
	}
	if prefix == "/" {
		return get(prefix+"*", h)
	}
	return get(prefix+"/*", h)
}

func (common) file(path, file string, get func(string, HandlerFunc, ...MiddlewareFunc) *Route,
	m ...MiddlewareFunc) *Route {
	return get(path, func(c Context) error {
		return c.File(file)
	}, m...)
}

// File registers a new route with path to serve a static file with optional route-level middleware.
func (e *Gig) File(path, file string, m ...MiddlewareFunc) *Route {
	return e.file(path, file, e.Handle, m...)
}

func (e *Gig) add(path string, handler HandlerFunc, middleware ...MiddlewareFunc) *Route {
	name := handlerName(handler)
	e.router.Add(path, func(c Context) error {
		h := handler
		// Chain middleware
		for i := len(middleware) - 1; i >= 0; i-- {
			h = middleware[i](h)
		}
		return h(c)
	})
	r := &Route{
		Path: path,
		Name: name,
	}
	e.router.routes[path] = r
	return r
}

// Group creates a new router group with prefix and optional group-level middleware.
func (e *Gig) Group(prefix string, m ...MiddlewareFunc) (g *Group) {
	g = &Group{prefix: prefix, gig: e}
	g.Use(m...)
	return
}

// URL generates a URL from handler.
func (e *Gig) URL(handler HandlerFunc, params ...interface{}) string {
	name := handlerName(handler)
	return e.Reverse(name, params...)
}

// Reverse generates an URL from route name and provided parameters.
func (e *Gig) Reverse(name string, params ...interface{}) string {
	uri := new(bytes.Buffer)
	ln := len(params)
	n := 0
	for _, r := range e.router.routes {
		if r.Name == name {
			for i, l := 0, len(r.Path); i < l; i++ {
				if r.Path[i] == ':' && n < ln {
					for ; i < l && r.Path[i] != '/'; i++ {
					}
					uri.WriteString(fmt.Sprintf("%v", params[n]))
					n++
				}
				if i < l {
					uri.WriteByte(r.Path[i])
				}
			}
			break
		}
	}
	return uri.String()
}

// Routes returns the registered routes.
func (e *Gig) Routes() []*Route {
	routes := make([]*Route, 0, len(e.router.routes))
	for _, v := range e.router.routes {
		routes = append(routes, v)
	}
	return routes
}

// AcquireContext returns an empty `Context` instance from the pool.
// You must return the context by calling `ReleaseContext()`.
func (e *Gig) AcquireContext() Context {
	return e.pool.Get().(Context)
}

// ReleaseContext returns the `Context` instance back to the pool.
// You must call it after `AcquireContext()`.
func (e *Gig) ReleaseContext(c Context) {
	e.pool.Put(c)
}

// ServeGemini serves Gemini request
func (e *Gig) ServeGemini(c Context) {
	var h HandlerFunc

	URL := c.URL()

	if e.premiddleware == nil {
		e.router.Find(GetPath(URL), c)
		h = c.Handler()
		h = applyMiddleware(h, e.middleware...)
	} else {
		h = func(c Context) error {
			e.router.Find(GetPath(URL), c)
			h := c.Handler()
			h = applyMiddleware(h, e.middleware...)
			return h(c)
		}
		h = applyMiddleware(h, e.premiddleware...)
	}

	// Execute chain
	if err := h(c); err != nil {
		e.GeminiErrorHandler(err, c)
	}
}

// StartTLS starts a Gemini server.
// If `certFile` or `keyFile` is `string` the values are treated as file paths.
// If `certFile` or `keyFile` is `[]byte` the values are treated as the certificate or key as-is.
func (e *Gig) StartTLS(address string, certFile, keyFile interface{}) (err error) {
	var cert []byte
	if cert, err = filepathOrContent(certFile); err != nil {
		return
	}

	var key []byte
	if key, err = filepathOrContent(keyFile); err != nil {
		return
	}

	e.TLSConfig.Certificates = make([]tls.Certificate, 1)
	if e.TLSConfig.Certificates[0], err = tls.X509KeyPair(cert, key); err != nil {
		return
	}

	return e.startTLS(address)
}

func filepathOrContent(fileOrContent interface{}) (content []byte, err error) {
	switch v := fileOrContent.(type) {
	case string:
		return ioutil.ReadFile(v)
	case []byte:
		return v, nil
	default:
		return nil, ErrInvalidCertOrKeyType
	}
}

// StartAutoTLS starts a Gemini server using certificates automatically installed from https://letsencrypt.org.
func (e *Gig) StartAutoTLS(address string) error {
	e.TLSConfig.GetCertificate = e.AutoTLSManager.GetCertificate
	e.TLSConfig.NextProtos = append(e.TLSConfig.NextProtos, acme.ALPNProto)
	return e.startTLS(address)
}

func (e *Gig) startTLS(address string) error {
	e.Addr = address

	// Setup
	e.colorer.SetOutput(e.Logger.Output())
	if e.Debug {
		e.Logger.SetLevel(log.DEBUG)
	}

	if !e.HideBanner {
		e.colorer.Printf(banner, e.colorer.Red("v"+Version))
	}

	e.mu.Lock()
	if e.Listener == nil {
		l, err := newListener(e.Addr)
		if err != nil {
			return err
		}
		e.Listener = tls.NewListener(l, e.TLSConfig)
	}
	e.mu.Unlock()
	defer e.Listener.Close()

	if !e.HidePort {
		e.colorer.Printf("⇨ gemini server started on %s\n", e.colorer.Green(e.Listener.Addr()))
	}
	return e.serve()
}

func (e *Gig) serve() error {
	var tempDelay time.Duration // how long to sleep on accept failure

	for {
		conn, err := e.Listener.Accept()
		if err != nil {
			select {
			case <-e.doneChan:
				return ErrServerClosed
			default:
			}

			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				e.Logger.Errorf("gemini: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}

		tc, ok := conn.(*tls.Conn)
		if !ok {
			e.Logger.Errorf("gemini: non-tls connection")
			continue
		}

		go e.handleRequest(tc)
	}
}

func (e *Gig) handleRequest(conn *tls.Conn) {
	defer conn.Close()

	if d := e.ReadTimeout; d != 0 {
		err := conn.SetReadDeadline(time.Now().Add(d))
		if err != nil {
			e.Logger.Error(err)
		}
	}

	reader := bufio.NewReaderSize(conn, 1024)
	request, overflow, err := reader.ReadLine()
	if overflow {
		_ = NewResponse(conn, e.Logger).WriteHeader(StatusBadRequest, "Request too long!")
		return
	} else if err != nil {
		_ = NewResponse(conn, e.Logger).WriteHeader(StatusBadRequest, "Unknown error reading request! "+err.Error())
		return
	}

	RequestURI := string(request)
	URL, err := url.Parse(RequestURI)
	if err != nil {
		_ = NewResponse(conn, e.Logger).WriteHeader(StatusBadRequest, "Error parsing URL!")
		return
	}
	if URL.Scheme == "" {
		URL.Scheme = "gemini"
	}

	if URL.Scheme != "gemini" {
		_ = NewResponse(conn, e.Logger).WriteHeader(StatusBadRequest, "No proxying to non-Gemini content!")
		return
	}

	if d := e.WriteTimeout; d != 0 {
		err := conn.SetWriteDeadline(time.Now().Add(d))
		if err != nil {
			e.Logger.Error(err)
		}
	}

	tlsState := new(tls.ConnectionState)
	*tlsState = conn.ConnectionState()

	// Acquire context
	c := e.pool.Get().(*context)
	c.Reset(conn, URL, RequestURI, tlsState)

	e.ServeGemini(c)

	// Release context
	e.pool.Put(c)
}

// Close immediately stops the server.
// It internally calls `net.Listener#Close()`.
func (e *Gig) Close() error {
	e.closeOnce.Do(func() {
		close(e.doneChan)
	})
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.Listener != nil {
		return e.Listener.Close()
	}
	return nil
}

// NewGeminiError creates a new GeminiError instance.
func NewGeminiError(code Status, message string) *GeminiError {
	return &GeminiError{Code: code, Message: message}
}

// NewGeminiErrorFrom creates a new GeminiError instance using Code from existing GeminiError.
func NewGeminiErrorFrom(err *GeminiError, message string) *GeminiError {
	return &GeminiError{Code: err.Code, Message: message}
}

// Error makes it compatible with `error` interface.
func (ge *GeminiError) Error() string {
	return fmt.Sprintf("code=%d, message=%v", ge.Code, ge.Message)
}

// GetPath returns RawPath, if it's empty returns Path from URL
func GetPath(URL *url.URL) string {
	path := URL.RawPath
	if path == "" {
		path = URL.Path
	}
	return path
}

func handlerName(h HandlerFunc) string {
	t := reflect.ValueOf(h).Type()
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	}
	return t.String()
}

// // PathUnescape is wraps `url.PathUnescape`
// func PathUnescape(s string) (string, error) {
// 	return url.PathUnescape(s)
// }

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	if c, err = ln.AcceptTCP(); err != nil {
		return
	} else if err = c.(*net.TCPConn).SetKeepAlive(true); err != nil {
		return
	}
	// Ignore error from setting the KeepAlivePeriod as some systems, such as
	// OpenBSD, do not support setting TCP_USER_TIMEOUT on IPPROTO_TCP
	_ = c.(*net.TCPConn).SetKeepAlivePeriod(3 * time.Minute)
	return
}

func newListener(address string) (*tcpKeepAliveListener, error) {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return &tcpKeepAliveListener{l.(*net.TCPListener)}, nil
}

func applyMiddleware(h HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}
