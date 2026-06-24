package generator

import (
	"fmt"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestPattern(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "pattern numbers",
			req: &Request{
				Schema: schematest.New("string",
					schematest.WithPattern("[0-9]+"),
					schematest.WithMinLength(10),
					schematest.WithMaxLength(15),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "8026409319", v)
			},
		},
		{
			name: "pattern with min length but cannot reach min length",
			req: &Request{
				Schema: schematest.New("string",
					schematest.WithPattern("[0-5]{1,4}"),
					schematest.WithMinLength(10),
					schematest.WithMaxLength(15),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "cannot generate value for pattern [0-5]{1,4} and minimum length 10")
			},
		},
		{
			name: "pattern with start/end and min/max",
			req: &Request{
				Schema: schematest.New("string",
					schematest.WithPattern("^[a-z]+[A-Z0-9_]+$"),
					schematest.WithMinLength(8),
					schematest.WithMaxLength(20),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "cqopBLWT2", v)
			},
		},
		{
			name: "pattern repeat",
			req: &Request{
				Schema: schematest.New("string",
					schematest.WithPattern("[[0-5]{1,4}"),
					schematest.WithMinLength(4),
					schematest.WithMaxLength(4),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2135", v)
			},
		},
		{
			name: "pattern repeat but cannot reach min length",
			req: &Request{
				Schema: schematest.New("string",
					schematest.WithPattern("[0-5]{1,4}"),
					schematest.WithMinLength(5),
					schematest.WithMaxLength(4),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "cannot generate value for pattern [0-5]{1,4} and minimum length 5")
			},
		},
		{
			name: "complex pattern",
			req: &Request{
				Schema: schematest.New("string",
					schematest.WithPattern("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "auRQ@H\\?TtUNg72G7\\XusB", v)
			},
		},
		{
			name: "complex pattern with minLength",
			req: &Request{
				Schema: schematest.New("string",
					schematest.WithPattern("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"),
					schematest.WithMinLength(50),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.GreaterOrEqual(t, len(fmt.Sprintf("%v", v)), 50)
				require.Equal(t, "auRQ@UKCH\\?TtUNwKXfFv\\)AHqmk9jx7n\\*c\\*II5MgHbQbY38QWug72G7\\XusB", v)
			},
		},
		{
			name: "fix length",
			req: &Request{
				Schema: schematest.New("string",
					schematest.WithPattern("\\d{16}"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Len(t, v, 16)
				require.Equal(t, "8029109364893011", v)
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
