package generator

import (
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestStringUrl(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v any, err error)
	}{
		{
			name: "uri",
			req: &Request{
				Path:   []string{"uri"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.headenvisioneer.io/24-7/mesh/functionalities", v)
			},
		},
		{
			name: "uri no schema",
			req: &Request{
				Path: []string{"uri"},
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.headenvisioneer.io/24-7/mesh/functionalities", v)
			},
		},
		{
			name: "url",
			req: &Request{
				Path:   []string{"url"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.headenvisioneer.io/24-7/mesh/functionalities", v)
			},
		},
		{
			name: "url no schema",
			req: &Request{
				Path:   []string{"url"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.headenvisioneer.io/24-7/mesh/functionalities", v)
			},
		},
		{
			name: "curl no schema should not be a URL",
			req: &Request{
				Path: []string{"curl"},
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.InDelta(t, -170715.30581115812, v, 0.000001)
			},
		},
		{
			name: "updateUrl",
			req: &Request{
				Path:   []string{"url"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.headenvisioneer.io/24-7/mesh/functionalities", v)
			},
		},
		{
			name: "updateURL",
			req: &Request{
				Path:   []string{"updateURL"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.headenvisioneer.io/24-7/mesh/functionalities", v)
			},
		},
		{
			name: "url with schema array",
			req: &Request{
				Path: []string{"url"},
				Schema: schematest.New("array",
					schematest.WithMinItems(1),
					schematest.WithItems("string"),
				),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, []any{
					"http://www.directend-to-end.com/mesh",
				}, v)
			},
		},
		{
			name: "photoUrls",
			req: &Request{
				Path: []string{"photoUrls"},
				Schema: schematest.New("array",
					schematest.WithMinItems(1),
					schematest.WithItems("string"),
				),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, []any{
					"http://www.directend-to-end.com/mesh",
				}, v)
			},
		},
		{
			name: "urls no schema",
			req: &Request{
				Path:   []string{"urls"},
				Schema: schematest.NewTypes(nil, schematest.WithMinItems(1)),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, []any{"http://www.directend-to-end.com/mesh"}, v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)
			Seed(1234567)

			v, err := New(tc.req)
			tc.test(t, v, err)
		})
	}
}
