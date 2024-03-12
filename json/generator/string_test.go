package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/json/schematest"
	"testing"
)

func TestString(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "language",
			req: &Request{
				Names: []string{"language"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "nl-BE", v)
			},
		},
		{
			name: "lang maxLenghth = 2",
			req: &Request{
				Names: []string{"language"},
				Schema: schematest.New("string",
					schematest.WithMaxLength(2),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "sm", v)
			},
		},
		{
			name: "lang maxLenghth = 5",
			req: &Request{
				Names: []string{"language"},
				Schema: schematest.New("string",
					schematest.WithMaxLength(5),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "nl-BE", v)
			},
		},
		{
			name: "lang maxLength = 15",
			req: &Request{
				Names: []string{"language"},
				Schema: schematest.New("string",
					schematest.WithMaxLength(15),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Samoan", v)
			},
		},
		{
			name: "langs",
			req: &Request{
				Names: []string{"langs"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"pt-BR", "pt-PT"}, v)
			},
		},
		{
			name: "error",
			req: &Request{
				Names: []string{"error"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "named cookie not present", v)
			},
		},
		{
			name: "partyNumber",
			req: &Request{
				Names:  []string{"partyNumber"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "80291093648", v)
			},
		},
		{
			name: "employeeNumber with equal min max",
			req: &Request{
				Names:  []string{"employeeNumber"},
				Schema: schematest.New("string", schematest.WithMinLength(4), schematest.WithMaxLength(4)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "8029", v)
			},
		},
		{
			name: "employeeNumber with range",
			req: &Request{
				Names:  []string{"employeeNumber"},
				Schema: schematest.New("string", schematest.WithMinLength(4), schematest.WithMaxLength(30)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "129109364893011805735", v)
			},
		},
		{
			name: "email",
			req: &Request{
				Names:  []string{"email"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "shanellewehner@cruickshank.biz", v)
			},
		},
		{
			name: "email with min value",
			req: &Request{
				Names: []string{"email"},
				Schema: schematest.New("string",
					schematest.WithMinLength(20),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "shanellewehner@cruickshank.biz", v)
			},
		},
		{
			name: "email with max value but to short",
			req: &Request{
				Names: []string{"email"},
				Schema: schematest.New("string",
					schematest.WithMaxLength(5),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "sEOoL", v)
			},
		},
		{
			name: "hash",
			req: &Request{
				Names:  []string{"hash"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "54686520636f6d70616e7920736d656c6c2eda39a3ee5e6b4b0d3255bfef95601890afd80709", v)
			},
		},
		{
			name: "houseNumber",
			req: &Request{
				Names: []string{"houseNumber"},
				Schema: schematest.New("string",
					schematest.WithMaxLength(4)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "7", v)
			},
		},
		{
			name: "description",
			req: &Request{
				Names:  []string{"description"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Tonight daily regularly you for recline her choir insufficient poised tribe Caesarian mustering will might.", v)
			},
		},
		{
			name: "currency",
			req: &Request{
				Names:  []string{"currency"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "BYR", v)
			},
		},
		{
			name: "color",
			req: &Request{
				Names:  []string{"color"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "LightYellow", v)
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
