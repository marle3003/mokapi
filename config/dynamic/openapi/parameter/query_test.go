package parameter

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"net/url"
	"testing"
)

func TestParseQuery(t *testing.T) {
	testcases := []struct {
		u *url.URL
		p *Parameter
		e interface{}
	}{
		{&url.URL{RawQuery: "id=5"},
			&Parameter{
				Name:    "id",
				Schema:  &schema.Ref{Value: &schema.Schema{Type: "integer"}},
				Style:   "",
				Explode: false,
			},
			5,
		},
		{&url.URL{RawQuery: ""},
			&Parameter{
				Name:    "id",
				Schema:  &schema.Ref{Value: &schema.Schema{Type: "integer"}},
				Style:   "",
				Explode: false,
			},
			nil,
		},
		{&url.URL{RawQuery: "id=3&id=4&id=5"},
			&Parameter{
				Name:    "id",
				Schema:  &schema.Ref{Value: &schema.Schema{Type: "array", Items: &schema.Ref{Value: &schema.Schema{Type: "integer"}}}},
				Style:   "",
				Explode: true,
			},
			[]interface{}{3, 4, 5},
		},
		{&url.URL{RawQuery: "id=3,4,5"},
			&Parameter{
				Name:    "id",
				Schema:  &schema.Ref{Value: &schema.Schema{Type: "array", Items: &schema.Ref{Value: &schema.Schema{Type: "integer"}}}},
				Style:   "",
				Explode: false,
			},
			[]interface{}{3, 4, 5},
		},
		{&url.URL{RawQuery: "id=3%204%205"},
			&Parameter{
				Name:    "id",
				Schema:  &schema.Ref{Value: &schema.Schema{Type: "array", Items: &schema.Ref{Value: &schema.Schema{Type: "integer"}}}},
				Style:   "spaceDelimited",
				Explode: false,
			},
			[]interface{}{3, 4, 5},
		},
		{&url.URL{RawQuery: "id=3|4|5"},
			&Parameter{
				Name:    "id",
				Schema:  &schema.Ref{Value: &schema.Schema{Type: "array", Items: &schema.Ref{Value: &schema.Schema{Type: "integer"}}}},
				Style:   "pipeDelimited",
				Explode: false,
			},
			[]interface{}{3, 4, 5},
		},
		{&url.URL{RawQuery: "role=admin&firstName=Alex"},
			&Parameter{
				Name: "id",
				Schema: &schema.Ref{Value: schematest.New("object",
					schematest.WithProperty("role", schematest.New("string")),
					schematest.WithProperty("firstName", schematest.New("string")),
				)},
				Style:   "",
				Explode: true,
			},
			map[string]interface{}{"role": "admin", "firstName": "Alex"},
		},
		{&url.URL{RawQuery: "id=role,admin,firstName,Alex"},
			&Parameter{
				Name: "id",
				Schema: &schema.Ref{Value: schematest.New("object",
					schematest.WithProperty("role", schematest.New("string")),
					schematest.WithProperty("firstName", schematest.New("string")),
				)},
				Style:   "",
				Explode: false,
			},
			map[string]interface{}{"role": "admin", "firstName": "Alex"},
		},
		{&url.URL{RawQuery: "id[role]=admin&id[firstName]=Alex"},
			&Parameter{
				Name: "id",
				Schema: &schema.Ref{Value: schematest.New("object",
					schematest.WithProperty("role", schematest.New("string")),
					schematest.WithProperty("firstName", schematest.New("string")),
				)},
				Style:   "deepObject",
				Explode: true,
			},
			map[string]interface{}{"role": "admin", "firstName": "Alex"},
		},
	}

	for _, testcase := range testcases {
		test := testcase
		t.Run(test.u.String(), func(t *testing.T) {
			i, err := parseQuery(test.p, test.u)
			require.NoError(t, err)
			require.Equal(t, test.e, i.Value)
		})

	}
}
