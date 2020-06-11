package middleware

import (
	"github.com/pitr/gig"
)

type (
	// Skipper defines a function to skip middleware. Returning true skips processing
	// the middleware.
	Skipper func(gig.Context) bool

	// BeforeFunc defines a function which is executed just before the middleware.
	BeforeFunc func(gig.Context)
)

// DefaultSkipper returns false which processes the middleware.
func DefaultSkipper(gig.Context) bool {
	return false
}
