package encoding_test

import (
	"fmt"
	"mokapi/media"
	"mokapi/schema/encoding"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormUrlEncodeDecoder(t *testing.T) {
	testcases := []struct {
		name string
		data []byte
		opts []encoding.DecodeOptions
		test func(t *testing.T, v any, err error)
	}{
		{
			name: "with valid data",
			data: []byte(`key1=value1&key2=value2&key3=value3`),
			opts: []encoding.DecodeOptions{
				encoding.WithContentType(media.ParseContentType("application/x-www-form-urlencoded")),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]any{
						"key1": []string{"value1"},
						"key2": []string{"value2"},
						"key3": []string{"value3"},
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
