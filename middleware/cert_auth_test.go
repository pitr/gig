package middleware

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"

	"github.com/pitr/gig"
	"github.com/pitr/gig/gigtest"
	"github.com/stretchr/testify/assert"
)

func TestCertAuth(t *testing.T) {
	assert := assert.New(t)

	e := gig.New()
	c, _ := gigtest.NewContext(e, "/", nil)

	f := func(cert *x509.Certificate, c gig.Context) *gig.GeminiError {
		if cert == nil {
			return gig.ErrClientCertificateRequired
		}
		if cert.Subject.CommonName != "gig-tester" {
			return gig.ErrCertificateNotAccepted
		}
		return nil
	}
	h := CertAuth(f)(func(c gig.Context) error {
		return c.Gemini(gig.StatusSuccess, "test")
	})

	// No certificate
	assert.Equal(h(c), gig.ErrClientCertificateRequired)

	// Invalid certificate
	c, _ = gigtest.NewContext(e, "/", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{
			&x509.Certificate{Subject: pkix.Name{CommonName: "wrong"}},
		},
	})
	assert.Equal(h(c), gig.ErrCertificateNotAccepted)

	// Valid certificate
	c, _ = gigtest.NewContext(e, "/", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{
			&x509.Certificate{Subject: pkix.Name{CommonName: "gig-tester"}},
		},
	})
	assert.NoError(h(c))
}

func TestCertAuth_Validators(t *testing.T) {
	assert := assert.New(t)
	e := gig.New()

	testCases := []struct {
		validator   CertAuthValidator
		expectedErr error
		name        string
	}{
		{
			validator:   ValidateHasCertificate,
			expectedErr: gig.ErrClientCertificateRequired,
			name:        `ValidateHasCertificate`,
		},
		{
			validator:   ValidateHasTransientCertificate,
			expectedErr: gig.ErrTransientCertificateRequested,
			name:        `ValidateHasTransientCertificate`,
		},
		{
			validator:   ValidateHasAuthorisedCertificate,
			expectedErr: gig.ErrAuthorisedCertificateRequired,
			name:        `ValidateHasAuthorisedCertificate`,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			h := CertAuth(test.validator)(func(c gig.Context) error {
				return c.Gemini(gig.StatusSuccess, "test")
			})

			// No certificate
			c, _ := gigtest.NewContext(e, "/", nil)
			assert.Equal(h(c), test.expectedErr)

			// Invalid certificate
			c, _ = gigtest.NewContext(e, "/", &tls.ConnectionState{
				PeerCertificates: []*x509.Certificate{
					&x509.Certificate{Subject: pkix.Name{CommonName: "tester"}},
				},
			})
			assert.NoError(h(c), nil)
			assert.Equal("tester", c.Get("subject").(string))
		})
	}
}
