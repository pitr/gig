# Gig - Gemini framework

[![Sourcegraph](https://sourcegraph.com/github.com/pitr/gig/-/badge.svg?style=flat-square)](https://sourcegraph.com/github.com/pitr/gig?badge)
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/pitr/gig)
[![Go Report Card](https://goreportcard.com/badge/github.com/pitr/gig?style=flat-square)](https://goreportcard.com/report/github.com/pitr/gig)
[![Codecov](https://img.shields.io/codecov/c/github/pitr/gig.svg?style=flat-square)](https://codecov.io/gh/pitr/gig)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/pitr/gig/master/LICENSE)

API is subject to change until v1.0

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
   * [Custom middleware](#custom-middleware)
   * [Custom port](#custom-port)
   * [Custom TLS config](#custom-tls-config)
   * [Testing](#testing)
* [Contribute](#contribute)
* [License](#license)

## Feature Overview

- Client certificate suppport (access `x509.Certificate` directly from context)
- Optimized router which smartly prioritize routes
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
    return c.Gemini(gig.StatusSuccess, "# Hello, World!")
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
    return c.Gemini(gig.StatusSuccess, "# Hello, %s!", c.Param("name"))
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
    return c.Gemini(gig.StatusSuccess, "# Hello, %s!", query)
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
      return NewError(gig.StatusClientCertificateRequired, "We need a certificate")
    }
    return c.Gemini(gig.StatusSuccess, "# Hello, %s!", cert.Subject.CommonName)
  })

  // OR
  g.Handle("/user", func(c gig.Context) error {
    return c.Gemini(gig.StatusSuccess, "# Hello, %s!", c.Get("subject"))
  }, gig.CertAuth(gig.ValidateHasAuthorisedCertificate))

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
  // same as private := g.Group("/private", gig.CertAuth(gig.ValidateHasAuthorisedCertificate))
  private := g.Group("/private")
  private.Use(gig.CertAuth(gig.ValidateHasAuthorisedCertificate))
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
      return c.Gemini(gig.StatusSuccess, "# Hello, World!")
  })

  g.Run("my.crt", "my.key")
}
```

### Custom Log Format
```go
func main() {
  g := gig.New()

  // See LoggerConfig documentation for more
  g.Use(gig.LoggerWithConfig(gig.LoggerConfig{Format: "${remote_ip} ${status}"}))

  g.Handle("/", func(c gig.Context) error {
      return c.Gemini(gig.StatusSuccess, "# Hello, World!")
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
      return gig.ErrProxyError
    }

    return c.Stream(gig.StatusSuccess, "text/html", response.Body)
  })

  g.Run("my.crt", "my.key")
}
```

### Templates

Use `text/template`, [https://github.com/valyala/quicktemplate](https://github.com/valyala/quicktemplate), or anything else. This example uses `text/template`

```go
import (
  "text/template"

  "github.com/pitr/gig"
)

type Template struct {
    templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c gig.Context) error {
    return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
  g := gig.Default()

  // Register renderer
  g.Renderer = &Template{
    templates: template.Must(template.ParseGlob("public/views/*.gmi")),
  }

  g.Handle("/user/:name", func(c gig.Context) error {
    return c.Render(gig.StatusSuccess, "user", c.Param("name"))
  })

  g.Run("my.crt", "my.key")
}
```

Consider bundling assets with the binary by using [go-assets](https://github.com/jessevdk/go-assets) or similar.

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
    return c.Gemini(gig.StatusSuccess, "# Example %s", c.Get("example"))
  })

  g.Run("my.crt", "my.key")
}
```

### Custom port
```go
func main() {
  g := gig.Default()

  g.Handle("/", func(c gig.Context) error {
    return c.Gemini(gig.StatusSuccess, "# Hello world")
  })

  g.Run(":12345", "my.key")
}
```

### Custom TLS config
```go
func main() {
  g := gig.Default()
  g.TLSConfig.MinVersion = tls.VersionTLS13

  g.Handle("/", func(c gig.Context) error {
    return c.Gemini(gig.StatusSuccess, "# Hello world")
  })

  g.Run("my.crt", "my.key")
}
```

### Testing
```go
func setupServer() *gig.Gig {
  g := gig.Default()

  g.Handle("/private", func(c gig.Context) error {
    return c.Gemini(gig.StatusSuccess, "Hello %s", c.Get("subject"))
  }, gig.CertAuth(gig.ValidateHasAuthorisedCertificate))

  g.Run("my.crt", "my.key")
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

## Contribute

If something is missing, please open an issue. If possible, send a PR.

## License

[MIT](https://github.com/pitr/gig/blob/master/LICENSE)
