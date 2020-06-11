package middleware

import (
	"crypto/x509"

	"github.com/pitr/gig"
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
	CertAuthValidator func(*x509.Certificate, gig.Context) *gig.GeminiError
)

var (
	// DefaultCertAuthConfig is the default CertAuth middleware config.
	DefaultCertAuthConfig = CertAuthConfig{
		Skipper: DefaultSkipper,
	}
)

// ValidateHasCertificate returns ErrClientCertificateRequired if no certificate is sent.
// It also stores subject name in context under "subject"
func ValidateHasCertificate(cert *x509.Certificate, c gig.Context) *gig.GeminiError {
	if cert == nil {
		return gig.ErrClientCertificateRequired
	}
	name := cert.Subject.CommonName
	c.Set("subject", name)
	return nil
}

// ValidateHasTransientCertificate returns ErrTransientCertificateRequested if no certificate is sent.
// It also stores subject name in context under "subject"
func ValidateHasTransientCertificate(cert *x509.Certificate, c gig.Context) *gig.GeminiError {
	if cert == nil {
		return gig.ErrTransientCertificateRequested
	}
	name := cert.Subject.CommonName
	c.Set("subject", name)
	return nil
}

// ValidateHasAuthorisedCertificate returns ErrAuthorisedCertificateRequired if no certificate is sent.
// It also stores subject name in context under "subject"
func ValidateHasAuthorisedCertificate(cert *x509.Certificate, c gig.Context) *gig.GeminiError {
	if cert == nil {
		return gig.ErrAuthorisedCertificateRequired
	}
	name := cert.Subject.CommonName
	c.Set("subject", name)
	return nil
}

// CertAuth returns an CertAuth middleware.
//
// For valid credentials it calls the next handler.
func CertAuth(fn CertAuthValidator) gig.MiddlewareFunc {
	c := DefaultCertAuthConfig
	c.Validator = fn
	return CertAuthWithConfig(c)
}

// CertAuthWithConfig returns an CertAuth middleware with config.
// See `CertAuth()`.
func CertAuthWithConfig(config CertAuthConfig) gig.MiddlewareFunc {
	// Defaults
	if config.Validator == nil {
		config.Validator = DefaultCertAuthConfig.Validator
	}
	if config.Skipper == nil {
		config.Skipper = DefaultCertAuthConfig.Skipper
	}

	return func(next gig.HandlerFunc) gig.HandlerFunc {
		return func(c gig.Context) error {
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
