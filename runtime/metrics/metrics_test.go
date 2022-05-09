package metrics

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCounter_Add(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			name: "add",
			f: func(t *testing.T) {
				c := NewCounter(WithName("foo"))
				c.Add(1)
				require.Equal(t, 1.0, c.Value())
			},
		},
		{
			name: "sub",
			f: func(t *testing.T) {
				c := NewCounter(WithName("foo"))
				c.Add(-1)
				require.Equal(t, -1.0, c.Value())
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.f(t)
		})
	}
}

func TestNewCounterMap(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			name: "add",
			f: func(t *testing.T) {
				c := NewCounterMap(WithName("foo"), WithLabelNames("foo"))
				c.WithLabel("bar").Add(1)
				require.Equal(t, 1.0, c.Sum())
			},
		},
		{
			name: "other label value",
			f: func(t *testing.T) {
				c := NewCounterMap(WithName("foo"), WithLabelNames("foo"))
				c.WithLabel("bar").Add(1)
				require.Equal(t, 0.0, c.Value(&Query{Labels: []*Label{{Name: "foo", Value: "foo"}}}))
			},
		},
		{
			name: "json",
			f: func(t *testing.T) {
				c := NewCounterMap(WithName("foo"), WithLabelNames("foo"))
				c.WithLabel("bar").Add(1)
				result := make([]Metric, 0)
				ch := make(chan Metric)
				go func() {
					c.Collect(ch)
					close(ch)
				}()
				for m := range ch {
					result = append(result, m)
				}
				data, err := json.Marshal(result)
				require.NoError(t, err)
				require.Equal(t, `[{"name":"foo{foo=\"bar\"}","value":1}]`, string(data))
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.f(t)
		})
	}
}
