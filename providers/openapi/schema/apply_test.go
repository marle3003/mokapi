package schema_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/schema/json/schema"
	"testing"
)

func TestSchema_ApplyRef(t *testing.T) {
	testcases := []struct {
		name  string
		input string
		test  func(t *testing.T, s *schema.Schema, err error)
	}{
		{
			name: "boolean used from ref",
			input: `
{
	"$defs": {
  		"foo": false
	},
  	"$ref": "#/$defs/foo"
}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, false, *s.Boolean)
			},
		},
		{
			name: "boolean from ref not used",
			input: `
{
	"$defs": {
  		"foo": false
	},
	"type": "integer",
  	"$ref": "#/$defs/foo"
}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Nil(t, s.Boolean)
				require.Equal(t, "integer", s.Type.String())
			},
		},
		{
			name: "type used from ref",
			input: `
{
	"$defs": {
  		"foo": {
        	"type": "string"
      	}
	},
  	"$ref": "#/$defs/foo"
}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "string", s.Type.String())
			},
		},
		{
			name: "type not overwritten by ref",
			input: `
{
	"$defs": {
  		"foo": {
        	"type": "string"
      	}
	},
	"type": "integer",
  	"$ref": "#/$defs/foo"
}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "integer", s.Type.String())
			},
		},
		{
			name: "enum used from ref",
			input: `
{
	"$defs": {
  		"foo": {
        	"type": "string",
			"enum": ["foo", "bar"]
      	}
	},
	"type": "integer",
  	"$ref": "#/$defs/foo"
}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "integer", s.Type.String())
				require.Equal(t, []interface{}{"foo", "bar"}, s.Enum)
			},
		},
		{
			name: "enum not overwritten by ref",
			input: `
{
	"$defs": {
  		"foo": {
        	"type": "string",
			"enum": ["foo", "bar"]
      	}
	},
	"type": "integer",
	"enum": [1,2],
  	"$ref": "#/$defs/foo"
}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "integer", s.Type.String())
				require.Equal(t, []interface{}{float64(1), float64(2)}, s.Enum)
			},
		},
		{
			name: "const used from ref",
			input: `
{
	"$defs": {
  		"foo": {
        	"type": "string",
			"const": "foo"
      	}
	},
	"type": "integer",
  	"$ref": "#/$defs/foo"
}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "integer", s.Type.String())
				require.Equal(t, "foo", *s.Const)
			},
		},
		{
			name: "const not overwritten by ref",
			input: `
{
	"$defs": {
  		"foo": {
        	"type": "string",
			"const": "foo"
      	}
	},
	"type": "integer",
	"const": 2,
  	"$ref": "#/$defs/foo"
}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "integer", s.Type.String())
				require.Equal(t, float64(2), *s.Const)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var s *schema.Schema
			err := json.Unmarshal([]byte(tc.input), &s)
			if err != nil {
				tc.test(t, s, err)
			} else {
				err = s.Parse(&dynamic.Config{Data: s}, &dynamictest.Reader{})
				tc.test(t, s, err)
			}
		})
	}
}
