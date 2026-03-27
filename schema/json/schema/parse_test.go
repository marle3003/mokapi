package schema_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParse(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "recursion",
			test: func(t *testing.T) {
				s := schematest.New("object",
					schematest.WithDescription("root"),
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithProperty("children",
						schematest.New("array", schematest.WithItemsRef("#")),
					),
				)

				err := s.Parse(&dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.NoError(t, err)
				children := s.Properties.Get("children")
				require.Equal(t, s.Description, children.Items.Description)
			},
		},
		{
			name: "recursion using $def",
			test: func(t *testing.T) {
				data := `
$defs:
  a:
    $ref: #/$defs/b
  b:
    $ref: '#'
$ref: '#/$defs/a'
`
				var s *schema.Schema
				err := yaml.Unmarshal([]byte(data), &s)
				require.NoError(t, err)

				err = s.Parse(&dynamic.Config{Data: s}, &dynamictest.Reader{})
				require.NoError(t, err)
				require.Equal(t, "empty schema", s.String())
			},
		},
		{
			name: "self-recursion",
			test: func(t *testing.T) {
				data := `
$defs:
  a:
    properties:
      part:
        $ref: '#/$defs/a'
$ref: '#/$defs/a'
`
				var s *schema.Schema
				err := yaml.Unmarshal([]byte(data), &s)
				require.NoError(t, err)

				err = s.Parse(&dynamic.Config{Data: s}, &dynamictest.Reader{})
				require.NoError(t, err)
				require.Equal(t, "schema properties=[part]", s.String())
			},
		},
		{
			name: "parsing twice with $refs",
			test: func(t *testing.T) {
				reader := &dynamictest.Reader{
					Data: map[string]*dynamic.Config{
						"https://example.com/schemas/foo": {
							Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("https://example.com/schemas/foo")),
							Raw: []byte(`{
"$defs": { "items": { "type": "integer" } },
"type": "array",
"items": { "$ref": "#/$defs/items" },
}`),
						},
					},
				}

				person := &dynamic.Config{
					Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("https://example.com/schemas/bar")),
					Data: &schema.Schema{
						Ref: "https://example.com/schemas/foo",
					},
				}

				// 1. parsing
				err := person.Data.(*schema.Schema).Parse(person, reader)
				require.NoError(t, err)

				// 2. parsing
				err = person.Data.(*schema.Schema).Parse(person, reader)

				require.NoError(t, err)
				require.Equal(t, "integer", person.Data.(*schema.Schema).Items.Type.String())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}
