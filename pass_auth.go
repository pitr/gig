package gig

import (
	"crypto/md5"
	"fmt"
	"strings"
)

type (
	// PassAuthConfig defines the config for PassAuth middleware.
	PassAuthConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// CertCheck is a function to validate client certificate.
		// Required.
		CertCheck PassAuthCertCheck

		// Login is a function to login the user, should check credentials and
		// return path where to redirect user to, or return an error if
		// credentials are incorrect.
		Login PassAuthLogin
	}

	// PassAuthCertCheck defines a function to validate certificate fingerprint.
	PassAuthCertCheck func(string, Context) (bool, error)
	// PassAuthLogin defines a function to login user.
	PassAuthLogin func(string, string, string, Context) (string, error)
)

// PassAuth is a middleware that implements username/password authentication
// by first requiring a certificate, checking username/password using PassAuthValidator,
// and then pinning certificate to .
//
// For valid credentials it calls the next handler.
func PassAuth(config PassAuthConfig) MiddlewareFunc {
	// Defaults
	if config.CertCheck == nil {
		panic("PassAuthCertCheck must be set")
	}

	if config.Login == nil {
		panic("PassAuthLogin must be set")
	}

	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}

	return func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			// If no client certificate is sent, request it
			cert := c.Certificate()
			if cert == nil {
				return c.NoContent(StatusClientCertificateRequired, "Please create a certificate")
			}

			var (
				sig  = fmt.Sprintf("%x", md5.Sum(cert.Raw))
				path = c.URL().Path
			)

			// Handle asking for username
			if path == "/login" {
				username, err := c.QueryString()
				if err != nil {
					debugPrintf("could not extract username from URL: %s", err)
					return c.NoContent(StatusBadRequest, "Invalid username received")
				}

				if username == "" {
					return c.NoContent(StatusInput, "Enter username")
				}

				return c.NoContent(StatusRedirectTemporary, "/login/%s", username)
			}

			// Handle asking for password
			if strings.HasPrefix(path, "/login/") {
				username := strings.TrimSpace(strings.TrimPrefix(path, "/login/"))
				if username == "" {
					return c.NoContent(StatusRedirectTemporary, "/login")
				}

				password, err := c.QueryString()

				if err != nil {
					debugPrintf("could not extract password from URL: %s", err)
					return c.NoContent(StatusBadRequest, "Invalid password received")
				}

				if password == "" {
					return c.NoContent(StatusSensitiveInput, "Enter password")
				}

				path, err := config.Login(username, password, sig, c)

				if err != nil {
					return err
				}

				return c.NoContent(StatusRedirectTemporary, path)
			}

			ok, err := config.CertCheck(sig, c)
			if err != nil {
				debugPrintf("could not check certificate: %s", err)
				return c.NoContent(StatusBadRequest, "Try again later")
			}

			if !ok {
				return c.NoContent(StatusRedirectTemporary, "/login")
			}

			return next(c)
		}
	}
}
