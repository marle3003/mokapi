package encoding_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/media"
	"mokapi/schema/encoding"
	"testing"
)

func TestMultipartDecoder(t *testing.T) {
	testcases := []struct {
		name string
		data []byte
		opts []encoding.DecodeOptions
		test func(t *testing.T, v any, err error)
	}{
		{
			name: "with valid data",
			data: []byte(`--abcde12345
Content-Disposition: form-data; name="id"
Content-Type: text/plain

123e4567-e89b-12d3-a456-426655440000
--abcde12345
Content-Disposition: form-data; name="address"
Content-Type: application/json

{
  "street": "3, Garden St",
  "city": "Hillsbery, UT"
}
--abcde12345--
`),
			opts: []encoding.DecodeOptions{
				encoding.WithContentType(media.ParseContentType("multipart/form-data; boundary=\"abcde12345\"")),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]any{
						"id": "123e4567-e89b-12d3-a456-426655440000",
						"address": map[string]any{
							"city":   "Hillsbery, UT",
							"street": "3, Garden St",
						},
					}, v)
			},
		},
		{
			name: "with custom decoding function",
			data: []byte(`key1=value1`),
			opts: []encoding.DecodeOptions{
				encoding.WithContentType(media.ParseContentType("application/x-www-form-urlencoded")),
				encoding.WithDecodeFormUrlParam(func(name string, value interface{}) (interface{}, error) {
					return "custom", nil
				}),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]any{
						"key1": "custom",
					}, v)
			},
		},
		{
			name: "with custom decoding function that returns an error",
			data: []byte(`key1=value1`),
			opts: []encoding.DecodeOptions{
				encoding.WithContentType(media.ParseContentType("application/x-www-form-urlencoded")),
				encoding.WithDecodeFormUrlParam(func(name string, value interface{}) (interface{}, error) {
					return nil, fmt.Errorf("TEST ERROR")
				}),
			},
			test: func(t *testing.T, v any, err error) {
				require.EqualError(t, err, "TEST ERROR")
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
