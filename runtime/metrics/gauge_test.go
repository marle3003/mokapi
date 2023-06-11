package metrics

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGauge_Value(t *testing.T) {
	g := NewGauge()
	g.Set(1.0)
	require.Equal(t, 1.0, g.Value())
}

func TestGauge_Info_Name(t *testing.T) {
	g := NewGauge(WithName("foo"))
	require.Equal(t, "foo", g.Info().Name)
}

func TestGauge_Info_FQName(t *testing.T) {
	g := NewGauge(WithFQName("foo", "bar"))
	require.Equal(t, "foo", g.Info().Namespace)
	require.Equal(t, "bar", g.Info().Name)
	require.Equal(t, "foo_bar", g.Info().FQName())
}

func TestGauge_MarshalJSON(t *testing.T) {
	g := NewGauge(WithName("foo"))
	g.Set(1.0)
	b, err := g.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, `{"name":"foo","value":1}`, string(b))
}

func TestGauge_Collect(t *testing.T) {
	g := NewGauge()
	ch := make(chan Metric, 1)
	g.Collect(ch)
	collected := <-ch
	require.Equal(t, g, collected)
}

func TestGaugeMap_Value(t *testing.T) {
	m := NewGaugeMap(WithName("map"), WithLabelNames("label"))
	m.WithLabel("foo").Set(1.0)
	m.WithLabel("bar").Set(2.0)
	require.Equal(t, 3.0, m.Value(&Query{Name: "map"}))
}

func TestGaugeMap_Set_Value_Panic_When_Labels_Count_Not_Match(t *testing.T) {
	m := NewGaugeMap(WithName("map"))

	defer func() {
		r := recover()
		require.NotNil(t, r, "should panic")
	}()

	m.WithLabel("foo").Set(1.0)
}

func TestGaugeMap_Info_Name(t *testing.T) {
	m := NewGaugeMap(WithName("foo"))
	require.Equal(t, "foo", m.Info().Name)
}

func TestGaugeMap_Info_FQName(t *testing.T) {
	m := NewGaugeMap(WithFQName("foo", "bar"))
	require.Equal(t, "foo", m.Info().Namespace)
	require.Equal(t, "bar", m.Info().Name)
	require.Equal(t, "foo_bar", m.Info().FQName())
}

func TestGaugeMap_FindAll(t *testing.T) {
	m := NewGaugeMap(WithName("map"), WithLabelNames("label"))
	expected := m.WithLabel("foo")
	m.WithLabel("bar")
	found := m.FindAll(
		&Query{Labels: []*Label{
			{"label", "foo"},
		}})
	require.Len(t, found, 1)
	require.Equal(t, expected, found[0])
}

func TestGaugeMap_Collect(t *testing.T) {
	m := NewGaugeMap(WithName("map"), WithLabelNames("label"))
	m.WithLabel("foo")
	m.WithLabel("bar")
	ch := make(chan Metric, 2)
	m.Collect(ch)
	require.Len(t, ch, 2)
}

func TestGaugeMap_MarshalJSON(t *testing.T) {
	m := NewGaugeMap(WithName("map"), WithLabelNames("label"))
	b, err := m.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t,
		`{"name":"map","value":0}`,
		string(b))
}

func TestGaugeMap_MarshalJSON_With_Entries(t *testing.T) {
	m := NewGaugeMap(WithName("map"), WithLabelNames("label"))
	m.WithLabel("foo").Set(1.0)
	m.WithLabel("bar").Set(2.0)
	b, err := m.MarshalJSON()
	require.NoError(t, err)
	// order of output is not fix
	require.True(t,
		`[{"name":"map{label=\"foo\"}","value":1},{"name":"map{label=\"bar\"}","value":2}]` == string(b) ||
			`[{"name":"map{label=\"bar\"}","value":2},{"name":"map{label=\"foo\"}","value":1}]` == string(b))
}
