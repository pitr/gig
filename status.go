package gig

// Status is a Gemini status code type.
type Status int

// Gemini status codes as documented by specification.
// See: https://gemini.circumlunar.space/docs/spec-spec.txt
const (
	StatusInput                     Status = 10
	StatusSensitiveInput            Status = 11
	StatusSuccess                   Status = 20
	StatusRedirectTemporary         Status = 30
	StatusRedirectPermanent         Status = 31
	StatusTemporaryFailure          Status = 40
	StatusServerUnavailable         Status = 41
	StatusCGIError                  Status = 42
	StatusProxyError                Status = 43
	StatusSlowDown                  Status = 44
	StatusPermanentFailure          Status = 50
	StatusNotFound                  Status = 51
	StatusGone                      Status = 52
	StatusProxyRequestRefused       Status = 53
	StatusBadRequest                Status = 59
	StatusClientCertificateRequired Status = 60
	StatusCertificateNotAuthorised  Status = 61
	StatusCertificateNotValid       Status = 62
)
