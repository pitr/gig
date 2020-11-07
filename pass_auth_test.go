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
		mw         = PassAuth(func(sig string, c Context) (string, error) {
			if sig == fmt.Sprintf("%x", md5.Sum(goodCert.Raw)) {
				return "", nil
			}
			return "/login", nil
		})
	)

	g.Handle("/", func(c Context) error {
		return c.Gemini("ok")
	})
	g.PassAuthLoginHandle("/login", func(u, p, sig string, c Context) (string, error) {
		if u == "valid-user" && p == "secret" {
			return "/private", nil
		}
		return "", invalidErr
	})
	g.Handle("/private", func(c Context) error {
		return c.Gemini("private")
	}, mw)

	// Public endpoint
	c, res := g.NewFakeContext("/", nil)
	g.ServeGemini(c)
	is.Equal(res.Written, "20 text/gemini\r\nok")

	// No certificate
	c, res = g.NewFakeContext("/private", nil)
	g.ServeGemini(c)
	is.Equal(res.Written, "60 Please create a certificate\r\n")

	// New certificate
	c, res = g.NewFakeContext("/private", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&newCert},
	})
	g.ServeGemini(c)
	is.Equal(res.Written, "30 /login\r\n")

	// Try login with new certificate
	c, res = g.NewFakeContext("/login", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&newCert},
	})
	g.ServeGemini(c)
	is.Equal(res.Written, "10 Enter username\r\n")

	c, res = g.NewFakeContext("/login?valid-user", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&newCert},
	})
	g.ServeGemini(c)
	is.Equal(res.Written, "30 /login/valid-user\r\n")

	c, res = g.NewFakeContext("/login/valid-user", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&newCert},
	})
	g.ServeGemini(c)
	is.Equal(res.Written, "11 Enter password\r\n")

	c, res = g.NewFakeContext("/login/valid-user?bad-pass", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&newCert},
	})
	g.ServeGemini(c)
	is.Equal(res.Written, "50 invalid credentials\r\n")

	c, res = g.NewFakeContext("/login/valid-user?secret", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&newCert},
	})
	g.ServeGemini(c)
	is.Equal(res.Written, "30 /private\r\n")

	// Logged in certificate
	c, res = g.NewFakeContext("/private", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&goodCert},
	})
	g.ServeGemini(c)
	is.Equal(res.Written, "20 text/gemini\r\nprivate")
}

func TestPassAuth_Errors(t *testing.T) {
	var (
		is      = is.New(t)
		g       = New()
		newCert = x509.Certificate{Raw: []byte{2}}
		mw      = PassAuth(func(sig string, c Context) (string, error) {
			return "/login", errors.New("oops")
		})
	)

	g.Handle("/", func(c Context) error {
		return c.Gemini("ok")
	})
	g.PassAuthLoginHandle("/login", func(u, p, sig string, c Context) (string, error) {
		return "", errors.New("oops")
	})
	g.Handle("/private", func(c Context) error {
		return c.Gemini("private")
	}, mw)

	// CertCheck fails
	c, res := g.NewFakeContext("/private", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&newCert},
	})
	g.ServeGemini(c)
	is.Equal(res.Written, "59 Try again later\r\n")

	// Bad username
	c, res = g.NewFakeContext("/login?%%", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&newCert},
	})
	g.ServeGemini(c)
	is.Equal(res.Written, "59 Invalid username received\r\n")

	// Bad password
	c, res = g.NewFakeContext("/login/valid-user?%%", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&newCert},
	})
	g.ServeGemini(c)
	is.Equal(res.Written, "59 Invalid password received\r\n")

	// Login fails
	c, res = g.NewFakeContext("/login/valid-user?secret", &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{&newCert},
	})
	g.ServeGemini(c)
	is.Equal(res.Written, "50 oops\r\n")
}
