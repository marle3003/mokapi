package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestStringDate(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "created",
			req: &Request{
				Path:   []string{"created"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2020-08-26", v)
			},
		},
		{
			name: "createdAt",
			req: &Request{
				Path: []string{"createdAt"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2020-08-26", v)
			},
		},
		{
			name: "creationDate",
			req: &Request{
				Path: []string{"creationDate"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2020-08-26", v)
			},
		},
		{
			name: "modified",
			req: &Request{
				Path: []string{"modified"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2020-08-26", v)
			},
		},
		{
			name: "modifyDate",
			req: &Request{
				Path: []string{"modifyDate"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2020-08-26", v)
			},
		},
		{
			name: "updateDate",
			req: &Request{
				Path: []string{"updateDate"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2020-08-26", v)
			},
		},
		{
			name: "deleted",
			req: &Request{
				Path: []string{"deleted"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2020-08-26", v)
			},
		},
		{
			name: "deletedAt",
			req: &Request{
				Path: []string{"deletedAt"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2020-08-26", v)
			},
		},
		{
			name: "deleteDate",
			req: &Request{
				Path: []string{"deleteDate"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2020-08-26", v)
			},
		},
		{
			name: "foundationDate",
			req: &Request{
				Path: []string{"foundationDate"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "1970-08-26", v)
			},
		},
		{
			name: "activeFrom and inactiveFrom",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("activeFrom", schematest.New("string", schematest.WithFormat("date-time"))),
					schematest.WithProperty("inactiveFrom", schematest.New("string", schematest.WithFormat("date-time"))),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{"activeFrom": "2020-08-10T23:06:20Z", "inactiveFrom": "2026-09-12T04:37:26Z"}, v)
			},
		},
		{
			name: "publishedFrom and publishedUntil",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("publishedFrom", schematest.New("string", schematest.WithFormat("date-time"))),
					schematest.WithProperty("publishedUntil", schematest.New("string", schematest.WithFormat("date-time"))),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{"publishedFrom": "2020-08-10T23:06:20Z", "publishedUntil": "2026-09-12T04:37:26Z"}, v)
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
