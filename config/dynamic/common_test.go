package dynamic_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/try"
	"mokapi/version"
	"testing"
)

func TestAnyVersion(t *testing.T) {
	require.True(t, dynamic.AnyVersion(version.Version{}))
}

func TestRegister(t *testing.T) {
	type foo struct {
		Foo string `json:"foo"`
	}

	c := &dynamic.Config{
		Info: dynamic.ConfigInfo{
			Url: try.MustUrl("test.json"),
		},
		Raw: []byte(`{"foo": "1.0"}`),
	}
	err := dynamic.Parse(c, &dynamictest.Reader{})
	require.NoError(t, err)
	require.Nil(t, c.Data)

	dynamic.Register("foo", func(v version.Version) bool {
		return v.String() == "1.0.0"
	}, &foo{})

	err = dynamic.Parse(c, &dynamictest.Reader{})
	require.NoError(t, err)
	require.Equal(t, &foo{Foo: "1.0"}, c.Data)
}
