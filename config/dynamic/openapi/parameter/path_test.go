package parameter

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"testing"
)

func TestParsePath(t *testing.T) {
	testcases := []struct {
		s string
		p *Parameter
		e interface{}
	}{
		{
			"foo",
			&Parameter{
				Schema:  &schema.Ref{Value: &schema.Schema{Type: "string"}},
				Style:   "",
				Explode: false,
			},
			"foo",
		},
		{
			".foo",
			&Parameter{
				Schema:  &schema.Ref{Value: &schema.Schema{Type: "string"}},
				Style:   "label",
				Explode: false,
			},
			"foo",
		},
		{
			";foo",
			&Parameter{
				Schema:  &schema.Ref{Value: &schema.Schema{Type: "string"}},
				Style:   "matrix",
				Explode: false,
			},
			"foo",
		},
		{
			"3,4,5",
			&Parameter{
				Schema:  &schema.Ref{Value: &schema.Schema{Type: "array", Items: &schema.Ref{Value: &schema.Schema{Type: "integer"}}}},
				Style:   "",
				Explode: false,
			},
			[]interface{}{int64(3), int64(4), int64(5)},
		},
		{
			".3,4,5",
			&Parameter{
				Schema:  &schema.Ref{Value: &schema.Schema{Type: "array", Items: &schema.Ref{Value: &schema.Schema{Type: "integer"}}}},
				Style:   "label",
				Explode: false,
			},
			[]interface{}{int64(3), int64(4), int64(5)},
		},
		{
			";3,4,5",
			&Parameter{
				Schema:  &schema.Ref{Value: &schema.Schema{Type: "array", Items: &schema.Ref{Value: &schema.Schema{Type: "integer"}}}},
				Style:   "matrix",
				Explode: false,
			},
			[]interface{}{int64(3), int64(4), int64(5)},
		},
		{
			"role,admin,firstName,Alex",
			&Parameter{
				Schema: &schema.Ref{Value: schematest.New("object",
					schematest.WithProperty("role", schematest.New("string")),
					schematest.WithProperty("firstName", schematest.New("string")),
				)},
				Style:   "",
				Explode: false,
			},
			map[string]interface{}{"role": "admin", "firstName": "Alex"},
		},
		{
			"role=admin,firstName=Alex",
			&Parameter{
				Schema: &schema.Ref{Value: schematest.New("object",
					schematest.WithProperty("role", schematest.New("string")),
					schematest.WithProperty("firstName", schematest.New("string")),
				)},
				Style:   "",
				Explode: true,
			},
			map[string]interface{}{"role": "admin", "firstName": "Alex"},
		},
		{
			".role,admin,firstName,Alex",
			&Parameter{
				Schema: &schema.Ref{Value: schematest.New("object",
					schematest.WithProperty("role", schematest.New("string")),
					schematest.WithProperty("firstName", schematest.New("string")),
				)},
				Style:   "label",
				Explode: false,
			},
			map[string]interface{}{"role": "admin", "firstName": "Alex"},
		},
		{
			";role=admin,firstName=Alex",
			&Parameter{
				Schema: &schema.Ref{Value: schematest.New("object",
					schematest.WithProperty("role", schematest.New("string")),
					schematest.WithProperty("firstName", schematest.New("string")),
				)},
				Style:   "matrix",
				Explode: true,
			},

			map[string]interface{}{"role": "admin", "firstName": "Alex"},
		},
	}

	for _, testcase := range testcases {
		test := testcase
		t.Run(test.s, func(t *testing.T) {
			i, err := parsePath(test.s, test.p)
			require.NoError(t, err)
			require.Equal(t, test.e, i.Value)
		})
	}
}
