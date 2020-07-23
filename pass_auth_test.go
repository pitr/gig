package gig

import (
	"crypto/md5"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"testing"

	"github.com/matryer/is"
)

func TestPassAuth(t *testing.T) {
	var (
		is         = is.New(t)
		g          = New()
		invalidErr = errors.New("invalid credentials")
		goodCert   = x509.Certificate{Raw: []byte{1}}
		newCert    = x509.Certificate{Raw: []byte{2}}
		mw         = PassAuth(PassAuthConfig{
			Skipper: func(c Context) bool {
				return c.URL().Path == "/skip-me"
			},
			CertCheck: func(sig string, c Context) (bool, error) {
				return sig == fmt.Sprintf("%x", md5.Sum(goodCert.Raw)), nil
			},
			Login: func(u, p, sig string, c Context) (string, error) {
				if u == "valid-user" && p == "secret" {
					return "/my-profile", nil
				}
				return "", invalidErr
			},
		})
		h = mw(func(c Context) error {
			return c.Gemini("private")
		})
	)

	// No certificate, with skipper
	c, res := g.NewFakeContext("/skip-me", nil)
	is.NoErr(h(c))
	is.Equal(res.Written, "20 text/gemini\r\nprivate")

	// No certificate
	c, res = g.NewFakeContext("/", nil)
	is.NoErr(h(c))
	is.Equal(res.Written, "60 Please create a certificate\r\n")

	// New certificate
	c, res = g.NewFakeContext("/", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&newCert},
	})
	is.NoErr(h(c))
	is.Equal(res.Written, "30 /login\r\n")

	// Try login with new certificate
	c, res = g.NewFakeContext("/login", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&newCert},
	})
	is.NoErr(h(c))
	is.Equal(res.Written, "10 Enter username\r\n")

	c, res = g.NewFakeContext("/login?valid-user", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&newCert},
	})
	is.NoErr(h(c))
	is.Equal(res.Written, "30 /login/valid-user\r\n")

	c, res = g.NewFakeContext("/login/", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&newCert},
	})
	is.NoErr(h(c))
	is.Equal(res.Written, "30 /login\r\n")

	c, res = g.NewFakeContext("/login/valid-user", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&newCert},
	})
	is.NoErr(h(c))
	is.Equal(res.Written, "11 Enter password\r\n")

	c, _ = g.NewFakeContext("/login/valid-user?bad-pass", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&newCert},
	})
	is.Equal(h(c), invalidErr)

	c, res = g.NewFakeContext("/login/valid-user?secret", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&newCert},
	})
	is.NoErr(h(c))
	is.Equal(res.Written, "30 /my-profile\r\n")

	// Logged in certificate
	c, res = g.NewFakeContext("/", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&goodCert},
	})
	is.NoErr(h(c))
	is.Equal(res.Written, "20 text/gemini\r\nprivate")
}

func TestPassAuth_Errors(t *testing.T) {
	var (
		config = PassAuthConfig{
			CertCheck: func(sig string, c Context) (bool, error) {
				return false, errors.New("oops")
			},
			Login: func(u, p, sig string, c Context) (string, error) {
				return "", errors.New("oops")
			},
		}
		is      = is.New(t)
		g       = New()
		newCert = x509.Certificate{Raw: []byte{2}}
		mw      = PassAuth(config)
		h       = mw(func(c Context) error {
			return c.Gemini("private")
		})
	)

	// CertCheck fails
	c, res := g.NewFakeContext("/", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&newCert},
	})
	is.NoErr(h(c))
	is.Equal(res.Written, "59 Try again later\r\n")

	// Bad username
	c, res = g.NewFakeContext("/login?%%", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&newCert},
	})
	is.NoErr(h(c))
	is.Equal(res.Written, "59 Invalid username received\r\n")

	// Bad password
	c, res = g.NewFakeContext("/login/valid-user?%%", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&newCert},
	})
	is.NoErr(h(c))
	is.Equal(res.Written, "59 Invalid password received\r\n")

	// Login fails
	c, _ = g.NewFakeContext("/login/valid-user?secret", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&newCert},
	})
	is.True(h(c) != nil)
}
