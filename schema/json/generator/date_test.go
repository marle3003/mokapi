package generator

import (
	"mokapi/schema/json/schema/schematest"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
)

func TestStringDate(t *testing.T) {
	// tests depends on current year so without this, all tests will break in next year
	isDateString := func(t *testing.T, s any) {
		_, err := time.Parse("2006-01-02", s.(string))
		require.NoError(t, err)
	}
	isDateTimeString := func(t *testing.T, s any) {
		_, err := time.Parse(time.RFC3339, s.(string))
		require.NoError(t, err)
	}

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
				isDateString(t, v)
			},
		},
		{
			name: "createdAt",
			req: &Request{
				Path: []string{"createdAt"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				isDateString(t, v)
			},
		},
		{
			name: "creationDate",
			req: &Request{
				Path: []string{"creationDate"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				isDateString(t, v)
			},
		},
		{
			name: "modified",
			req: &Request{
				Path: []string{"modified"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				isDateString(t, v)
			},
		},
		{
			name: "modifyDate",
			req: &Request{
				Path: []string{"modifyDate"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				isDateString(t, v)
			},
		},
		{
			name: "updateDate",
			req: &Request{
				Path: []string{"updateDate"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				isDateString(t, v)
			},
		},
		{
			name: "deleted",
			req: &Request{
				Path: []string{"deleted"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				isDateString(t, v)
			},
		},
		{
			name: "deletedAt",
			req: &Request{
				Path: []string{"deletedAt"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				isDateString(t, v)
			},
		},
		{
			name: "deleteDate",
			req: &Request{
				Path: []string{"deleteDate"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				isDateString(t, v)
			},
		},
		{
			name: "foundationDate",
			req: &Request{
				Path: []string{"foundationDate"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				isDateString(t, v)
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
				m := v.(map[string]interface{})
				isDateTimeString(t, m["activeFrom"])
				isDateTimeString(t, m["inactiveFrom"])
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
				m := v.(map[string]interface{})
				isDateTimeString(t, m["publishedFrom"])
				isDateTimeString(t, m["publishedUntil"])
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
