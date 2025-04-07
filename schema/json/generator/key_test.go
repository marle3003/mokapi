package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestKey(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "businessKey",
			req: &Request{
				Path:   []string{"businessKey"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "44f4ae5d-233e-4f89-ae02-126591065f49", v)
			},
		},
		{
			name: "key with pattern",
			req: &Request{
				Path: []string{"businessKey"},
				Schema: schematest.New("string",
					schematest.WithPattern("[a-z]{3}"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "cqo", v)
			},
		},
		{
			name: "key string with min",
			req: &Request{
				Path:   []string{"key"},
				Schema: schematest.New("string", schematest.WithMinLength(4)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "44f4ae5d-233e-4f89-ae02-126591065f49", v)
			},
		},
		{
			name: "key string with min & max",
			req: &Request{
				Path: []string{"key"},
				Schema: schematest.New("string",
					schematest.WithMinLength(4),
					schematest.WithMaxLength(10),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "7291093", v)
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
