package schema_test

import (
	"encoding/json"
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
					Wrapped:   new(true),
					Name:      "foo",
					Attribute: new(true),
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

func TestXml_Marshal(t *testing.T) {
	testcases := []struct {
		name     string
		xml      *schema.Xml
		expected string
	}{
		{name: "flags not set are not written", xml: &schema.Xml{Name: "foo"}, expected: `{"name":"foo","prefix":"","namespace":""}`},
		{name: "explicit false is kept", xml: &schema.Xml{Name: "foo", Wrapped: new(false), Attribute: new(false)}, expected: `{"wrapped":false,"name":"foo","attribute":false,"prefix":"","namespace":""}`},
		{name: "true", xml: &schema.Xml{Name: "foo", Wrapped: new(true), Attribute: new(true)}, expected: `{"wrapped":true,"name":"foo","attribute":true,"prefix":"","namespace":""}`},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.xml)
			require.NoError(t, err)
			require.JSONEq(t, tc.expected, string(b))
		})
	}
}
