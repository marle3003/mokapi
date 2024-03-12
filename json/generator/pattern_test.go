package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/json/schematest"
	"testing"
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
				require.Equal(t, "80364891092", v)
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
				require.Equal(t, "cqoABWTL", v)
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
				require.Equal(t, "1510", v)
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
				require.Equal(t, "auR@j\\Cc\\8J\\UWN9\\AeG7k\\AzsB5UK7vwW", v)
			},
		},
		{
			name: "complex pattern",
			req: &Request{
				Schema: schematest.New("string",
					schematest.WithPattern("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"),
					schematest.WithMinLength(50),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "au/R@j\\fF38QbCuGvgHbbYW1THqmk9Qcj\\*cAHT2\\uMeKvpeve4x7M\\8J\\UWN9\\AeG7k\\AzsB5UK7vwW", v)
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
