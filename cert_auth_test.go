package gig

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"

	"github.com/matryer/is"
)

func TestCertAuth(t *testing.T) {
	is := is.New(t)
	g := New()

	testCases := []struct {
		mw                 MiddlewareFunc
		expectedErrNoCert  error
		expectedErrBadCert error
		name               string
	}{
		{
			mw:                 CertAuth(ValidateHasCertificate),
			expectedErrNoCert:  ErrClientCertificateRequired,
			expectedErrBadCert: nil,
			name:               `ValidateHasCertificate`,
		},
		{
			mw: CertAuth(func(cert *x509.Certificate, c Context) *GeminiError {
				if cert == nil {
					return ErrClientCertificateRequired
				}

				if cert.Subject.CommonName != "tester" {
					return ErrCertificateNotValid
				}

				c.Set("subject", cert.Subject.CommonName)

				return nil
			}),
			expectedErrNoCert:  ErrClientCertificateRequired,
			expectedErrBadCert: ErrCertificateNotValid,
			name:               `CustomValidator`,
		},
		{
			mw: CertAuthWithConfig(CertAuthConfig{
				Skipper:   nil,
				Validator: nil,
			}),
			expectedErrNoCert:  ErrClientCertificateRequired,
			expectedErrBadCert: nil,
			name:               `NilConfig`,
		},
		{
			mw: CertAuthWithConfig(CertAuthConfig{
				Skipper: func(c Context) bool {
					c.Set("subject", "tester")

					return true
				},
			}),
			expectedErrNoCert:  nil,
			expectedErrBadCert: nil,
			name:               `CustomSkipper`,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			h := test.mw(func(c Context) error {
				return c.Gemini("test")
			})

			// No certificate
			c, _ := g.NewFakeContext("/", nil)
			is.Equal(h(c), test.expectedErrNoCert)

			// Invalid certificate
			c, _ = g.NewFakeContext("/", &tls.ConnectionState{
				PeerCertificates: []*x509.Certificate{
					{Subject: pkix.Name{CommonName: "not-tester"}},
				},
			})
			is.Equal(h(c), test.expectedErrBadCert)

			// Valid certificate
			c, _ = g.NewFakeContext("/", &tls.ConnectionState{
				PeerCertificates: []*x509.Certificate{
					{Subject: pkix.Name{CommonName: "tester"}},
				},
			})
			is.NoErr(h(c))
			is.Equal("tester", c.Get("subject"))
		})
	}
}
