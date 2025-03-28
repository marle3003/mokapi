package asyncApi_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/providers/asyncapi3"
	"mokapi/schema/json/schema"
	"net/url"
	"testing"
)

type readFunc func(cfg *dynamic.Config) error

type testReader struct {
	readFunc readFunc
}

func (tr *testReader) Read(u *url.URL, v any) (*dynamic.Config, error) {
	cfg := &dynamic.Config{Info: dynamic.ConfigInfo{Url: u}}
	if err := tr.readFunc(cfg); err != nil {
		return cfg, err
	}
	if p, ok := cfg.Data.(dynamic.Parser); ok {
		return cfg, p.Parse(cfg, tr)
	}
	return cfg, nil
}

func (tr *testReader) Close() {}

func TestServerResolve(t *testing.T) {
	testcases := []struct {
		name string
		cfg  *asyncApi.Config
		read readFunc
		test func(t *testing.T, cfg *asyncApi.Config, err error)
	}{
		{
			name: "no error when server value is nil",
			read: func(cfg *dynamic.Config) error {
				return nil
			},
			cfg: &asyncApi.Config{Servers: map[string]*asyncApi.ServerRef{"foo": nil}},
			test: func(t *testing.T, cfg *asyncApi.Config, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "server with local reference",
			read: func(cfg *dynamic.Config) error {
				return nil
			},
			cfg: &asyncApi.Config{
				Servers: map[string]*asyncApi.ServerRef{
					"foo": {Ref: "#/components/servers/foo"},
				},
				Components: &asyncApi.Components{Servers: map[string]*asyncApi.Server{"foo": {Description: "foo"}}},
			},
			test: func(t *testing.T, cfg *asyncApi.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cfg.Servers["foo"].Value.Description)
			},
		},
		{
			name: "server with reference",
			read: func(cfg *dynamic.Config) error {
				require.Equal(t, "/foo.yml", cfg.Info.Url.String())
				config := &asyncApi.Config{Servers: map[string]*asyncApi.ServerRef{
					"foo": {Value: &asyncApi.Server{Description: "foo"}},
				}}
				cfg.Data = config
				return nil
			},
			cfg: &asyncApi.Config{Servers: map[string]*asyncApi.ServerRef{
				"foo": {Ref: "foo.yml#/servers/foo"},
			}},
			test: func(t *testing.T, cfg *asyncApi.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cfg.Servers["foo"].Value.Description)
			},
		},
		{
			name: "server with reference components",
			read: func(cfg *dynamic.Config) error {
				require.Equal(t, "/foo.yml", cfg.Info.Url.String())
				config := &asyncApi.Config{Components: &asyncApi.Components{Servers: map[string]*asyncApi.Server{"foo": {Description: "foo"}}}}
				cfg.Data = config
				return nil
			},
			cfg: &asyncApi.Config{Servers: map[string]*asyncApi.ServerRef{
				"foo": {Ref: "foo.yml#/components/servers/foo"},
			}},
			test: func(t *testing.T, cfg *asyncApi.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cfg.Servers["foo"].Value.Description)
			},
		},
		{
			name: "file reference but nil",
			read: func(cfg *dynamic.Config) error {
				require.Equal(t, "/foo.yml", cfg.Info.Url.String())
				config := &asyncApi.Config{Servers: map[string]*asyncApi.ServerRef{
					"foo": {},
				}}
				cfg.Data = config
				return nil
			},
			cfg: &asyncApi.Config{Servers: map[string]*asyncApi.ServerRef{
				"foo": {Ref: "foo.yml#/servers/foo"},
			}},
			test: func(t *testing.T, cfg *asyncApi.Config, err error) {
				require.NoError(t, err)
				require.Nil(t, cfg.Servers["foo"].Value)
			},
		},
		{
			name: "reader returns error",
			read: func(cfg *dynamic.Config) error {
				return fmt.Errorf("TEST ERROR")
			},
			cfg: &asyncApi.Config{Servers: map[string]*asyncApi.ServerRef{
				"foo": {Ref: "foo.yml#/servers/foo"},
			}},
			test: func(t *testing.T, cfg *asyncApi.Config, err error) {
				require.EqualError(t, err, "resolve reference 'foo.yml#/servers/foo' failed: TEST ERROR")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			reader := &testReader{readFunc: tc.read}
			err := tc.cfg.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: tc.cfg}, reader)
			tc.test(t, tc.cfg, err)
		})
	}
}

