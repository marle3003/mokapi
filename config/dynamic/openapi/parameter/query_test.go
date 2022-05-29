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
		name string
		u    *url.URL
		p    *Parameter
		e    interface{}
	}{
		{
			"integer",
			&url.URL{RawQuery: "id=5"},
			&Parameter{
				Name:    "id",
				Schema:  &schema.Ref{Value: &schema.Schema{Type: "integer"}},
				Style:   "",
				Explode: false,
			},
			int64(5),
		},
		{
			"empty integer",
			&url.URL{RawQuery: ""},
			&Parameter{
				Name:    "id",
				Schema:  &schema.Ref{Value: &schema.Schema{Type: "integer"}},
				Style:   "",
				Explode: false,
			},
			nil,
		},
		{
			"integer array as explode",
			&url.URL{RawQuery: "id=3&id=4&id=5"},
			&Parameter{
				Name:    "id",
				Schema:  &schema.Ref{Value: &schema.Schema{Type: "array", Items: &schema.Ref{Value: &schema.Schema{Type: "integer"}}}},
				Style:   "",
				Explode: true,
			},
			[]interface{}{int64(3), int64(4), int64(5)},
		},
		{
			"integer array",
			&url.URL{RawQuery: "id=3,4,5"},
			&Parameter{
				Name:    "id",
				Schema:  &schema.Ref{Value: &schema.Schema{Type: "array", Items: &schema.Ref{Value: &schema.Schema{Type: "integer"}}}},
				Style:   "",
				Explode: false,
			},
			[]interface{}{int64(3), int64(4), int64(5)},
		},
		{
			"integer array space delimited",
			&url.URL{RawQuery: "id=3%204%205"},
			&Parameter{
				Name:    "id",
				Schema:  &schema.Ref{Value: &schema.Schema{Type: "array", Items: &schema.Ref{Value: &schema.Schema{Type: "integer"}}}},
				Style:   "spaceDelimited",
				Explode: false,
			},
			[]interface{}{int64(3), int64(4), int64(5)},
		},
		{
			"integer array pipe delimited",
			&url.URL{RawQuery: "id=3|4|5"},
			&Parameter{
				Name:    "id",
				Schema:  &schema.Ref{Value: &schema.Schema{Type: "array", Items: &schema.Ref{Value: &schema.Schema{Type: "integer"}}}},
				Style:   "pipeDelimited",
				Explode: false,
			},
			[]interface{}{int64(3), int64(4), int64(5)},
		},
		{
			"object explode",
			&url.URL{RawQuery: "role=admin&firstName=Alex"},
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
		{
			"free form object explode",
			&url.URL{RawQuery: "role=admin&firstName=Alex"},
			&Parameter{
				Name:    "id",
				Schema:  &schema.Ref{Value: schematest.New("object")},
				Style:   "",
				Explode: true,
			},
			map[string]interface{}{"role": []string{"admin"}, "firstName": []string{"Alex"}},
		},
		{
			"dictionary explode",
			&url.URL{RawQuery: "role=admin&firstName=Alex"},
			&Parameter{
				Name:    "id",
				Schema:  &schema.Ref{Value: schematest.New("object", schematest.WithAdditionalProperties(schematest.New("string")))},
				Style:   "",
				Explode: true,
			},
			map[string]interface{}{"role": "admin", "firstName": "Alex"},
		},
		{
			"object",
			&url.URL{RawQuery: "id=role,admin,firstName,Alex"},
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
		{
			"object deep",
			&url.URL{RawQuery: "id[role]=admin&id[firstName]=Alex"},
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

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			i, err := parseQuery(tc.p, tc.u)
			require.NoError(t, err)
			require.Equal(t, tc.e, i.Value)
		})

	}
}
