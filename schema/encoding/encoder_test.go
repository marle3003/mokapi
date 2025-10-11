package encoding

import (
	"github.com/stretchr/testify/require"
	"mokapi/media"
	"mokapi/schema/json/schema"
	"testing"
)

func TestEncode(t *testing.T) {
	testcases := []struct {
		name   string
		data   any
		schema *schema.Schema
		ct     media.ContentType
		test   func(t *testing.T, b []byte, err error)
	}{
		{
			name: "json",
			data: map[string]any{"foo": "bar"},
			ct:   media.ParseContentType("application/json"),
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"foo":"bar"}`, string(b))
			},
		},
		{
			name: "text",
			data: "hello world",
			ct:   media.ParseContentType("text/plain"),
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, `hello world`, string(b))
			},
		},
		{
			name: "float as text",
			data: 12.3,
			ct:   media.ParseContentType("text/plain"),
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, `12.3`, string(b))
			},
		},
		{
			name: "int as text",
			data: 123,
			ct:   media.ParseContentType("text/plain"),
			test: func(t *testing.T, b []byte, err error) {
				require.NoError(t, err)
				require.Equal(t, `123`, string(b))
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := NewEncoder(tc.schema)
			b, err := e.Write(tc.data, tc.ct)
			tc.test(t, b, err)
		})
	}
}
