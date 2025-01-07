package asyncapi3_test

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schematest"
	"mokapi/try"
	"net/url"
	"testing"
)

func TestComponents_Ref(t *testing.T) {
	testcases := []struct {
		name   string
		config string
		test   func(cfg *asyncapi3.Config)
	}{
		{
			name: "servers",
			config: `
components:
  servers:
    foo:
      $ref: 'test.yaml#/components/servers/foo'
`,
			test: func(cfg *asyncapi3.Config) {
				c := &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: try.MustUrl("/foo")}}

				err := cfg.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, `resolve reference 'test.yaml#/components/servers/foo' failed: TestReader: config not found`)

				err = cfg.Parse(c, dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
					require.Equal(t, "/test.yaml", u.String())
					return &dynamic.Config{Data: asyncapi3test.NewConfig(asyncapi3test.WithComponentServer("foo", &asyncapi3.Server{Title: "FOO"}))}, nil
				}))
				require.NoError(t, err)
				require.Equal(t, "FOO", cfg.Components.Servers["foo"].Value.Title)
			},
		},
		{
			name: "tags",
			config: `
components:
  tags:
    foo:
      $ref: 'test.yaml#/components/tags/foo'
`,
			test: func(cfg *asyncapi3.Config) {
				c := &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: try.MustUrl("/foo")}}

				err := cfg.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, `resolve reference 'test.yaml#/components/tags/foo' failed: TestReader: config not found`)

				err = cfg.Parse(c, dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
					require.Equal(t, "/test.yaml", u.String())
					return &dynamic.Config{Data: asyncapi3test.NewConfig(asyncapi3test.WithComponentTag("foo", &asyncapi3.Tag{Name: "FOO"}))}, nil
				}))
				require.NoError(t, err)
				require.Equal(t, "FOO", cfg.Components.Tags["foo"].Value.Name)
			},
		},
		{
			name: "channels",
			config: `
components:
  channels:
    foo:
      $ref: 'test.yaml#/components/channels/foo'
`,
			test: func(cfg *asyncapi3.Config) {
				c := &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: try.MustUrl("/foo")}}

				err := cfg.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, `resolve reference 'test.yaml#/components/channels/foo' failed: TestReader: config not found`)

				err = cfg.Parse(c, dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
					require.Equal(t, "/test.yaml", u.String())
					return &dynamic.Config{Data: asyncapi3test.NewConfig(asyncapi3test.WithComponentChannel("foo", &asyncapi3.Channel{Title: "FOO"}))}, nil
				}))
				require.NoError(t, err)
				require.Equal(t, "FOO", cfg.Components.Channels["foo"].Value.Title)
			},
		},
		{
			name: "schemas",
			config: `
components:
  schemas:
    foo:
      $ref: 'test.yaml#/components/schemas/foo'
`,
			test: func(cfg *asyncapi3.Config) {
				c := &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: try.MustUrl("/foo")}}

				err := cfg.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, `resolve reference 'test.yaml#/components/schemas/foo' failed: TestReader: config not found`)

				err = cfg.Parse(c, dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
					require.Equal(t, "/test.yaml", u.String())
					return &dynamic.Config{Data: asyncapi3test.NewConfig(asyncapi3test.WithComponentSchema("foo", schematest.New("string")))}, nil
				}))
				require.NoError(t, err)
				require.Equal(t, "string", cfg.Components.Schemas["foo"].Value.Schema.(*schema.Ref).Value.Type[0])
			},
		},
		{
			name: "messages",
			config: `
components:
  messages:
    foo:
      $ref: 'test.yaml#/components/messages/foo'
`,
			test: func(cfg *asyncapi3.Config) {
				c := &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: try.MustUrl("/foo")}}

				err := cfg.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, `resolve reference 'test.yaml#/components/messages/foo' failed: TestReader: config not found`)

				err = cfg.Parse(c, dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
					require.Equal(t, "/test.yaml", u.String())
					return &dynamic.Config{Data: asyncapi3test.NewConfig(asyncapi3test.WithComponentMessage("foo", &asyncapi3.Message{Title: "FOO"}))}, nil
				}))
				require.NoError(t, err)
				require.Equal(t, "FOO", cfg.Components.Messages["foo"].Value.Title)
			},
		},
		{
			name: "operations",
			config: `
components:
  operations:
    foo:
      $ref: 'test.yaml#/components/operations/foo'
`,
			test: func(cfg *asyncapi3.Config) {
				c := &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: try.MustUrl("/foo")}}

				err := cfg.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, `resolve reference 'test.yaml#/components/operations/foo' failed: TestReader: config not found`)

				err = cfg.Parse(c, dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
					require.Equal(t, "/test.yaml", u.String())
					return &dynamic.Config{Data: asyncapi3test.NewConfig(asyncapi3test.WithComponentOperation("foo", &asyncapi3.Operation{Title: "FOO"}))}, nil
				}))
				require.NoError(t, err)
				require.Equal(t, "FOO", cfg.Components.Operations["foo"].Value.Title)
			},
		},
		{
			name: "parameters",
			config: `
components:
  parameters:
    foo:
      $ref: 'test.yaml#/components/parameters/foo'
`,
			test: func(cfg *asyncapi3.Config) {
				c := &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: try.MustUrl("/foo")}}

				err := cfg.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, `resolve reference 'test.yaml#/components/parameters/foo' failed: TestReader: config not found`)

				err = cfg.Parse(c, dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
					require.Equal(t, "/test.yaml", u.String())
					return &dynamic.Config{Data: asyncapi3test.NewConfig(asyncapi3test.WithComponentParameter("foo", &asyncapi3.Parameter{Description: "FOO"}))}, nil
				}))
				require.NoError(t, err)
				require.Equal(t, "FOO", cfg.Components.Parameters["foo"].Value.Description)
			},
		},
		{
			name: "correlationIds",
			config: `
components:
  correlationIds:
    foo:
      $ref: 'test.yaml#/components/correlationIds/foo'
`,
			test: func(cfg *asyncapi3.Config) {
				c := &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: try.MustUrl("/foo")}}

				err := cfg.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, `resolve reference 'test.yaml#/components/correlationIds/foo' failed: TestReader: config not found`)

				err = cfg.Parse(c, dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
					require.Equal(t, "/test.yaml", u.String())
					return &dynamic.Config{Data: asyncapi3test.NewConfig(asyncapi3test.WithComponentCorrelationId("foo", &asyncapi3.CorrelationId{Description: "FOO"}))}, nil
				}))
				require.NoError(t, err)
				require.Equal(t, "FOO", cfg.Components.CorrelationIds["foo"].Value.Description)
			},
		},
		{
			name: "externalDocs",
			config: `
components:
  externalDocs:
    foo:
      $ref: 'test.yaml#/components/externalDocs/foo'
`,
			test: func(cfg *asyncapi3.Config) {
				c := &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: try.MustUrl("/foo")}}

				err := cfg.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, `resolve reference 'test.yaml#/components/externalDocs/foo' failed: TestReader: config not found`)

				err = cfg.Parse(c, dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
					require.Equal(t, "/test.yaml", u.String())
					return &dynamic.Config{Data: asyncapi3test.NewConfig(asyncapi3test.WithComponentExternalDoc("foo", &asyncapi3.ExternalDoc{Description: "FOO"}))}, nil
				}))
				require.NoError(t, err)
				require.Equal(t, "FOO", cfg.Components.ExternalDocs["foo"].Value.Description)
			},
		},
		{
			name: "operationTraits",
			config: `
components:
  operationTraits:
    foo:
      $ref: 'test.yaml#/components/operationTraits/foo'
`,
			test: func(cfg *asyncapi3.Config) {
				c := &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: try.MustUrl("/foo")}}

				err := cfg.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, `resolve reference 'test.yaml#/components/operationTraits/foo' failed: TestReader: config not found`)

				err = cfg.Parse(c, dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
					require.Equal(t, "/test.yaml", u.String())
					return &dynamic.Config{Data: asyncapi3test.NewConfig(asyncapi3test.WithComponentOperationTrait("foo", &asyncapi3.OperationTrait{Description: "FOO"}))}, nil
				}))
				require.NoError(t, err)
				require.Equal(t, "FOO", cfg.Components.OperationTraits["foo"].Value.Description)
			},
		},
		{
			name: "messageTraits",
			config: `
components:
  messageTraits:
    foo:
      $ref: 'test.yaml#/components/messageTraits/foo'
`,
			test: func(cfg *asyncapi3.Config) {
				c := &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: try.MustUrl("/foo")}}

				err := cfg.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, `resolve reference 'test.yaml#/components/messageTraits/foo' failed: TestReader: config not found`)

				err = cfg.Parse(c, dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
					require.Equal(t, "/test.yaml", u.String())
					return &dynamic.Config{Data: asyncapi3test.NewConfig(asyncapi3test.WithComponentMessageTrait("foo", &asyncapi3.MessageTrait{Description: "FOO"}))}, nil
				}))
				require.NoError(t, err)
				require.Equal(t, "FOO", cfg.Components.MessageTraits["foo"].Value.Description)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &asyncapi3.Config{}
			err := yaml.Unmarshal([]byte(tc.config), &cfg)
			require.NoError(t, err)

			tc.test(cfg)
		})
	}
}
