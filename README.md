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
  g := gig.New()

  // Middleware
  g.Use(middleware.Logger())
  g.Use(middleware.Recover())

  // Routes
  g.Handle("/", func(c gig.Context) error {
      return c.Gemini(gig.StatusSuccess, "# Hello, World!")
  })

  // Start server
  panic(g.Run(":1323", "my.crt", "my.key"))
}
```

## Contribute

If something is missing, please open an issue. If possible, send a PR.

## License

[MIT](https://github.com/pitr/gig/blob/master/LICENSE)
