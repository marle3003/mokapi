package web

import (
	"mokapi/models"
	"reflect"
	"testing"
)

func TestParsePath(t *testing.T) {
	data := []struct {
		s string
		p *models.Parameter
		e interface{}
	}{
		{"foo",
			&models.Parameter{
				Schema:  &models.Schema{Type: "string"},
				Style:   "",
				Explode: false,
			},
			"foo",
		},
		{".foo",
			&models.Parameter{
				Schema:  &models.Schema{Type: "string"},
				Style:   "label",
				Explode: false,
			},
			"foo",
		},
		{";foo",
			&models.Parameter{
				Schema:  &models.Schema{Type: "string"},
				Style:   "matrix",
				Explode: false,
			},
			"foo",
		},
		{"3,4,5",
			&models.Parameter{
				Schema:  &models.Schema{Type: "array", Items: &models.Schema{Type: "integer"}},
				Style:   "",
				Explode: false,
			},
			[]interface{}{3, 4, 5},
		},
		{".3,4,5",
			&models.Parameter{
				Schema:  &models.Schema{Type: "array", Items: &models.Schema{Type: "integer"}},
				Style:   "label",
				Explode: false,
			},
			[]interface{}{3, 4, 5},
		},
		{";3,4,5",
			&models.Parameter{
				Schema:  &models.Schema{Type: "array", Items: &models.Schema{Type: "integer"}},
				Style:   "matrix",
				Explode: false,
			},
			[]interface{}{3, 4, 5},
		},
		{"role,admin,firstName,Alex",
			&models.Parameter{
				Schema: &models.Schema{Type: "object",
					Properties: map[string]*models.Schema{
						"role":      &models.Schema{Type: "string"},
						"firstName": &models.Schema{Type: "string"},
					}},
				Style:   "",
				Explode: false,
			},
			map[string]interface{}{"role": "admin", "firstName": "Alex"},
		},
		{"role=admin,firstName=Alex",
			&models.Parameter{
				Schema: &models.Schema{Type: "object",
					Properties: map[string]*models.Schema{
						"role":      &models.Schema{Type: "string"},
						"firstName": &models.Schema{Type: "string"},
					}},
				Style:   "",
				Explode: true,
			},
			map[string]interface{}{"role": "admin", "firstName": "Alex"},
		},
		{".role,admin,firstName,Alex",
			&models.Parameter{
				Schema: &models.Schema{Type: "object",
					Properties: map[string]*models.Schema{
						"role":      &models.Schema{Type: "string"},
						"firstName": &models.Schema{Type: "string"},
					}},
				Style:   "label",
				Explode: false,
			},
			map[string]interface{}{"role": "admin", "firstName": "Alex"},
		},
		{";role=admin,firstName=Alex",
			&models.Parameter{
				Schema: &models.Schema{Type: "object",
					Properties: map[string]*models.Schema{
						"role":      &models.Schema{Type: "string"},
						"firstName": &models.Schema{Type: "string"},
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
