package gig

import (
	"fmt"
	"io"
)

type (
	// Response wraps net.Conn, to be used by a context to construct a response.
	Response struct {
		Writer    io.Writer
		Status    Status
		Meta      string
		Size      int64
		Committed bool
		err       error
	}
)

// NewResponse creates a new instance of Response. Typically used for tests.
func NewResponse(w io.Writer) (r *Response) {
	return &Response{Writer: w}
}

// WriteHeader sends a Gemini response header with status code. If WriteHeader is
// not called explicitly, the first call to Write will trigger an implicit
// WriteHeader(StatusSuccess, "text/gemini"). Thus explicit calls to WriteHeader
// are mainly used to send error codes.
func (r *Response) WriteHeader(code Status, meta string) error {
	if r.Committed {
		debugPrintf("gemini: response already committed")
		return nil
	}

	r.Status = code
	r.Meta = meta

	var n int

	n, r.err = r.Writer.Write([]byte(fmt.Sprintf("%d %s\r\n", code, meta)))
	r.Committed = true

	if r.err != nil {
		return r.err
	}

	r.Size += int64(n)

	return nil
}

// Write writes the data to the connection as part of a reply.
func (r *Response) Write(b []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}

	if !r.Committed {
		if r.Status == 0 {
			r.Status = StatusSuccess
		}

		r.err = r.WriteHeader(r.Status, "text/gemini")

		if r.err != nil {
			return 0, r.err
		}
	}

	var n int
	n, r.err = r.Writer.Write(b)

	if r.err != nil {
		return n, r.err
	}

	r.Size += int64(n)

	return n, nil
}

func (r *Response) reset(w io.Writer) {
	r.Writer = w
	r.Size = 0
	r.Meta = ""
	r.Status = StatusSuccess
	r.Committed = false
	r.err = nil
}
