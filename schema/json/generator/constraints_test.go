package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestConstraints(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "enum",
			req: &Request{
				Schema: schematest.New("string",
					schematest.WithEnum([]any{"foo", "bar"}),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
			},
		},
		{
			name: "enum with dependsOn",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("sex", schematest.New("string",
						schematest.WithEnum([]any{"female", "male"}),
					)),
					schematest.WithProperty("name", schematest.New("string")),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{
					"name": "Zoey Nguyen",
					"sex":  "female",
				}, v)
			},
		},
		{
			name: "const",
			req: &Request{
				Schema: schematest.New("string",
					schematest.WithConst("foo"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
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
