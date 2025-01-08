package schema

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAdditionalProperties_Unmarshal(t *testing.T) {
	testcases := []struct {
		name string
		s    string
		test func(t *testing.T, r *Schema, err error)
	}{
		{
			name: "additional properties true",
			s:    `{ "type": "object", "additionalProperties": true }`,
			test: func(t *testing.T, r *Schema, err error) {
				require.NoError(t, err)
				require.True(t, r.AdditionalProperties.IsFreeForm())
			},
		},
		{
			name: "additional properties false",
			s:    `{ "type": "object", "additionalProperties": false }`,
			test: func(t *testing.T, r *Schema, err error) {
				require.NoError(t, err)
				require.False(t, r.IsFreeForm())
			},
		},
		{
			name: "additional properties {}",
			s:    `{ "type": "object", "additionalProperties": {} }`,
			test: func(t *testing.T, r *Schema, err error) {
				require.NoError(t, err)
				require.True(t, r.AdditionalProperties.IsAny())
			},
		},
		{
			name: "additional properties values are string",
			s:    `{ "type": "object", "additionalProperties": { "type": "string" } }`,
			test: func(t *testing.T, r *Schema, err error) {
				require.NoError(t, err)
				require.True(t, r.AdditionalProperties.IsString())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := &Schema{}
			err := json.Unmarshal([]byte(tc.s), &r)
			tc.test(t, r, err)
		})
	}
}