func TestChannelResolve(t *testing.T) {
	testcases := []struct {
		name string
		cfg  *asyncApi.Config
		read readFunc
		test func(t *testing.T, cfg *asyncApi.Config, err error)
	}{
		{
			name: "empty should not error",
			read: func(cfg *dynamic.Config) error {
				return nil
			},
			cfg: &asyncApi.Config{},
			test: func(t *testing.T, cfg *asyncApi.Config, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "no error when channel value is nil",
			read: func(cfg *dynamic.Config) error {
				return nil
			},
			cfg: &asyncApi.Config{Channels: map[string]*asyncApi.ChannelRef{"foo": nil}},
			test: func(t *testing.T, cfg *asyncApi.Config, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "channel with reference",
			read: func(cfg *dynamic.Config) error {
				require.Equal(t, "/foo.yml", cfg.Info.Url.String())
				target := &asyncApi.Channel{Description: "reference"}
				config := &asyncApi.Config{Channels: map[string]*asyncApi.ChannelRef{
					"foo": {Value: target},
				}}
				cfg.Data = config
				return nil
			},
			cfg: &asyncApi.Config{Channels: map[string]*asyncApi.ChannelRef{
				"foo": {Ref: "foo.yml#/channels/foo"},
			}},
			test: func(t *testing.T, cfg *asyncApi.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "reference", cfg.Channels["foo"].Value.Description)
			},
		},
		{
			name: "file reference but nil",
			read: func(cfg *dynamic.Config) error {
				require.Equal(t, "/foo.yml", cfg.Info.Url.String())
				config := &asyncApi.Config{Channels: map[string]*asyncApi.ChannelRef{
					"foo": {},
				}}
				cfg.Data = config
				return nil
			},
			cfg: &asyncApi.Config{Channels: map[string]*asyncApi.ChannelRef{
				"foo": {Ref: "foo.yml#/channels/foo"},
			}},
			test: func(t *testing.T, cfg *asyncApi.Config, err error) {
				require.NoError(t, err)
				require.Nil(t, cfg.Channels["foo"].Value)
			},
		},
		{
			name: "reader returns error",
			read: func(cfg *dynamic.Config) error {
				return fmt.Errorf("TEST ERROR")
			},
			cfg: &asyncApi.Config{Channels: map[string]*asyncApi.ChannelRef{
				"foo": {Ref: "foo.yml#/channels/foo"},
			}},
			test: func(t *testing.T, cfg *asyncApi.Config, err error) {
				require.EqualError(t, err, "resolve reference 'foo.yml#/channels/foo' failed: TEST ERROR")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			reader := &testReader{readFunc: tc.read}
			err := tc.cfg.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: tc.cfg}, reader)
			tc.test(t, tc.cfg, err)
		})
	}
}

