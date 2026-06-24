package generator

import (
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestStringDescription(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "description",
			req: &Request{
				Path:   []string{"description"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Protect the hand under cute load.", v)
			},
		},
		{
			name: "description with max length",
			req: &Request{
				Path:   []string{"description"},
				Schema: schematest.New("string", schematest.WithMaxLength(50)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				s := v.(string)
				require.Less(t, len(s), 51)
				require.Equal(t, "Protect the hand under cute load.", v)
			},
		},
		{
			name: "message",
			req: &Request{
				Path:   []string{"message"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Protect the hand under cute load.", v)
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

func TestCategory(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "category",
			req: &Request{
				Path:   []string{"category"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Apps", v)
			},
		},
		{
			name: "category with max length",
			req: &Request{
				Path:   []string{"category"},
				Schema: schematest.New("string", schematest.WithMaxLength(5)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Tech", v)
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
