package sortedmap_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/sortedmap"
	"testing"
)

func TestLinkedHashMap_Empty(t *testing.T) {
	empty := &sortedmap.LinkedHashMap[string, int]{}
	require.Equal(t, 0, empty.Len())
	v, ok := empty.Get("foo")
	require.Equal(t, false, ok)
	require.Equal(t, 0, v)

	require.Equal(t, []string{}, empty.Keys())
	require.Equal(t, []int{}, empty.Values())

	require.Equal(t, "{}", empty.String())

	it := empty.Iter()
	require.False(t, it.Next())
}

func TestLinkedHashMap_WithValues(t *testing.T) {
	m := &sortedmap.LinkedHashMap[string, int]{}
	m.Set("foo", 1)
	m.Set("bar", 2)

	require.Equal(t, 2, m.Len())
	v, ok := m.Get("foo")
	require.Equal(t, true, ok)
	require.Equal(t, 1, v)

	require.Equal(t, []string{"foo", "bar"}, m.Keys())
	require.Equal(t, []int{1, 2}, m.Values())

	require.Equal(t, "{foo: 1, bar: 2}", m.String())

	it := m.Iter()
	require.True(t, it.Next())
	k, v := it.Item()
	require.Equal(t, "foo", k)
	require.Equal(t, 1, v)
	require.True(t, it.Next())
	k, v = it.Item()
	require.Equal(t, "bar", k)
	require.Equal(t, 2, v)

	require.Equal(t, map[string]int{"foo": 1, "bar": 2}, m.ToMap())

	r, err := m.Resolve("bar")
	require.NoError(t, err)
	require.Equal(t, 2, r)

	_, err = m.Resolve("yuh")
	require.EqualError(t, err, "unable to resolve yuh")

	b, err := json.Marshal(m)
	require.NoError(t, err)
	require.Equal(t, `{"foo":1,"bar":2}`, string(b))

	m.Del("foo")
	require.Equal(t, map[string]int{"bar": 2}, m.ToMap())

	require.Equal(t, 2, m.Lookup("bar"))
}

func TestLinkedHashMap_Merge(t *testing.T) {
	m1 := &sortedmap.LinkedHashMap[string, int]{}
	m1.Set("foo", 1)
	m1.Set("yuh", 3)
	m2 := &sortedmap.LinkedHashMap[string, int]{}
	m2.Set("foo", 10)
	m2.Set("bar", 2)

	m1.Merge(m2)
	require.Equal(t, map[string]int{"foo": 10, "yuh": 3, "bar": 2}, m1.ToMap())
}
