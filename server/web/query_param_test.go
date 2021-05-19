package web

import (
	"mokapi/config/dynamic/openapi"
	"net/url"
	"reflect"
	"testing"
)

func TestParseQuery(t *testing.T) {
	data := []struct {
		u *url.URL
		p *openapi.Parameter
		e interface{}
	}{
		{&url.URL{RawQuery: "id=5"},
			&openapi.Parameter{
				Name:    "id",
				Schema:  &openapi.SchemaRef{Value: &openapi.Schema{Type: "integer"}},
				Style:   "",
				Explode: false,
			},
			5,
		},
		{&url.URL{RawQuery: ""},
			&openapi.Parameter{
				Name:    "id",
				Schema:  &openapi.SchemaRef{Value: &openapi.Schema{Type: "integer"}},
				Style:   "",
				Explode: false,
			},
			nil,
		},
		{&url.URL{RawQuery: "id=3&id=4&id=5"},
			&openapi.Parameter{
				Name:    "id",
				Schema:  &openapi.SchemaRef{Value: &openapi.Schema{Type: "array", Items: &openapi.SchemaRef{Value: &openapi.Schema{Type: "integer"}}}},
				Style:   "",
				Explode: true,
			},
			[]interface{}{3, 4, 5},
		},
		{&url.URL{RawQuery: "id=3,4,5"},
			&openapi.Parameter{
				Name:    "id",
				Schema:  &openapi.SchemaRef{Value: &openapi.Schema{Type: "array", Items: &openapi.SchemaRef{Value: &openapi.Schema{Type: "integer"}}}},
				Style:   "",
				Explode: false,
			},
			[]interface{}{3, 4, 5},
		},
		{&url.URL{RawQuery: "id=3%204%205"},
			&openapi.Parameter{
				Name:    "id",
				Schema:  &openapi.SchemaRef{Value: &openapi.Schema{Type: "array", Items: &openapi.SchemaRef{Value: &openapi.Schema{Type: "integer"}}}},
				Style:   "spaceDelimited",
				Explode: false,
			},
			[]interface{}{3, 4, 5},
		},
		{&url.URL{RawQuery: "id=3|4|5"},
			&openapi.Parameter{
				Name:    "id",
				Schema:  &openapi.SchemaRef{Value: &openapi.Schema{Type: "array", Items: &openapi.SchemaRef{Value: &openapi.Schema{Type: "integer"}}}},
				Style:   "pipeDelimited",
				Explode: false,
			},
			[]interface{}{3, 4, 5},
		},
		{&url.URL{RawQuery: "role=admin&firstName=Alex"},
			&openapi.Parameter{
				Name: "id",
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
		{&url.URL{RawQuery: "id=role,admin,firstName,Alex"},
			&openapi.Parameter{
				Name: "id",
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
		{&url.URL{RawQuery: "id[role]=admin&id[firstName]=Alex"},
			&openapi.Parameter{
				Name: "id",
				Schema: &openapi.SchemaRef{Value: &openapi.Schema{Type: "object",
					Properties: &openapi.Schemas{
						Value: map[string]*openapi.SchemaRef{
							"role":      {Value: &openapi.Schema{Type: "string"}},
							"firstName": {Value: &openapi.Schema{Type: "string"}},
						}}}},
				Style:   "deepObject",
				Explode: true,
			},
			map[string]interface{}{"role": "admin", "firstName": "Alex"},
		},
	}

	for _, d := range data {
		i, err := parseQuery(d.p, d.u)
		if err != nil {
			t.Errorf("parseQuery(%v): %v", d.u, err)
		} else if !reflect.DeepEqual(d.e, i) {
			t.Errorf("parsePath(%v): got %v; expected %v", d.u, i, d.e)
		}
	}
}
