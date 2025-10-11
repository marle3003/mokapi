package encoding_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/media"
	"mokapi/schema/encoding"
	"mokapi/schema/json/schema"
	"testing"
)

func TestXmlEncoder_Encode(t *testing.T) {
	testcases := []struct {
		name string
		s    *schema.Schema
		data any
		test func(t *testing.T, s string, err error)
	}{
		{
			name: "string",
			data: "hello world",
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<data>hello world</data>`, s)
			},
		},
		{
			name: "object",
			data: map[string]any{"foo": "bar", "baz": 42},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<data><foo>bar</foo><baz>42</baz></data>`, s)
			},
		},
		{
			name: "array",
			data: []any{1, 2, 3},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<data><items>1</items><items>2</items><items>3</items></data>`, s)
			},
		},
		{
			name: "array with $id",
			data: []any{1, 2, 3},
			s:    &schema.Schema{Id: "/foo"},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>
<foo><items>1</items><items>2</items><items>3</items></foo>`, s)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			e := encoding.NewEncoder(tc.s)
			b, err := e.Write(tc.data, media.ParseContentType("application/xml"))
			tc.test(t, string(b), err)
		})
	}
}
