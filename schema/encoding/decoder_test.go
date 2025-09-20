package encoding_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"mokapi/media"
	"mokapi/schema/encoding"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema/schematest"
	"strings"
	"testing"
)

func TestDecode(t *testing.T) {
	testcases := []struct {
		name string
		data []byte
		opts []encoding.DecodeOptions
		test func(t *testing.T, v any, err error)
	}{
		{
			name: "no content type and no schema specified",
			data: []byte(`{"foo":"bar"}`),
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, []byte(`{"foo":"bar"}`), v)
			},
		},
		{
			name: "no content type should return error because data is an array of bytes and not object",
			data: []byte(`{"foo":"bar"}`),
			opts: []encoding.DecodeOptions{
				encoding.WithParser(
					&parser.Parser{Schema: schematest.New("object",
						schematest.WithProperty("foo", schematest.New("string")),
					)},
				),
			},
			test: func(t *testing.T, v any, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/type: invalid type, expected object but got array")
			},
		},
		{
			name: "json",
			data: []byte(`{"foo":"bar"}`),
			opts: []encoding.DecodeOptions{
				encoding.WithParser(
					&parser.Parser{Schema: schematest.New("object",
						schematest.WithProperty("foo", schematest.New("string")),
					)},
				),
				encoding.WithContentType(media.ParseContentType("application/json")),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{"foo": "bar"}, v)
			},
		},
		{
			name: "text",
			data: []byte(`hello world`),
			opts: []encoding.DecodeOptions{
				encoding.WithContentType(media.ParseContentType("text/plain")),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, "hello world", v)
			},
		},
		{
			name: "binary",
			data: []byte{0x01, 0x02, 0xFF, 0x00, 0x10},
			opts: []encoding.DecodeOptions{
				encoding.WithParser(
					&parser.Parser{Schema: schematest.New("string")},
				),
				encoding.WithContentType(media.ParseContentType("application/octet-stream")),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, "\x01\x02\xff\x00\x10", v)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var v any
			var err error
			if tc.opts == nil {
				v, err = encoding.Decode(tc.data)
			} else {
				v, err = encoding.Decode(tc.data, tc.opts...)
			}
			tc.test(t, v, err)
		})
	}
}

func TestDecodeFrom(t *testing.T) {
	testcases := []struct {
		name string
		r    io.Reader
		opts []encoding.DecodeOptions
		test func(t *testing.T, v any, err error)
	}{
		{
			name: "reader is nil",
			r:    nil,
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Nil(t, v)
			},
		},
		{
			name: "reader returns error",
			r: &reader{r: func(p []byte) (n int, err error) {
				return 0, fmt.Errorf("TEST ERROR")
			}},
			test: func(t *testing.T, v any, err error) {
				require.EqualError(t, err, "TEST ERROR")
			},
		},
		{
			name: "json",
			r:    strings.NewReader(`{"foo":"bar"}`),
			opts: []encoding.DecodeOptions{
				encoding.WithParser(
					&parser.Parser{Schema: schematest.New("object",
						schematest.WithProperty("foo", schematest.New("string")),
					)},
				),
				encoding.WithContentType(media.ParseContentType("application/json")),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{"foo": "bar"}, v)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var v any
			var err error
			if tc.opts == nil {
				v, err = encoding.DecodeFrom(tc.r)
			} else {
				v, err = encoding.DecodeFrom(tc.r, tc.opts...)
			}
			tc.test(t, v, err)
		})
	}
}

type reader struct {
	r func(p []byte) (n int, err error)
}

func (r *reader) Read(p []byte) (n int, err error) {
	return r.r(p)
}
