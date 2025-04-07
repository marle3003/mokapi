package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestStringUrl(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "uri",
			req: &Request{
				Path:   []string{"uri"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.regionale-enable.info/virtual/portals/redefine", v)
			},
		},
		{
			name: "uri no schema",
			req: &Request{
				Path: []string{"uri"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.regionale-enable.info/virtual/portals/redefine", v)
			},
		},
		{
			name: "url",
			req: &Request{
				Path:   []string{"url"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.regionale-enable.info/virtual/portals/redefine", v)
			},
		},
		{
			name: "url no schema",
			req: &Request{
				Path:   []string{"url"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.regionale-enable.info/virtual/portals/redefine", v)
			},
		},
		{
			name: "curl no schema should not be a URL",
			req: &Request{
				Path: []string{"curl"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(-3652171958352792229), v)
			},
		},
		{
			name: "updateUrl",
			req: &Request{
				Path:   []string{"url"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.regionale-enable.info/virtual/portals/redefine", v)
			},
		},
		{
			name: "updateURL",
			req: &Request{
				Path:   []string{"updateURL"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://www.regionale-enable.info/virtual/portals/redefine", v)
			},
		},
		{
			name: "url with schema array",
			req: &Request{
				Path:   []string{"url"},
				Schema: schematest.New("array", schematest.WithItems("string")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{
					"http://www.dynamicembrace.name/portals/redefine/deliver/cultivate",
					"http://www.dynamiccommunities.io/embrace/frictionless/deploy/granular",
				}, v)
			},
		},
		{
			name: "photoUrls",
			req: &Request{
				Path:   []string{"photoUrls"},
				Schema: schematest.New("array", schematest.WithItems("string")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{
					"http://www.dynamicembrace.name/portals/redefine/deliver/cultivate",
					"http://www.dynamiccommunities.io/embrace/frictionless/deploy/granular",
				}, v)
			},
		},
		{
			name: "urls no schema",
			req: &Request{
				Path: []string{"urls"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{
					"http://www.dynamicembrace.name/portals/redefine/deliver/cultivate",
					"http://www.dynamiccommunities.io/embrace/frictionless/deploy/granular",
				}, v)
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
