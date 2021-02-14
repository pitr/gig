# Gig - Gemini framework

[![Used By](https://img.shields.io/badge/used%20by-5%2B%20projects-brightgreen)](#who-uses-gig)
[![godocs.io](https://godocs.io/github.com/pitr/gig?status.svg)](https://godocs.io/github.com/pitr/gig)
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/pitr/gig)
[![Go Report Card](https://goreportcard.com/badge/github.com/pitr/gig?style=flat-square)](https://goreportcard.com/report/github.com/pitr/gig)
[![Codecov](https://img.shields.io/codecov/c/github/pitr/gig.svg?style=flat-square)](https://codecov.io/gh/pitr/gig)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/pitr/gig/master/LICENSE)

API is subject to change until v1.0

## Protocol compatibility

| Version | Supported Gemini version |
| ------- | ------------------------ |
| 0.9.4   | v0.14.*                  |
| < 0.9.4 | v0.13.*                  |

## Contents

* [Feature Overview](#feature-overview)
* [Guide](#guide)
   * [Quick Start](#quick-start)
   * [Parameters in path](#parameters-in-path)
   * [Query](#query)
   * [Client Certificate](#client-certificate)
   * [Grouping routes](#grouping-routes)
   * [Blank Gig without middleware by default](#blank-gig-without-middleware-by-default)
   * [Using middleware](#using-middleware)
   * [Writing logs to file](#writing-logs-to-file)
   * [Custom Log Format](#custom-log-format)
   * [Serving static files](#serving-static-files)
   * [Serving data from file](#serving-data-from-file)
   * [Serving data from reader](#serving-data-from-reader)
   * [Templates](#templates)
   * [Redirects](#redirects)
   * [Subdomains](#subdomains)
   * [Username/password authentication middleware](#usernamepassword-authentication-middleware)
   * [Custom middleware](#custom-middleware)
   * [Custom port](#custom-port)
   * [Custom TLS config](#custom-tls-config)
   * [Testing](#testing)
* [Who uses Gig](#who-uses-gig)
* [Benchmarks](#benchmarks)
* [Contribute](#contribute)
* [License](#license)

## Feature Overview

- Client certificate suppport (access `x509.Certificate` directly from context)
- Highly optimized router with zero dynamic memory allocation which smartly prioritizes routes
- Group APIs
- Extensible middleware framework
- Define middleware at root, group or route level
- Handy functions to send variety of Gemini responses
- Centralized error handling
- Template rendering with any template engine
- Define your format for the logger
- Highly customizable

## Guide

### Quick Start

```go
package main

import "github.com/pitr/gig"

func main() {
  // Gig instance
  g := gig.Default()

  // Routes
  g.Handle("/", func(c gig.Context) error {
    return c.Gemini("# Hello, World!")
  })

  // Start server on PORT or default port
  g.Run("my.crt", "my.key")
}
```
```bash
$ go run main.go
```

### Parameters in path

```go
package main

import "github.com/pitr/gig"

func main() {
  g := gig.Default()

  g.Handle("/user/:name", func(c gig.Context) error {
    return c.Gemini("# Hello, %s!", c.Param("name"))
  })

  g.Run("my.crt", "my.key")
}
```

### Query

```go
package main

import "github.com/pitr/gig"

func main() {
  g := gig.Default()

  g.Handle("/user", func(c gig.Context) error {
    query, err := c.QueryString()
    if err != nil {
      return err
    }
    return c.Gemini("# Hello, %s!", query)
  })

  g.Run("my.crt", "my.key")
}
```

### Client Certificate

```go
package main

import "github.com/pitr/gig"

func main() {
  g := gig.Default()

  g.Handle("/user", func(c gig.Context) error {
    cert := c.Certificate()
    if cert == nil {
      return c.NoContent(gig.StatusClientCertificateRequired, "We need a certificate")
    }
    return c.Gemini("# Hello, %s!", cert.Subject.CommonName)
  })

  // OR using middleware

  g.Handle("/user", func(c gig.Context) error {
    return c.Gemini("# Hello, %s!", c.Get("subject"))
  }, gig.CertAuth(gig.ValidateHasCertificate))

  g.Run("my.crt", "my.key")
}
```

### Grouping routes
```go
func main() {
  g := gig.Default()

  // Simple group: v1
  v1 := g.Group("/v1")
  {
    v1.Handle("/page1", page1Endpoint)
    v1.Handle("/page2", page2Endpoint)
  }

  // Simple group: v2
  v2 := g.Group("/v2")
  {
    v2.Handle("/page1", page1Endpoint)
    v2.Handle("/page2", page2Endpoint)
  }

  g.Run("my.crt", "my.key")
}
```

### Blank Gig without middleware by default
Use
```go
g := gig.New()
```
instead of
```go
// Default With the Logger and Recovery middleware already attached
g := gig.Default()
```

### Using middleware
```go
func main() {
  // Creates a router without any middleware by default
  g := gig.New()

  // Global middleware
  // Logger middleware will write the logs to gig.DefaultWriter.
  // By default gig.DefaultWriter = os.Stdout
  g.Use(gig.Logger())

  // Recovery middleware recovers from any panics and return StatusPermanentFailure.
  g.Use(gig.Recovery())

  // Private group
  // same as private := g.Group("/private", gig.CertAuth(gig.ValidateHasCertificate))
  private := g.Group("/private")
  private.Use(gig.CertAuth(gig.ValidateHasCertificate))
  {
    private.Handle("/user", userEndpoint)
  }

  g.Run("my.crt", "my.key")
}
```

### Writing logs to file
```go
func main() {
  f, _ := os.Create("access.log")
  gig.DefaultWriter = io.MultiWriter(f)

  // Use the following code if you need to write the logs to file and console at the same time.
  // gig.DefaultWriter = io.MultiWriter(f, os.Stdout)

  g := gig.Default()

  g.Handle("/", func(c gig.Context) error {
      return c.Gemini("# Hello, World!")
  })

  g.Run("my.crt", "my.key")
}
```

### Custom Log Format
```go
func main() {
  g := gig.New()

  // See LoggerConfig documentation for format
  g.Use(gig.LoggerWithConfig(gig.LoggerConfig{Format: "${remote_ip} ${status}"}))

  g.Handle("/", func(c gig.Context) error {
      return c.Gemini("# Hello, World!")
  })

  g.Run("my.crt", "my.key")
}
```

### Serving static files
```go
func main() {
  g := gig.Default()

  g.Static("/images", "images")
  g.Static("/robots.txt", "files/robots.txt")

  g.Run("my.crt", "my.key")
}
```

### Serving data from file
```go
func main() {
  g := gig.Default()

  g.Handle("/robots.txt", func(c gig.Context) error {
      return c.File("robots.txt")
  })

  g.Run("my.crt", "my.key")
}
```

### Serving data from reader
```go
func main() {
  g := gig.Default()

  g.Handle("/data", func(c gig.Context) error {
    response, err := http.Get("https://google.com/")
    if err != nil || response.StatusCode != http.StatusOK {
      return c.NoContent(gig.StatusProxyError, "could not fetch google")
    }

    return c.Stream("text/html", response.Body)
  })

  g.Run("my.crt", "my.key")
}
```

### Templates

Set `Gig.Renderer` to something that responds to `Render(io.Writer, string, interface{}, gig.Context) error`.

Use any templating library, such as `text/template`, [https://github.com/valyala/quicktemplate](https://github.com/valyala/quicktemplate), etc. The following example uses `text/template`:

```go
import (
  "text/template"

  "github.com/pitr/gig"
)

type Template struct {
  templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c gig.Context) error {
  // Execute named template with data
  return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
  g := gig.Default()

  // Register renderer
  g.Renderer = &Template{template.Must(template.ParseGlob("public/views/*.gmi"))}

  g.Handle("/user/:name", func(c gig.Context) error {
    // Render template "user" with username passed as data.
    return c.Render("user", c.Param("name"))
  })

  g.Run("my.crt", "my.key")
}
```

Consider bundling assets with the binary by using [go:ember](https://tip.golang.org/pkg/embed/), [go-assets](https://github.com/jessevdk/go-assets) or similar.

### Redirects
```go
func main() {
  g := gig.Default()

  g.Handle("/old", func(c gig.Context) error {
    return c.NoContent(gig.StatusRedirectPermanent, "/new")
  })

  g.Run("my.crt", "my.key")
}
```

### Subdomains

```go
func main() {
  apps := map[string]*gig.Gig{}

  // App A
  a := gig.Default()
  apps["app-a.example.com"] = a

  a.Handle("/", func(c gig.Context) error {
      return c.Gemini("I am App A")
  })

  // App B
  b := gig.Default()
  apps["app-b.example.com"] = b

  b.Handle("/", func(c gig.Context) error {
      return c.Gemini("I am App B")
  })

  // Server (without default middleware to prevent double logging)
  g := gig.New()
  g.Handle("/*", func(c gig.Context) error {
      app := apps[c.URL().Host]

      if app == nil {
          return gig.ErrNotFound
      }

      app.ServeGemini(c)
      return nil
  })

  g.Run("my.crt", "my.key") // must be wildcard SSL certificate for *.example.com
}
```

### Username/password authentication middleware

Status: EXPERIMENTAL

`PassAuth` middleware ensures that request has a client certificate, validates its fingerprint using function passed to middleware. If authentication is required, this function should return a path where user should be redirect to.

Login handlers are setup using `PassAuthLoginHandle` function, which collects username and password, and passes them to the provided function. That function should return an error if login failed, or absolute path to redirect user to.

User registration is expected to be implemented by developer.

The example assumes that there is a `db` module that does user management.

```go
func main() {
  g := Default()

  secret := g.Group("/secret", gig.PassAuth(func(sig string, c gig.Context) (string, error) {
    ok, err := db.CheckValid(sig)
    if err != nil {
      return "/login", err
    }
    if !ok {
      return "/login", nil
    }
    return "", nil
  }))
  // secret.Handle("/page", func(c gig.Context) {...})

  g.PassAuthLoginHandle("/login", func(user, pass, sig string, c Context) (string, error) {
    // check user/pass combo, and activate cert signature if valid
    err := db.Login(user, pass, sig)
    if err != nil {
      return "", err
    }
    return "/secret/page", nil
  })

  g.Run("my.crt", "my.key")
}
```

### Custom middleware
```go
func MyMiddleware(next gig.HandlerFunc) gig.HandlerFunc {
  return func(c gig.Context) error {
    // Set example variable
    c.Set("example", "123")

    if err := next(c); err != nil {
      c.Error(err)
    }

    // Do something after request is done
    // ...

    return err
  }
}

func main() {
  g := gig.Default()
  g.Use(MyMiddleware)

  g.Handle("/", func(c gig.Context) error {
    return c.Gemini("# Example %s", c.Get("example"))
  })

  g.Run("my.crt", "my.key")
}
```

### Custom port

Use `PORT` environment variable:

```
PORT=12345 ./myapp
```

Alternatively, pass it to Run:

```go
func main() {
  g := gig.Default()

  g.Handle("/", func(c gig.Context) error {
    return c.Gemini("# Hello world")
  })

  g.Run(":12345", "my.crt", "my.key")
}
```

### Custom TLS config
```go
func main() {
  g := gig.Default()
  g.TLSConfig.MinVersion = tls.VersionTLS13

  g.Handle("/", func(c gig.Context) error {
    return c.Gemini("# Hello world")
  })

  g.Run("my.crt", "my.key")
}
```

### Testing
```go
func setupServer() *gig.Gig {
  g := gig.Default()

  g.Handle("/private", func(c gig.Context) error {
    return c.Gemini("Hello %s", c.Get("subject"))
  }, gig.CertAuth(gig.ValidateHasCertificate))

  return g
}

func TestServer(t *testing.T) {
  g := setupServer()
  c, res := g.NewFakeContext("/private", nil)

  g.ServeGemini(c)

  if res.Written != "60 Client Certificate Required\r\n" {
    t.Fail()
  }
}

func TestCertificate(t *testing.T) {
  g := setupServer()
  c, res := g.NewFakeContext("/", &tls.ConnectionState{
    PeerCertificates: []*x509.Certificate{
      {Subject: pkix.Name{CommonName: "john"}},
    },
  })

  g.ServeGemini(c)

  if resp.Written != "20 text/gemini\r\nHello john" {
    t.Fail()
  }
}
```

## Who uses Gig

Gig is used by the following capsules:

- [gemif.fedi.farm](https://portal.mozz.us/gemini/gemif.fedi.farm) - GemIf, Interactive Fiction engine
- [geddit.glv.one](https://portal.mozz.us/gemini/geddit.glv.one) - Link aggregator
- [wp.glv.one](https://portal.mozz.us/gemini/wp.glv.one) - Wikipedia proxy
- [egsam.glv.one](https://portal.mozz.us/gemini/egsam.glv.one) - Egsam, client torture test
- [paste.gemigrep.com](https://portal.mozz.us/gemini/paste.gemigrep.com) - Paste service
- [gemini.tunerapp.org](https://portal.mozz.us/gemini/gemini.tunerapp.org) - Internet Radio Stations Directory

If you use Gig, open a PR to add your capsule to this list.

## Benchmarks

| Benchmark name                 |      (1) |           (2) |        (3) |            (4) |
| ------------------------------ | --------:| -------------:| ----------:| --------------:|
| BenchmarkRouterStaticRoutes    |   104677 |   11105 ns/op |     0 B/op |    0 allocs/op |
| BenchmarkRouterGitHubAPI       |    50859 |   22973 ns/op |     0 B/op |    0 allocs/op |
| BenchmarkRouterParseAPI        |   302828 |    3717 ns/op |     0 B/op |    0 allocs/op |
| BenchmarkRouterGooglePlusAPI   |   185558 |    6136 ns/op |     0 B/op |    0 allocs/op |

Generated using `make bench` in [router_test.go](https://github.com/pitr/gig/blob/master/router_test.go). APIs are based on [Go HTTP Router Benchmark repository](https://github.com/gin-gonic/go-http-routing-benchmark) and adapted to Gemini protocol, eg. verbs
GET/POST/etc are ignored since Gemini does not support them.

- (1): Total Repetitions achieved in constant time, higher means more confident result
- (2): Single Repetition Duration (ns/op), lower is better
- (3): Heap Memory (B/op), lower is better
- (4): Average Allocations per Repetition (allocs/op), lower is better

## Contribute

If something is missing, please open an issue. If possible, send a PR.

## License

[MIT](https://github.com/pitr/gig/blob/master/LICENSE)
