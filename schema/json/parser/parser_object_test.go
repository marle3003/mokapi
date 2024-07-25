package parser

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schematest"
	"testing"
)

func TestParser_ParseObject(t *testing.T) {
	testcases := []struct {
		name   string
		data   interface{}
		schema *schema.Schema
		test   func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "required string but it is empty",
			data: map[string]interface{}{"foo": ""},
			schema: schematest.New("object",
				schematest.WithProperty("foo", schematest.New("string")),
				schematest.WithRequired("foo")),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "missing required field 'foo', expected schema type=object properties=[foo] required=[foo]")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &Parser{}
			v, err := p.ParseObject(tc.data, tc.schema)
			tc.test(t, v, err)
		})
	}
}
