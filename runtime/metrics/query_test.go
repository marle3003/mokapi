package metrics

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQueryByName(t *testing.T) {
	q := NewQuery(ByName("foo"))
	require.Equal(t, "foo", q.Name)
}

func TestQueryByFQName(t *testing.T) {
	q := NewQuery(ByFQName("foo"))
	require.Equal(t, "foo", q.FQName)
}

func TestQueryByNamespace(t *testing.T) {
	q := NewQuery(ByNamespace("foo"))
	require.Equal(t, "foo", q.Namespace)
}

func TestQueryByLabel(t *testing.T) {
	q := NewQuery(ByLabel("foo", "bar"))
	require.Len(t, q.Labels, 1)
	require.Equal(t, "foo", q.Labels[0].Name)
	require.Equal(t, "bar", q.Labels[0].Value)
}
