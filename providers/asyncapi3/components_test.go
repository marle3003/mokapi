package asyncapi3_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/asyncapi3"
	"mokapi/providers/asyncapi3/asyncapi3test"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"mokapi/try"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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
servers:
  foo:
    $ref: '#/components/servers/foo'
components:
  servers:
    foo:
      $ref: 'test.yaml#/components/servers/foo'
`,
			test: func(cfg *asyncapi3.Config) {
				c := &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: try.MustUrl("/foo")}}

				err := cfg.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, "resolve reference '#/components/servers/foo' failed: resolve reference 'test.yaml#/components/servers/foo' failed: TestReader: config not found")

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
				require.Equal(t, "FOO", cfg.Channels["foo"].Value.Title)
			},
		},
		{
			name: "schemas",
			config: `
channels:
  foo:
    messages:
      msg:
        payload:
          $ref: '#/components/schemas/foo'
components:
  schemas:
    foo:
      $ref: 'test.yaml#/components/schemas/foo'
`,
			test: func(cfg *asyncapi3.Config) {
				c := &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: try.MustUrl("/foo")}}

				err := cfg.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, "resolve reference '#/components/schemas/foo' failed: resolve reference 'test.yaml#/components/schemas/foo' failed: TestReader: config not found")

				err = cfg.Parse(c, dynamictest.ReaderFunc(func(u *url.URL, v any) (*dynamic.Config, error) {
					require.Equal(t, "/test.yaml", u.String())
					return &dynamic.Config{Data: asyncapi3test.NewConfig(asyncapi3test.WithComponentSchema("foo", schematest.New("string")))}, nil
				}))
				require.NoError(t, err)
				require.Equal(t, "string", cfg.Components.Schemas["foo"].Value.(*schema.Schema).Type[0])
			},
		},
		{
			name: "messages",
			config: `
channels:
  foo:
    messages:
      msg:
        $ref: '#/components/messages/foo'
components:
  messages:
    foo:
      $ref: 'test.yaml#/components/messages/foo'
`,
			test: func(cfg *asyncapi3.Config) {
				c := &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: try.MustUrl("/foo")}}

				err := cfg.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, "resolve reference '#/components/messages/foo' failed: resolve reference 'test.yaml#/components/messages/foo' failed: TestReader: config not found")

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
operations:
  foo:
    $ref: '#/components/operations/foo'
components:
  operations:
    foo:
      $ref: 'test.yaml#/components/operations/foo'
`,
			test: func(cfg *asyncapi3.Config) {
				c := &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: try.MustUrl("/foo")}}

				err := cfg.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, "resolve reference '#/components/operations/foo' failed: resolve reference 'test.yaml#/components/operations/foo' failed: TestReader: config not found")

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
channels:
  foo:
    parameters:
      foo:
        $ref: '#/components/parameters/foo'
components:
  parameters:
    foo:
      $ref: 'test.yaml#/components/parameters/foo'
`,
			test: func(cfg *asyncapi3.Config) {
				c := &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: try.MustUrl("/foo")}}

				err := cfg.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, "resolve reference '#/components/parameters/foo' failed: resolve reference 'test.yaml#/components/parameters/foo' failed: TestReader: config not found")

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
channels:
  foo:
    messages:
      msg:
        correlationId:
          $ref: '#/components/correlationIds/foo'
components:
  correlationIds:
    foo:
      $ref: 'test.yaml#/components/correlationIds/foo'
`,
			test: func(cfg *asyncapi3.Config) {
				c := &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: try.MustUrl("/foo")}}

				err := cfg.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, "resolve reference '#/components/correlationIds/foo' failed: resolve reference 'test.yaml#/components/correlationIds/foo' failed: TestReader: config not found")

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
channels:
  foo:
    messages:
      msg:
        externalDocs:
          - $ref: '#/components/externalDocs/foo'
components:
  externalDocs:
    foo:
      $ref: 'test.yaml#/components/externalDocs/foo'
`,
			test: func(cfg *asyncapi3.Config) {
				c := &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: try.MustUrl("/foo")}}

				err := cfg.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, "resolve reference '#/components/externalDocs/foo' failed: resolve reference 'test.yaml#/components/externalDocs/foo' failed: TestReader: config not found")

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
operations:
  foo:
    traits:
      - $ref: '#/components/operationTraits/foo'
components:
  operationTraits:
    foo:
      $ref: 'test.yaml#/components/operationTraits/foo'
`,
			test: func(cfg *asyncapi3.Config) {
				c := &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: try.MustUrl("/foo")}}

				err := cfg.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, "resolve reference '#/components/operationTraits/foo' failed: resolve reference 'test.yaml#/components/operationTraits/foo' failed: TestReader: config not found")

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
channels:
  foo:
    messages:
      msg:
        traits:
          - $ref: '#/components/messageTraits/foo'
components:
  messageTraits:
    foo:
      $ref: 'test.yaml#/components/messageTraits/foo'
`,
			test: func(cfg *asyncapi3.Config) {
				c := &dynamic.Config{Data: cfg, Info: dynamic.ConfigInfo{Url: try.MustUrl("/foo")}}

				err := cfg.Parse(c, &dynamictest.Reader{})
				require.EqualError(t, err, "resolve reference '#/components/messageTraits/foo' failed: resolve reference 'test.yaml#/components/messageTraits/foo' failed: TestReader: config not found")

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
