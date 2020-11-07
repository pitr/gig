/*
Package gig implements high performance, minimalist Go framework for Gemini protocol.

Example:

  package main

  import (
    "github.com/pitr/gig"
  )

  func main() {
    // Gig instance
    g := gig.Default()

    // Routes
    g.Handle("/user/:name", func(c gig.Context) error {
        return c.Gemini(gig.StatusSuccess, "# Hello, %s!", c.Param("name"))
    })

    // Start server
    g.Run("my.crt", "my.key")
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
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"time"
)

type (
	// Gig is the top-level framework instance.
	Gig struct {
		common

		premiddleware []MiddlewareFunc
		middleware    []MiddlewareFunc
		maxParam      *int
		router        *router
		listener      net.Listener
		addr          string
		pool          sync.Pool
		doneChan      chan struct{}
		closeOnce     sync.Once
		mu            sync.Mutex

		// HideBanner disables banner on startup.
		HideBanner bool
		// HidePort disables startup message.
		HidePort bool
		// GeminiErrorHandler allows setting custom error handler
		GeminiErrorHandler GeminiErrorHandler
		// Renderer must be set for Context#Render to work
		Renderer Renderer
		// ReadTimeout set max read timeout on socket.
		// Default is none.
		ReadTimeout time.Duration
		// WriteTimeout set max write timeout on socket.
		// Default is none.
		WriteTimeout time.Duration
		// TLSConfig is passed to tls.NewListener and needs to be modified
		// before Run is called.
		TLSConfig *tls.Config
	}

	// Route contains a handler and information for matching against requests.
	Route struct {
		Path string
		Name string
	}

	// GeminiError represents an error that occurred while handling a request.
	GeminiError struct {
		Code    Status
		Message string
	}

	// MiddlewareFunc defines a function to process middleware.
	MiddlewareFunc func(HandlerFunc) HandlerFunc

	// HandlerFunc defines a function to serve requests.
	HandlerFunc func(Context) error

	// GeminiErrorHandler is a centralized error handler.
	GeminiErrorHandler func(error, Context)

	// Renderer is the interface that wraps the Render function.
	Renderer interface {
		Render(io.Writer, string, interface{}, Context) error
	}

	storeMap map[string]interface{}

	// Common struct for Gig & Group.
	common struct{}
)

// MIME types.
const (
	MIMETextGemini            = "text/gemini"
	MIMETextGeminiCharsetUTF8 = "text/gemini; charset=UTF-8"
	MIMETextPlain             = "text/plain"
	MIMETextPlainCharsetUTF8  = "text/plain; charset=UTF-8"
)

const (
	// Version of Gig
	Version = "0.9.7"
	// http://patorjk.com/software/taag/#p=display&f=Small%20Slant&t=gig
	banner = `
        _
  ___ _(_)__ _
 / _  / / _  /
 \_, /_/\_, /
/___/  /___/   %s

`
)

// Errors that can be inherited from using NewErrorFrom.
var (
	ErrTemporaryFailure          = NewError(StatusTemporaryFailure, "Temporary Failure")
	ErrServerUnavailable         = NewError(StatusServerUnavailable, "Server Unavailable")
	ErrCGIError                  = NewError(StatusCGIError, "CGI Error")
	ErrProxyError                = NewError(StatusProxyError, "Proxy Error")
	ErrSlowDown                  = NewError(StatusSlowDown, "Slow Down")
	ErrPermanentFailure          = NewError(StatusPermanentFailure, "Permanent Failure")
	ErrNotFound                  = NewError(StatusNotFound, "Not Found")
	ErrGone                      = NewError(StatusGone, "Gone")
	ErrProxyRequestRefused       = NewError(StatusProxyRequestRefused, "Proxy Request Refused")
	ErrBadRequest                = NewError(StatusBadRequest, "Bad Request")
	ErrClientCertificateRequired = NewError(StatusClientCertificateRequired, "Client Certificate Required")
	ErrCertificateNotAuthorised  = NewError(StatusCertificateNotAuthorised, "Certificate Not Authorised")
	ErrCertificateNotValid       = NewError(StatusCertificateNotValid, "Certificate Not Valid")

	ErrRendererNotRegistered = errors.New("renderer not registered")
	ErrInvalidCertOrKeyType  = errors.New("invalid cert or key type, must be string or []byte")

	ErrServerClosed = errors.New("gemini: Server closed")
)

// Error handlers.
var (
	NotFoundHandler = func(c Context) error {
		return ErrNotFound
	}
)

// DefaultGeminiErrorHandler is the default HTTP error handler. It sends a JSON response
// with status code.
func DefaultGeminiErrorHandler(err error, c Context) {
	he, ok := err.(*GeminiError)
	if !ok {
		he = &GeminiError{
			Code:    StatusPermanentFailure,
			Message: err.Error(),
		}
	}

	code := he.Code
	message := he.Message

	debugPrintf("%s", err)

	// Send response
	if !c.Response().Committed {
		err = c.NoContent(code, message)
		if err != nil {
			debugPrintf("%s", err)
		}
	}
}

// New creates an instance of Gig.
func New() *Gig {
	g := &Gig{
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			ClientAuth: tls.RequestClientCert,
		},
		maxParam: new(int),
		doneChan: make(chan struct{}),
	}
	g.GeminiErrorHandler = DefaultGeminiErrorHandler
	g.pool.New = func() interface{} {
		return g.newContext(nil, nil, "", nil)
	}
	g.router = newRouter(g)

	return g
}

// Default returns a Gig instance with Logger and Recover middleware enabled.
func Default() *Gig {
	g := New()

	// Default middlewares
	g.Use(Logger(), Recover())

	return g
}

func (g *Gig) newContext(c net.Conn, u *url.URL, requestURI string, tls *tls.ConnectionState) Context {
	return &context{
		conn:       c,
		TLS:        tls,
		u:          u,
		requestURI: requestURI,
		response:   NewResponse(c),
		store:      make(storeMap),
		gig:        g,
		pvalues:    make([]string, *g.maxParam),
		handler:    NotFoundHandler,
	}
}

// Pre adds middleware to the chain which is run before router.
func (g *Gig) Pre(middleware ...MiddlewareFunc) {
	g.premiddleware = append(g.premiddleware, middleware...)
}

// Use adds middleware to the chain which is run after router.
func (g *Gig) Use(middleware ...MiddlewareFunc) {
	g.middleware = append(g.middleware, middleware...)
}

// Handle registers a new route for a path with matching handler in the router
// with optional route-level middleware.
func (g *Gig) Handle(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.add(path, h, m...)
}

// Static registers a new route with path prefix to serve static files from the
// provided root directory.
func (g *Gig) Static(prefix, root string) *Route {
	if root == "" {
		root = "." // For security we want to restrict to CWD.
	}

	return g.static(prefix, root, g.Handle)
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
func (g *Gig) File(path, file string, m ...MiddlewareFunc) *Route {
	return g.file(path, file, g.Handle, m...)
}

func (g *Gig) add(path string, handler HandlerFunc, middleware ...MiddlewareFunc) *Route {
	name := handlerName(handler)

	g.router.add(path, func(c Context) error {
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

	g.router.routes[path] = r

	return r
}

// Group creates a new router group with prefix and optional group-level middleware.
func (g *Gig) Group(prefix string, m ...MiddlewareFunc) (gg *Group) {
	gg = &Group{prefix: prefix, gig: g}
	gg.Use(m...)

	return
}

// URL generates a URL from handler.
func (g *Gig) URL(handler HandlerFunc, params ...interface{}) string {
	name := handlerName(handler)
	return g.Reverse(name, params...)
}

// Reverse generates an URL from route name and provided parameters.
func (g *Gig) Reverse(name string, params ...interface{}) string {
	uri := new(bytes.Buffer)
	ln := len(params)
	n := 0

	for _, r := range g.router.routes {
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
func (g *Gig) Routes() []*Route {
	routes := make([]*Route, 0, len(g.router.routes))
	for _, v := range g.router.routes {
		routes = append(routes, v)
	}

	return routes
}

// ServeGemini serves Gemini request.
func (g *Gig) ServeGemini(c Context) {
	var h HandlerFunc

	URL := c.URL()

	if g.premiddleware == nil {
		g.router.find(getPath(URL), c)
		h = c.Handler()
		h = applyMiddleware(h, g.middleware...)
	} else {
		h = func(c Context) error {
			g.router.find(getPath(URL), c)
			h := c.Handler()
			h = applyMiddleware(h, g.middleware...)
			return h(c)
		}
		h = applyMiddleware(h, g.premiddleware...)
	}

	// Execute chain
	if err := h(c); err != nil {
		g.GeminiErrorHandler(err, c)
	}
}

// Run starts a Gemini server.
// If `certFile` or `keyFile` is `string` the values are treated as file paths.
// If `certFile` or `keyFile` is `[]byte` the values are treated as the certificate or key as-is.
func (g *Gig) Run(args ...interface{}) (err error) {
	var (
		cert, key         []byte
		certFile, keyFile interface{}
		addr              string
	)

	switch len(args) {
	case 2:
		addr, certFile, keyFile = os.Getenv("PORT"), args[0], args[1]
		if addr == "" {
			addr = ":1965"
		} else {
			addr = ":" + addr
		}
	case 3:
		addr, certFile, keyFile = args[0].(string), args[1], args[2]
	default:
		panic("must specify 2 or 3 arguments to Run")
	}

	if cert, err = filepathOrContent(certFile); err != nil {
		return
	}

	if key, err = filepathOrContent(keyFile); err != nil {
		return
	}

	g.TLSConfig.Certificates = make([]tls.Certificate, 1)

	if g.TLSConfig.Certificates[0], err = tls.X509KeyPair(cert, key); err != nil {
		return
	}

	return g.startTLS(addr)
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

func (g *Gig) startTLS(address string) error {
	g.addr = address

	// Setup
	if !g.HideBanner {
		debugPrintf(banner, "v"+Version)
	}

	g.mu.Lock()
	if g.listener == nil {
		l, err := newListener(g.addr)
		if err != nil {
			return err
		}

		g.listener = tls.NewListener(l, g.TLSConfig)
	}
	g.mu.Unlock()

	defer g.listener.Close()

	if !g.HidePort {
		debugPrintf("⇨ gemini server started on %s\n", g.listener.Addr())
	}

	return g.serve()
}

func (g *Gig) serve() error {
	var tempDelay time.Duration // how long to sleep on accept failure

	for {
		conn, err := g.listener.Accept()
		if err != nil {
			select {
			case <-g.doneChan:
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

				debugPrintf("gemini: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)

				continue
			}

			return err
		}

		tc, ok := conn.(*tls.Conn)
		if !ok {
			debugPrintf("gemini: non-tls connection")
			continue
		}

		go g.handleRequest(tc)
	}
}

func (g *Gig) handleRequest(conn *tls.Conn) {
	defer conn.Close()

	if d := g.ReadTimeout; d != 0 {
		err := conn.SetReadDeadline(time.Now().Add(d))
		if err != nil {
			debugPrintf("%s", err)
		}
	}

	reader := bufio.NewReaderSize(conn, 1024)
	request, overflow, err := reader.ReadLine()

	if overflow {
		debugPrintf("gemini: request overflow")

		_, _ = conn.Write([]byte(fmt.Sprintf("%d %s\r\n", StatusBadRequest, "Request too long!")))

		return
	} else if err != nil {
		debugPrintf("gemini: %s", err)

		_, _ = conn.Write([]byte(fmt.Sprintf("%d %s\r\n", StatusBadRequest, "Unknown error reading request!")))

		return
	}

	RequestURI := string(request)
	URL, err := url.Parse(RequestURI)

	if err != nil {
		debugPrintf("gemini: %s", err)

		_, _ = conn.Write([]byte(fmt.Sprintf("%d %s\r\n", StatusBadRequest, "Error parsing URL!")))

		return
	}

	if URL.Scheme == "" {
		URL.Scheme = "gemini"
	}

	if URL.Scheme != "gemini" {
		debugPrintf("gemini: non-gemini scheme: %s", RequestURI)

		_, _ = conn.Write([]byte(fmt.Sprintf("%d %s\r\n", StatusBadRequest, "No proxying to non-Gemini content!")))

		return
	}

	if d := g.WriteTimeout; d != 0 {
		err := conn.SetWriteDeadline(time.Now().Add(d))
		if err != nil {
			debugPrintf("%s", err)
		}
	}

	tlsState := new(tls.ConnectionState)
	*tlsState = conn.ConnectionState()

	// Acquire context
	c := g.pool.Get().(*context)
	c.reset(conn, URL, RequestURI, tlsState)

	g.ServeGemini(c)

	// Release context
	g.pool.Put(c)
}

// Close immediately stops the server.
// It internally calls `net.Listener#Close()`.
func (g *Gig) Close() error {
	g.closeOnce.Do(func() {
		close(g.doneChan)
	})
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.listener != nil {
		return g.listener.Close()
	}

	return nil
}

// NewError creates a new GeminiError instance.
func NewError(code Status, message string) *GeminiError {
	return &GeminiError{Code: code, Message: message}
}

// NewErrorFrom creates a new GeminiError instance using Code from existing GeminiError.
func NewErrorFrom(err *GeminiError, message string) *GeminiError {
	return &GeminiError{Code: err.Code, Message: message}
}

// Error makes it compatible with `error` interface.
func (ge *GeminiError) Error() string {
	return fmt.Sprintf("error=%s", ge.Message)
}

// getPath returns RawPath, if it's empty returns Path from URL.
func getPath(u *url.URL) string {
	path := u.RawPath
	if path == "" {
		path = u.Path
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
// connections. It's used by Run so dead TCP connections (e.g.
// closing laptop mid-download) eventually go away.
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
