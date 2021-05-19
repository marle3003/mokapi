package web

import (
	"mokapi/config/dynamic/openapi"
	"reflect"
	"testing"
)

func TestParsePath(t *testing.T) {
	data := []struct {
		s string
		p *openapi.Parameter
		e interface{}
	}{
		{
			"foo",
			&openapi.Parameter{
				Schema:  &openapi.SchemaRef{Value: &openapi.Schema{Type: "string"}},
				Style:   "",
				Explode: false,
			},
			"foo",
		},
		{
			".foo",
			&openapi.Parameter{
				Schema:  &openapi.SchemaRef{Value: &openapi.Schema{Type: "string"}},
				Style:   "label",
				Explode: false,
			},
			"foo",
		},
		{
			";foo",
			&openapi.Parameter{
				Schema:  &openapi.SchemaRef{Value: &openapi.Schema{Type: "string"}},
				Style:   "matrix",
				Explode: false,
			},
			"foo",
		},
		{
			"3,4,5",
			&openapi.Parameter{
				Schema:  &openapi.SchemaRef{Value: &openapi.Schema{Type: "array", Items: &openapi.SchemaRef{Value: &openapi.Schema{Type: "integer"}}}},
				Style:   "",
				Explode: false,
			},
			[]interface{}{3, 4, 5},
		},
		{
			".3,4,5",
			&openapi.Parameter{
				Schema:  &openapi.SchemaRef{Value: &openapi.Schema{Type: "array", Items: &openapi.SchemaRef{Value: &openapi.Schema{Type: "integer"}}}},
				Style:   "label",
				Explode: false,
			},
			[]interface{}{3, 4, 5},
		},
		{
			";3,4,5",
			&openapi.Parameter{
				Schema:  &openapi.SchemaRef{Value: &openapi.Schema{Type: "array", Items: &openapi.SchemaRef{Value: &openapi.Schema{Type: "integer"}}}},
				Style:   "matrix",
				Explode: false,
			},
			[]interface{}{3, 4, 5},
		},
		{
			"role,admin,firstName,Alex",
			&openapi.Parameter{
				Schema: &openapi.SchemaRef{Value: &openapi.Schema{Type: "object",
					Properties: &openapi.Schemas{
						Value: map[string]*openapi.SchemaRef{
							"role":      {Value: &openapi.Schema{Type: "string"}},
							"firstName": {Value: &openapi.Schema{Type: "string"}},
						}}}},
				Style:   "",
				Explode: false,
			},
			map[string]interface{}{"role": "admin", "firstName": "Alex"},
		},
		{
			"role=admin,firstName=Alex",
			&openapi.Parameter{
				Schema: &openapi.SchemaRef{Value: &openapi.Schema{Type: "object",
					Properties: &openapi.Schemas{
						Value: map[string]*openapi.SchemaRef{
							"role":      {Value: &openapi.Schema{Type: "string"}},
							"firstName": {Value: &openapi.Schema{Type: "string"}},
						}}}},
				Style:   "",
				Explode: true,
			},
			map[string]interface{}{"role": "admin", "firstName": "Alex"},
		},
		{
			".role,admin,firstName,Alex",
			&openapi.Parameter{
				Schema: &openapi.SchemaRef{Value: &openapi.Schema{Type: "object",
					Properties: &openapi.Schemas{
						Value: map[string]*openapi.SchemaRef{
							"role":      {Value: &openapi.Schema{Type: "string"}},
							"firstName": {Value: &openapi.Schema{Type: "string"}},
						}}}},
				Style:   "label",
				Explode: false,
			},
			map[string]interface{}{"role": "admin", "firstName": "Alex"},
		},
		{
			";role=admin,firstName=Alex",
			&openapi.Parameter{
				Schema: &openapi.SchemaRef{Value: &openapi.Schema{Type: "object",
					Properties: &openapi.Schemas{
						Value: map[string]*openapi.SchemaRef{
							"role":      {Value: &openapi.Schema{Type: "string"}},
							"firstName": {Value: &openapi.Schema{Type: "string"}},
						}}}},
				Style:   "matrix",
				Explode: true,
			},

			map[string]interface{}{"role": "admin", "firstName": "Alex"},
		},
	}

	for _, d := range data {
		i, err := parsePath(d.s, d.p)
		if err != nil {
			t.Errorf("parsePath(%v): %v", d.s, err)
		} else if !reflect.DeepEqual(d.e, i) {
			t.Errorf("parsePath(%v): got %v; expected %v", d.s, i, d.e)
		}
	}
}
