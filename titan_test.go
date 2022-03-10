package gig

import (
	"bytes"
	"errors"
	"io"
	"net/url"
	"testing"

	"github.com/matryer/is"
)

func TestTitanURLParser(t *testing.T) {
	tests := []struct {
		url   string
		token string
		mime  string
		size  int
	}{
		{"titan://f.org/raw/Test;token=hello;mime=plain/text;size=10", "hello", "plain/text", 10},
		{"titan://f.org/;mime=plain/text;size=10", "", "plain/text", 10},
		{"titan://f.org/", "", "text/gemini", 0},
		{"titan://f.org/;mime=a", "", "a", 0},
	}

	for _, tc := range tests {
		tc := tc
		t.Run("", func(t *testing.T) {
			is := is.New(t)
			url, err := url.Parse(tc.url)
			is.NoErr(err)

			got := newTitanParams(url)
			is.Equal(got, titanParams{
				token: tc.token,
				mime:  tc.mime,
				size:  tc.size,
			})
		})
	}
}

func TestTitanRequest(t *testing.T) {
	tests := []struct {
		name        string
		uri         string
		reader      io.Reader
		sizeLimit   int
		expect      string
		handlerHook func(Context, *testing.T) error
	}{
		// Size param tests
		{
			name:   "no size",
			uri:    "titan://a.b",
			reader: nil,
			expect: "59 Size parameter is incorrect or not provided\r\n",
		},
		{
			name:   "wrong size",
			uri:    "titan://a.b;size=-1",
			reader: nil,
			expect: "59 Size parameter is incorrect or not provided\r\n",
		},
		{
			name:   "size is not a number",
			uri:    "titan://a.b;size=foo",
			reader: nil,
			expect: "59 Size parameter is incorrect or not provided\r\n",
		},
		{
			name:   "zero size",
			uri:    "titan://a.b;size=0",
			reader: nil,
			expect: "59 Size parameter is incorrect or not provided\r\n",
		},
		{
			name: "size provided",
			uri:  "titan://a.b/;size=30",
			handlerHook: func(c Context, t *testing.T) error {
				is := is.New(t)
				is.Equal(c.Get("titan"), true)
				is.Equal(c.Get("size").(int), 30)
				return nil
			},
		},
		{
			name:      "size bigger than size limit",
			uri:       "titan://a.b/;size=10",
			expect:    "59 Request is bigger than allowed 5 bytes\r\n",
			sizeLimit: 5,
		},
		{
			name: "gemini request on titan enabled endpoint",
			uri:  "gemini://a.b/",
			handlerHook: func(c Context, t *testing.T) error {
				is := is.New(t)
				is.Equal(c.Get("titan"), false)
				return nil
			},
		},
		{
			name:   "read correct ammout of data",
			uri:    "titan://a.b/;size=10",
			reader: bytes.NewBuffer(make([]byte, 10)),
			handlerHook: func(c Context, t *testing.T) error {
				is := is.New(t)
				b, err := TitanReadFull(c)
				is.NoErr(err)
				is.True(b != nil)
				is.Equal(len(b), 10)
				return nil
			},
		},
		{
			name:   "read underflow",
			uri:    "titan://a.b/;size=5",
			reader: bytes.NewBuffer([]byte{1, 2, 3}),
			handlerHook: func(c Context, t *testing.T) error {
				is := is.New(t)
				b, err := TitanReadFull(c)
				is.True(errors.Is(err, io.ErrUnexpectedEOF))
				is.True(b != nil)
				is.Equal(len(b), 5)
				is.Equal(b, []byte{1, 2, 3, 0, 0})
				return nil
			},
		},
		{
			name:   "stop reading at size",
			uri:    "titan://a.b/;size=3",
			reader: bytes.NewBuffer([]byte{1, 2, 3, 4, 5}),
			handlerHook: func(c Context, t *testing.T) error {
				is := is.New(t)
				b, err := TitanReadFull(c)
				is.NoErr(err)
				is.True(b != nil)
				is.Equal(len(b), 3)
				is.Equal(b, []byte{1, 2, 3})
				return nil
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)
			g := New()
			g.Handle("/*", func(c Context) error {
				if tc.handlerHook != nil {
					return tc.handlerHook(c, t)
				}
				return nil
			})
			g.Use(Titan(tc.sizeLimit))
			ctx, conn := g.NewFakeContext(
				tc.uri,
				nil,
				WithFakeReader(tc.reader),
			)
			g.ServeGemini(ctx)
			is.Equal(tc.expect, conn.Written)
		})
	}
}

func TestTitanRedirect(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		result  string
		wantErr bool
	}{
		{
			name:   "good url",
			url:    "titan://f.org/raw/Test;token=hello;mime=plain/text;size=10",
			result: "gemini://f.org/raw/Test",
		},
		{
			name:    "no fragments",
			url:     "titan://f.org/foo",
			result:  "gemini://f.org/foo",
			wantErr: false,
		},
		{
			name:    "bad fragments",
			url:     "titan://f.org/foo;a=1;b=2",
			result:  "gemini://f.org/foo",
			wantErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			is := is.New(t)
			url, err := url.Parse(tc.url)
			is.NoErr(err)
			err = titanURLtoGemini(url)
			if !tc.wantErr {
				is.NoErr(err)
				is.Equal(tc.result, url.String())
			} else {
				t.Log(url)
				is.True(err != nil)
			}
		})
	}
}
