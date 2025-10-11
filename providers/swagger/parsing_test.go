package swagger_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/openapi"
	"mokapi/providers/swagger"
	"testing"
)

func TestConfig_Parse(t *testing.T) {
	var cfg *swagger.Config
	err := cfg.Parse(
		&dynamic.Config{
			Info: dynamictest.NewConfigInfo(),
			Data: cfg,
		},
		&dynamictest.Reader{})
	require.NoError(t, err)

	cfg = &swagger.Config{
		Info: openapi.Info{Name: "foo"},
	}
	c := &dynamic.Config{
		Info: dynamictest.NewConfigInfo(),
		Data: cfg,
	}
	err = cfg.Parse(c, &dynamictest.Reader{})
	require.NoError(t, err)
	require.IsType(t, &openapi.Config{}, c.Data)
	require.Equal(t, "foo", cfg.Info.Name)
}
