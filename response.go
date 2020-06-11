package gig

import (
	"fmt"
	"io"
)

type (
	// Response wraps an io.Conn and implements its interface to be used
	// by a handler to construct a response.
	Response struct {
		logger    Logger
		Writer    io.Writer
		Status    Status
		Meta      string
		Size      int64
		Committed bool
	}
)

// NewResponse creates a new instance of Response.
func NewResponse(w io.Writer, logger Logger) (r *Response) {
	return &Response{Writer: w, logger: logger}
}

// WriteHeader sends a Gemini response header with status code. If WriteHeader is
// not called explicitly, the first call to Write will trigger an implicit
// WriteHeader(StatusSuccess, "text/gemini"). Thus explicit calls to WriteHeader
// are mainly used to send error codes.
func (r *Response) WriteHeader(code Status, meta string) error {
	if r.Committed {
		r.logger.Warn("response already committed")
		return nil
	}
	r.Status = code
	r.Meta = meta
	n, err := r.Writer.Write([]byte(fmt.Sprintf("%d %s\r\n", code, meta)))
	r.Committed = true
	r.Size += int64(n)

	return err
}

// Write writes the data to the connection as part of a reply.
func (r *Response) Write(b []byte) (n int, err error) {
	if !r.Committed {
		if r.Status == 0 {
			r.Status = StatusSuccess
		}
		err = r.WriteHeader(r.Status, "text/gemini")
		if err != nil {
			return
		}
	}
	n, err = r.Writer.Write(b)
	r.Size += int64(n)
	return
}

func (r *Response) reset(w io.Writer) {
	r.Writer = w
	r.Size = 0
	r.Meta = ""
	r.Status = StatusSuccess
	r.Committed = false
}
