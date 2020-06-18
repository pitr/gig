package gig

import (
	"crypto/x509"
)

type (
	// CertAuthConfig defines the config for CertAuth middleware.
	CertAuthConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Validator is a function to validate client certificate.
		// Required.
		Validator CertAuthValidator
	}

	// CertAuthValidator defines a function to validate CertAuth credentials.
	CertAuthValidator func(*x509.Certificate, Context) *GeminiError
)

var (
	// DefaultCertAuthConfig is the default CertAuth middleware config.
	DefaultCertAuthConfig = CertAuthConfig{
		Skipper:   DefaultSkipper,
		Validator: ValidateHasCertificate,
	}
)

// ValidateHasCertificate returns ErrClientCertificateRequired if no certificate is sent.
// It also stores subject name in context under "subject".
func ValidateHasCertificate(cert *x509.Certificate, c Context) *GeminiError {
	if cert == nil {
		return ErrClientCertificateRequired
	}

	c.Set("subject", cert.Subject.CommonName)

	return nil
}

// CertAuth returns an CertAuth middleware.
//
// For valid credentials it calls the next handler.
func CertAuth(fn CertAuthValidator) MiddlewareFunc {
	c := DefaultCertAuthConfig
	c.Validator = fn

	return CertAuthWithConfig(c)
}

// CertAuthWithConfig returns an CertAuth middleware with config.
// See `CertAuth()`.
func CertAuthWithConfig(config CertAuthConfig) MiddlewareFunc {
	// Defaults
	if config.Validator == nil {
		config.Validator = DefaultCertAuthConfig.Validator
	}

	if config.Skipper == nil {
		config.Skipper = DefaultCertAuthConfig.Skipper
	}

	return func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			// Verify credentials
			err := config.Validator(c.Certificate(), c)
			if err != nil {
				return err
			}

			return next(c)
		}
	}
}
