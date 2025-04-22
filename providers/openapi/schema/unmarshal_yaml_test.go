package schema

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestSchema_UnmarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		data string
		test func(t *testing.T, v *Schema, err error)
	}{
		{
			name: "example date-time string",
			data: `example: 2025-04-22T14:30:00Z`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "2025-04-22T14:30:00Z", s.Example.Value)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := &Schema{}
			err := yaml.Unmarshal([]byte(tc.data), &s)
			tc.test(t, s, err)
		})
	}
}
