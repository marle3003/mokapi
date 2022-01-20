package schema

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestRef_UnmarshalYAML(t *testing.T) {
	for _, testcase := range []struct {
		name string
		s    string
		fn   func(t *testing.T, r *Ref)
	}{
		{
			name: "ref",
			s:    "$ref: '#/components/schema/Foo'",
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, "#/components/schema/Foo", r.Ref())
			},
		},
		{
			name: "value",
			s:    "type: string",
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, "string", r.Value.Type)
			},
		},
	} {
		test := testcase
		t.Run(test.name, func(t *testing.T) {
			r := &Ref{}
			err := yaml.Unmarshal([]byte(test.s), r)
			require.NoError(t, err)
			test.fn(t, r)
		})
	}
}
