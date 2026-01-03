package generator

import (
	"mokapi/schema/json/schema/schematest"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
)

func TestFile(t *testing.T) {
	// tests depends on current year so without this, all tests will break in next year
	isDateString := func(t *testing.T, s any) {
		_, err := time.Parse("2006-01-02", s.(string))
		require.NoError(t, err)
	}

	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "file name",
			req: &Request{
				Path:   []string{"fileName"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "ink.pl", v)
			},
		},
		{
			name: "file type",
			req: &Request{
				Path: []string{"file", "type"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "application/pkix-crl", v)
			},
		},
		{
			name: "created",
			req: &Request{
				Path: []string{"file", "created"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				isDateString(t, v)
			},
		},
		{
			name: "created bool",
			req: &Request{
				Path:   []string{"file", "created"},
				Schema: schematest.New("boolean"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, false, v)
			},
		},
		{
			name: "createdAt",
			req: &Request{
				Path: []string{"file", "createdAt"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				isDateString(t, v)
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
