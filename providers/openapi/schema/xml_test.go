package schema_test

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/providers/openapi/schema"
	"testing"
)

func TestXml_UnmarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		s    string
		test func(t *testing.T, r *schema.Xml)
	}{
		{
			name: "xml",
			s: `
  wrapped: true
  name: foo
  attribute: true
  prefix: bar
  namespace: ns1
`,
			test: func(t *testing.T, x *schema.Xml) {
				require.Equal(t, &schema.Xml{
					Wrapped:   true,
					Name:      "foo",
					Attribute: true,
					Prefix:    "bar",
					Namespace: "ns1",
				}, x)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			x := &schema.Xml{}
			err := yaml.Unmarshal([]byte(tc.s), &x)
			require.NoError(t, err)
			tc.test(t, x)
		})
	}
}
