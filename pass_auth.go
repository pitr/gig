package gig

type (
	// PassAuthCertCheck defines a function to validate certificate fingerprint.
	// Must return path on unsuccessful login.
	PassAuthCertCheck func(string, Context) (string, error)
	// PassAuthLogin defines a function to login user.
	// It may pin certificate to user if login is successful.
	// Must return path to redirect to after login.
	PassAuthLogin func(username, password, sig string, c Context) (string, error)
)

// PassAuth is a middleware that implements username/password authentication
// by first requiring a certificate, checking username/password using PassAuthValidator,
// and then pinning certificate to it.
//
// For valid credentials it calls the next handler.
func PassAuth(check PassAuthCertCheck) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			// If no client certificate is sent, request it
			var sig = c.CertHash()
			if sig == "" {
				return c.NoContent(StatusClientCertificateRequired, "Please create a certificate")
			}

			to, err := check(sig, c)
			if err != nil {
				debugPrintf("gemini: could not check certificate: %s", err)
				return c.NoContent(StatusBadRequest, "Try again later")
			}

			if to != "" {
				return c.NoContent(StatusRedirectTemporary, to)
			}

			return next(c)
		}
	}
}

// PassAuthLoginHandle sets up handlers to check username/password using PassAuthLogin.
func (g *Gig) PassAuthLoginHandle(path string, fn PassAuthLogin) {
	g.Handle(path, func(c Context) error {
		cert := c.Certificate()
		if cert == nil {
			return c.NoContent(StatusClientCertificateRequired, "Please create a certificate")
		}
		username, err := c.QueryString()
		if err != nil {
			debugPrintf("gemini: could not extract username from URL: %s", err)
			return c.NoContent(StatusBadRequest, "Invalid username received")
		}

		if username == "" {
			return c.NoContent(StatusInput, "Enter username")
		}

		return c.NoContent(StatusRedirectTemporary, "%s/%s", path, username)
	})

	g.Handle(path+"/:username", func(c Context) error {
		var (
			username = c.Param("username")
			sig      = c.CertHash()
		)

		if sig == "" {
			return c.NoContent(StatusClientCertificateRequired, "Please create a certificate")
		}

		if username == "" {
			return c.NoContent(StatusRedirectTemporary, path)
		}

		password, err := c.QueryString()

		if err != nil {
			debugPrintf("gemini: could not extract password from URL: %s", err)
			return c.NoContent(StatusBadRequest, "Invalid password received")
		}

		if password == "" {
			return c.NoContent(StatusSensitiveInput, "Enter password")
		}

		to, err := fn(username, password, sig, c)

		if err != nil {
			return err
		}

		return c.NoContent(StatusRedirectTemporary, to)
	})
}
