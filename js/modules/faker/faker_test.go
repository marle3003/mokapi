package faker

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"testing"
)

func TestModule_Fake(t *testing.T) {
	gofakeit.Seed(11)
	m := New(nil, nil).(*Module)
	rt := goja.New()
	s := rt.ToValue(schematest.New("string"))
	v, err := m.Fake(s)
	require.NoError(t, err)
	require.Equal(t, "gbRMaRxHkiJBPta", v)
}

func TestModule_Fake_Invalid_Parameter(t *testing.T) {
	gofakeit.Seed(11)
	m := New(nil, nil).(*Module)
	rt := goja.New()
	s := rt.ToValue("foo")
	_, err := m.Fake(s)
	require.EqualError(t, err, "expected parameter type of schema")
}
