# Gig - Gemini framework

[![Sourcegraph](https://sourcegraph.com/github.com/pitr/gig/-/badge.svg?style=flat-square)](https://sourcegraph.com/github.com/pitr/gig?badge)
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/pitr/gig)
[![Go Report Card](https://goreportcard.com/badge/github.com/pitr/gig?style=flat-square)](https://goreportcard.com/report/github.com/pitr/gig)
[![Build Status](http://img.shields.io/travis/pitr/gig.svg?style=flat-square)](https://travis-ci.org/pitr/gig)
[![Codecov](https://img.shields.io/codecov/c/github/pitr/gig.svg?style=flat-square)](https://codecov.io/gh/pitr/gig)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/pitr/gig/master/LICENSE)

## Feature Overview

- Optimized router which smartly prioritize routes
- Group APIs
- Extensible middleware framework
- Define middleware at root, group or route level
- Handy functions to send variety of Gemini responses
- Centralized Gemini error handling
- Template rendering with any template engine
- Define your format for the logger
- Highly customizable
- Automatic TLS via Let’s Encrypt

## Guide

### Example

```go
package main

import (
  "github.com/pitr/gig"
  "github.com/pitr/gig/middleware"
)

func main() {
  // Gig instance
  e := gig.New()

  // Middleware
  e.Use(middleware.Logger())
  e.Use(middleware.Recover())

  // Routes
  e.Handle("/", hello)

  // Start server
  e.Logger.Fatal(e.StartTLS(":1965", "cert.pem", "key.pem"))
  // or use automatic certificates installed from https://letsencrypt.org.
  // e.Logger.Fatal(e.StartAutoTLS(":1965"))
}

// Handler
func hello(c gig.Context) error {
  return c.String(gig.StatusSuccess, "Hello, World!")
}
```

## Contribute

If something is missing, please open an issue. If possible, send a PR.

## License

[MIT](https://github.com/pitr/gig/blob/master/LICENSE)
