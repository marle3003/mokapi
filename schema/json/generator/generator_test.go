package generator_test

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/generator"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestNew(t *testing.T) {
	testcases := []struct {
		name    string
		request *generator.Request
		test    func(t *testing.T, v interface{}, err error)
	}{
		{
			name:    "no schema",
			request: &generator.Request{},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(-3652171958352792229), v)
			},
		},
		{
			name:    "schema null",
			request: generator.NewRequest(generator.UsePathElement("", schematest.New("null"))),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, nil, v)
			},
		},
		{
			name: "user with id in context",
			request: generator.NewRequest(
				generator.UsePathElement("user",
					schematest.New("object",
						schematest.WithProperty("id", schematest.New("integer")),
						schematest.WithProperty("name", schematest.New("string")),
						schematest.WithProperty("email", schematest.New("string")),
					),
				),
				generator.UseContext(map[string]interface{}{"id": 123}),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"email": "kyliemcclure@kiehn.name", "id": int64(123), "name": "Lockman7291"}, v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)

			v, err := generator.New(tc.request)
			tc.test(t, v, err)
		})
	}
}