func TestMessage(t *testing.T) {
	testcases := []struct {
		name string
		cfg  *asyncApi.Config
		read readFunc
		test func(t *testing.T, cfg *asyncApi.Config, err error)
	}{
		{
			name: "local subscribe message reference",
			read: func(cfg *dynamic.Config) error {
				return nil
			},
			cfg: &asyncApi.Config{Channels: map[string]*asyncApi.ChannelRef{
				"foo": {Value: &asyncApi.Channel{Subscribe: &asyncApi.Operation{Message: &asyncApi.MessageRef{Ref: "#/components/messages/foo"}}}},
			}, Components: &asyncApi.Components{
				Messages: map[string]*asyncApi.Message{"foo": {Description: "foo"}},
			}},
			test: func(t *testing.T, cfg *asyncApi.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cfg.Channels["foo"].Value.Subscribe.Message.Value.Description)
			},
		},
		{
			name: "local publish message reference",
			read: func(cfg *dynamic.Config) error {
				return nil
			},
			cfg: &asyncApi.Config{Channels: map[string]*asyncApi.ChannelRef{
				"foo": {Value: &asyncApi.Channel{Publish: &asyncApi.Operation{Message: &asyncApi.MessageRef{Ref: "#/components/messages/foo"}}}},
			}, Components: &asyncApi.Components{
				Messages: map[string]*asyncApi.Message{"foo": {Description: "foo"}},
			}},
			test: func(t *testing.T, cfg *asyncApi.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cfg.Channels["foo"].Value.Publish.Message.Value.Description)
			},
		},
		{
			name: "subscribe message file reference",
			read: func(cfg *dynamic.Config) error {
				require.Equal(t, "/foo.yml", cfg.Info.Url.String())
				config := &asyncApi.Config{Components: &asyncApi.Components{
					Messages: map[string]*asyncApi.Message{"foo": {Description: "foo"}},
				}}
				cfg.Data = config
				return nil
			},
			cfg: &asyncApi.Config{Channels: map[string]*asyncApi.ChannelRef{
				"foo": {Value: &asyncApi.Channel{Subscribe: &asyncApi.Operation{Message: &asyncApi.MessageRef{Ref: "foo.yml#/components/messages/foo"}}}},
			}},
			test: func(t *testing.T, cfg *asyncApi.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cfg.Channels["foo"].Value.Subscribe.Message.Value.Description)
			},
		},
		{
			name: "publish message file reference",
			read: func(cfg *dynamic.Config) error {
				require.Equal(t, "/foo.yml", cfg.Info.Url.String())
				config := &asyncApi.Config{Components: &asyncApi.Components{
					Messages: map[string]*asyncApi.Message{"foo": {Description: "foo"}},
				}}
				cfg.Data = config
				return nil
			},
			cfg: &asyncApi.Config{Channels: map[string]*asyncApi.ChannelRef{
				"foo": {Value: &asyncApi.Channel{Publish: &asyncApi.Operation{Message: &asyncApi.MessageRef{Ref: "foo.yml#/components/messages/foo"}}}},
			}},
			test: func(t *testing.T, cfg *asyncApi.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cfg.Channels["foo"].Value.Publish.Message.Value.Description)
			},
		},
		{
			name: "reader returns error",
			read: func(cfg *dynamic.Config) error {
				return fmt.Errorf("TEST ERROR")
			},
			cfg: &asyncApi.Config{Channels: map[string]*asyncApi.ChannelRef{
				"foo": {Value: &asyncApi.Channel{Subscribe: &asyncApi.Operation{Message: &asyncApi.MessageRef{Ref: "foo.yml#/components/messages/foo"}}}},
			}},
			test: func(t *testing.T, cfg *asyncApi.Config, err error) {
				require.EqualError(t, err, "resolve reference 'foo.yml#/components/messages/foo' failed: TEST ERROR")
			},
		},
		{
			name: "use defaultContentType",
			cfg: &asyncApi.Config{
				DefaultContentType: "text/plain",
				Channels: map[string]*asyncApi.ChannelRef{
					"foo": {Value: &asyncApi.Channel{Publish: &asyncApi.Operation{Message: &asyncApi.MessageRef{Value: &asyncApi.Message{}}}}},
				},
			},
			test: func(t *testing.T, cfg *asyncApi.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "text/plain", cfg.Channels["foo"].Value.Publish.Message.Value.ContentType)
			},
		},
		{
			name: "missing content type",
			cfg: &asyncApi.Config{
				Channels: map[string]*asyncApi.ChannelRef{
					"foo": {Value: &asyncApi.Channel{Publish: &asyncApi.Operation{Message: &asyncApi.MessageRef{Value: &asyncApi.Message{}}}}},
				},
			},
			test: func(t *testing.T, cfg *asyncApi.Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "application/json", cfg.Channels["foo"].Value.Publish.Message.Value.ContentType)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			reader := &testReader{readFunc: tc.read}
			err := tc.cfg.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: tc.cfg}, reader)
			tc.test(t, tc.cfg, err)
		})
	}

	t.Run("modify file", func(t *testing.T) {
		target := &asyncApi.Message{ContentType: "application/json"}
		reader := &testReader{readFunc: func(cfg *dynamic.Config) error { return nil }}
		config := &asyncApi.Config{Channels: map[string]*asyncApi.ChannelRef{
			"foo": {Value: &asyncApi.Channel{Publish: &asyncApi.Operation{Message: &asyncApi.MessageRef{Ref: "#/components/messages/foo"}}}},
		}, Components: &asyncApi.Components{
			Messages: map[string]*asyncApi.Message{"foo": {}},
		}}
		file := &dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}
		err := config.Parse(file, reader)

		// modify file
		file.Data.(*asyncApi.Config).Components.Messages["foo"] = target

		require.NoError(t, err)
		require.Equal(t, target, config.Channels["foo"].Value.Publish.Message.Value)
	})
}

