package asyncapi3_test

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/asyncapi3"
	"testing"
)

func TestServerTags(t *testing.T) {
	testcases := []struct {
		name   string
		config string
		test   func(t *testing.T, cfg *asyncapi3.Config, err error)
	}{
		{
			name: "server with tag",
			config: `
servers:
  foo:
    tags:
      - name: foo
        description: bar
`,
			test: func(t *testing.T, cfg *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cfg.Servers["foo"].Value.Tags[0].Value.Name)
				require.Equal(t, "bar", cfg.Servers["foo"].Value.Tags[0].Value.Description)
			},
		},
		{
			name: "reference",
			config: `
servers:
  foo:
    tags:
      - $ref: '#/components/tags/foo'
components:
  tags:
    foo:
      name: foo
      description: bar
`,
			test: func(t *testing.T, cfg *asyncapi3.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cfg.Servers["foo"].Value.Tags[0].Value.Name)
				require.Equal(t, "bar", cfg.Servers["foo"].Value.Tags[0].Value.Description)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var cfg *asyncapi3.Config
			err := yaml.Unmarshal([]byte(tc.config), &cfg)
			if err != nil {
				tc.test(t, cfg, err)
				return
			}

			err = cfg.Parse(&dynamic.Config{Data: cfg}, &dynamictest.Reader{})

			tc.test(t, cfg, err)
		})
	}
}
