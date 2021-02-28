package web

import (
	"mokapi/models/rest"
	"reflect"
	"testing"
)

func TestParsePath(t *testing.T) {
	data := []struct {
		s string
		p *rest.Parameter
		e interface{}
	}{
		{"foo",
			&rest.Parameter{
				Schema:  &rest.Schema{Type: "string"},
				Style:   "",
				Explode: false,
			},
			"foo",
		},
		{".foo",
			&rest.Parameter{
				Schema:  &rest.Schema{Type: "string"},
				Style:   "label",
				Explode: false,
			},
			"foo",
		},
		{";foo",
			&rest.Parameter{
				Schema:  &rest.Schema{Type: "string"},
				Style:   "matrix",
				Explode: false,
			},
			"foo",
		},
		{"3,4,5",
			&rest.Parameter{
				Schema:  &rest.Schema{Type: "array", Items: &rest.Schema{Type: "integer"}},
				Style:   "",
				Explode: false,
			},
			[]interface{}{3, 4, 5},
		},
		{".3,4,5",
			&rest.Parameter{
				Schema:  &rest.Schema{Type: "array", Items: &rest.Schema{Type: "integer"}},
				Style:   "label",
				Explode: false,
			},
			[]interface{}{3, 4, 5},
		},
		{";3,4,5",
			&rest.Parameter{
				Schema:  &rest.Schema{Type: "array", Items: &rest.Schema{Type: "integer"}},
				Style:   "matrix",
				Explode: false,
			},
			[]interface{}{3, 4, 5},
		},
		{"role,admin,firstName,Alex",
			&rest.Parameter{
				Schema: &rest.Schema{Type: "object",
					Properties: map[string]*rest.Schema{
						"role":      &rest.Schema{Type: "string"},
						"firstName": &rest.Schema{Type: "string"},
					}},
				Style:   "",
				Explode: false,
			},
			map[string]interface{}{"role": "admin", "firstName": "Alex"},
		},
		{"role=admin,firstName=Alex",
			&rest.Parameter{
				Schema: &rest.Schema{Type: "object",
					Properties: map[string]*rest.Schema{
						"role":      &rest.Schema{Type: "string"},
						"firstName": &rest.Schema{Type: "string"},
					}},
				Style:   "",
				Explode: true,
			},
			map[string]interface{}{"role": "admin", "firstName": "Alex"},
		},
		{".role,admin,firstName,Alex",
			&rest.Parameter{
				Schema: &rest.Schema{Type: "object",
					Properties: map[string]*rest.Schema{
						"role":      &rest.Schema{Type: "string"},
						"firstName": &rest.Schema{Type: "string"},
					}},
				Style:   "label",
				Explode: false,
			},
			map[string]interface{}{"role": "admin", "firstName": "Alex"},
		},
		{";role=admin,firstName=Alex",
			&rest.Parameter{
				Schema: &rest.Schema{Type: "object",
					Properties: map[string]*rest.Schema{
						"role":      &rest.Schema{Type: "string"},
						"firstName": &rest.Schema{Type: "string"},
					}},
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
