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

func TestSchema_MarshalJSON_Recursion(t *testing.T) {
	testcases := []struct {
		name string
		s    func() *Schema
		test func(t *testing.T, s string, err error)
	}{
		{
			name: "recursion in property",
			s: func() *Schema {
				s := &Schema{
					Ref:        "foo",
					Properties: &Schemas{},
				}
				s.Properties.Set("foo", s)
				return s
			},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"$ref":"foo","properties":{"foo":{"$ref":"foo"}}}`, s)
			},
		},
		{
			name: "recursion in array",
			s: func() *Schema {
				s := &Schema{
					Ref: "foo",
				}
				s.Items = s
				return s
			},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"$ref":"foo","items":{"$ref":"foo"}}`, s)
			},
		},
		{
			name: "recursion in contains",
			s: func() *Schema {
				s := &Schema{
					Ref: "foo",
				}
				s.Contains = s
				return s
			},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"$ref":"foo","contains":{"$ref":"foo"}}`, s)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b, err := json.Marshal(tc.s())
			tc.test(t, string(b), err)
		})
	}
}
