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
	c, _ := g.NewFakeContext("/", nil)

	f := func(cert *x509.Certificate, c Context) *GeminiError {
		if cert == nil {
			return ErrClientCertificateRequired
		}

		if cert.Subject.CommonName != "gig-tester" {
			return ErrCertificateNotValid
		}

		return nil
	}
	h := CertAuth(f)(func(c Context) error {
		return c.Gemini("test")
	})

	// No certificate
	is.Equal(h(c), ErrClientCertificateRequired)

	// Invalid certificate
	c, _ = g.NewFakeContext("/", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{
			{Subject: pkix.Name{CommonName: "wrong"}},
		},
	})
	is.Equal(h(c), ErrCertificateNotValid)

	// Valid certificate
	c, _ = g.NewFakeContext("/", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{
			{Subject: pkix.Name{CommonName: "gig-tester"}},
		},
	})
	is.NoErr(h(c))
}

func TestCertAuth_Validators(t *testing.T) {
	is := is.New(t)
	g := New()

	testCases := []struct {
		validator   CertAuthValidator
		expectedErr error
		name        string
	}{
		{
			validator:   ValidateHasCertificate,
			expectedErr: ErrClientCertificateRequired,
			name:        `ValidateHasCertificate`,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			h := CertAuth(test.validator)(func(c Context) error {
				return c.Gemini("test")
			})

			// No certificate
			c, _ := g.NewFakeContext("/", nil)
			is.Equal(h(c), test.expectedErr)

			// Invalid certificate
			c, _ = g.NewFakeContext("/", &tls.ConnectionState{
				PeerCertificates: []*x509.Certificate{
					{Subject: pkix.Name{CommonName: "tester"}},
				},
			})
			is.NoErr(h(c))
			is.Equal("tester", c.Get("subject").(string))
		})
	}
}
