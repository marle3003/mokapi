package swagger_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/openapi"
	"mokapi/providers/swagger"
	"testing"
)

func TestConfig_Parse(t *testing.T) {
	var cfg *swagger.Config
	err := cfg.Parse(
		dynamictest.NewConfig(cfg, dynamictest.WithInfo(dynamictest.NewConfigInfo())),
		&dynamictest.Reader{})
	require.NoError(t, err)

	cfg = &swagger.Config{
		Info: openapi.Info{Name: "foo"},
	}
	c := dynamictest.NewConfig(cfg, dynamictest.WithInfo(dynamictest.NewConfigInfo()))
	err = cfg.Parse(c, &dynamictest.Reader{})
	require.NoError(t, err)
	require.IsType(t, &openapi.Config{}, c.Data())
	require.Equal(t, "foo", cfg.Info.Name)
}
