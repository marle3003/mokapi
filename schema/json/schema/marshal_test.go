package schema

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRef_MarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		s    *Schema
		test func(t *testing.T, s string, err error)
	}{
		{
			name: "empty type",
			s:    &Schema{},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, "{}", s)
			},
		},
		{
			name: "with type",
			s:    &Schema{Type: Types{"string"}},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"type":"string"}`, s)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b, err := json.Marshal(tc.s)
			tc.test(t, string(b), err)
		})
	}
}
