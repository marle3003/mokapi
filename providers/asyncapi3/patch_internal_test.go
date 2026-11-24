package asyncapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPatch_NilPatchKeepsValue(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "server ref",
			test: func(t *testing.T) {
				r := &ServerRef{Value: &Server{Host: "localhost"}}
				r.patch(nil, copies{})
				require.Equal(t, "localhost", r.Value.Host)
			},
		},
		{
			name: "channel ref",
			test: func(t *testing.T) {
				r := &ChannelRef{Value: &Channel{Address: "foo"}}
				r.patch(nil, copies{})
				require.Equal(t, "foo", r.Value.Address)
			},
		},
		{
			name: "operation ref",
			test: func(t *testing.T) {
				r := &OperationRef{Value: &Operation{Title: "foo"}}
				r.patch(nil, copies{})
				require.Equal(t, "foo", r.Value.Title)
			},
		},
		{
			name: "message ref",
			test: func(t *testing.T) {
				r := &MessageRef{Value: &Message{Name: "foo"}}
				r.patch(nil, copies{})
				require.Equal(t, "foo", r.Value.Name)
			},
		},
		{
			name: "server",
			test: func(t *testing.T) {
				s := &Server{Host: "localhost"}
				s.patch(nil)
				require.Equal(t, "localhost", s.Host)
			},
		},
		{
			name: "operation",
			test: func(t *testing.T) {
				o := &Operation{Title: "foo"}
				o.patch(nil)
				require.Equal(t, "foo", o.Title)
			},
		},
		{
			name: "message",
			test: func(t *testing.T) {
				m := &Message{Name: "foo"}
				m.patch(nil, copies{})
				require.Equal(t, "foo", m.Name)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}