func TestModifyFileResolve(t *testing.T) {
	target := &asyncApi.Channel{}
	var fooConfig *dynamic.Config
	reader := &testReader{readFunc: func(cfg *dynamic.Config) error {
		require.Equal(t, "/foo.yml", cfg.Info.Url.String())
		config := &asyncApi.Config{Channels: map[string]*asyncApi.ChannelRef{
			"foo": {Value: &asyncApi.Channel{}},
		}}
		cfg.Data = config
		fooConfig = cfg
		return nil
	}}
	config := &asyncApi.Config{Channels: map[string]*asyncApi.ChannelRef{
		"foo": {Ref: "foo.yml#/channels/foo"},
	}}
	err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
	require.NoError(t, err)

	fooConfig.Data.(*asyncApi.Config).Channels["foo"].Value = target
	err = fooConfig.Data.(dynamic.Parser).Parse(fooConfig, reader)

	require.NoError(t, err)
	require.Equal(t, target, config.Channels["foo"].Value)
}

func TestSchema(t *testing.T) {
	message := &asyncApi.Message{}
	config := &asyncApi.Config{
		Channels: map[string]*asyncApi.ChannelRef{
			"foo": {Value: &asyncApi.Channel{
				Publish: &asyncApi.Operation{
					Message: &asyncApi.MessageRef{
						Value: message,
					},
				},
			},
			},
		}}

	t.Run("nil should not error", func(t *testing.T) {
		reader := &testReader{readFunc: func(file *dynamic.Config) error { return nil }}
		err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
		require.NoError(t, err)
	})
	t.Run("reference inside", func(t *testing.T) {
		target := &schema.Schema{}
		schemas := map[string]*asyncapi3.SchemaRef{}
		schemas["foo"] = &asyncapi3.SchemaRef{Value: &asyncapi3.MultiSchemaFormat{Schema: target}}
		config.Components = &asyncApi.Components{Schemas: schemas}
		message.Payload = &asyncapi3.SchemaRef{Reference: dynamic.Reference{Ref: "#/components/schemas/foo"}}
		reader := &testReader{readFunc: func(cfg *dynamic.Config) error { return nil }}

		err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
		require.NoError(t, err)
		require.Equal(t, target, message.Payload.Value.Schema.(*schema.Schema))
	})
	t.Run("file reference direct", func(t *testing.T) {
		target := &schema.Schema{}
		message.Payload = &asyncapi3.SchemaRef{Reference: dynamic.Reference{Ref: "foo.yml"}}
		reader := &testReader{readFunc: func(cfg *dynamic.Config) error {
			cfg.Data = target
			return nil
		}}

		err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
		require.NoError(t, err)
		require.Equal(t, target, message.Payload.Value.Schema)
	})
	t.Run("modify file reference direct", func(t *testing.T) {
		target := &schema.Schema{}
		message.Payload = &asyncapi3.SchemaRef{Value: &asyncapi3.MultiSchemaFormat{Schema: &schema.Schema{Ref: "foo.yml"}}}
		var fooConfig *dynamic.Config
		reader := &testReader{readFunc: func(file *dynamic.Config) error {
			file.Data = &schema.Schema{}
			fooConfig = file
			return nil
		}}

		err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
		require.NoError(t, err)

		// modify
		fooConfig.Data = target
		err = fooConfig.Data.(dynamic.Parser).Parse(fooConfig, reader)

		require.NoError(t, err)
		require.Equal(t, target, message.Payload.Value.Schema.(*schema.Schema))
	})
}
