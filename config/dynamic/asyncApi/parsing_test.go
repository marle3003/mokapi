package asyncApi

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/providers/openapi/ref"
	"mokapi/providers/openapi/schema"
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
		cfg  *Config
		read readFunc
		test func(t *testing.T, cfg *Config, err error)
	}{
		{
			name: "no error when server value is nil",
			read: func(cfg *dynamic.Config) error {
				return nil
			},
			cfg: &Config{Servers: map[string]*ServerRef{"foo": nil}},
			test: func(t *testing.T, cfg *Config, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "server with local reference",
			read: func(cfg *dynamic.Config) error {
				return nil
			},
			cfg: &Config{
				Servers: map[string]*ServerRef{
					"foo": {Ref: "#/components/servers/foo"},
				},
				Components: &Components{Servers: map[string]*Server{"foo": {Description: "foo"}}},
			},
			test: func(t *testing.T, cfg *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cfg.Servers["foo"].Value.Description)
			},
		},
		{
			name: "server with reference",
			read: func(cfg *dynamic.Config) error {
				require.Equal(t, "/foo.yml", cfg.Info.Url.String())
				config := &Config{Servers: map[string]*ServerRef{
					"foo": {Value: &Server{Description: "foo"}},
				}}
				cfg.Data = config
				return nil
			},
			cfg: &Config{Servers: map[string]*ServerRef{
				"foo": {Ref: "foo.yml#/servers/foo"},
			}},
			test: func(t *testing.T, cfg *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cfg.Servers["foo"].Value.Description)
			},
		},
		{
			name: "server with reference components",
			read: func(cfg *dynamic.Config) error {
				require.Equal(t, "/foo.yml", cfg.Info.Url.String())
				config := &Config{Components: &Components{Servers: map[string]*Server{"foo": {Description: "foo"}}}}
				cfg.Data = config
				return nil
			},
			cfg: &Config{Servers: map[string]*ServerRef{
				"foo": {Ref: "foo.yml#/components/servers/foo"},
			}},
			test: func(t *testing.T, cfg *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cfg.Servers["foo"].Value.Description)
			},
		},
		{
			name: "file reference but nil",
			read: func(cfg *dynamic.Config) error {
				require.Equal(t, "/foo.yml", cfg.Info.Url.String())
				config := &Config{Servers: map[string]*ServerRef{
					"foo": {},
				}}
				cfg.Data = config
				return nil
			},
			cfg: &Config{Servers: map[string]*ServerRef{
				"foo": {Ref: "foo.yml#/servers/foo"},
			}},
			test: func(t *testing.T, cfg *Config, err error) {
				require.NoError(t, err)
				require.Nil(t, cfg.Servers["foo"].Value)
			},
		},
		{
			name: "reader returns error",
			read: func(cfg *dynamic.Config) error {
				return fmt.Errorf("TEST ERROR")
			},
			cfg: &Config{Servers: map[string]*ServerRef{
				"foo": {Ref: "foo.yml#/servers/foo"},
			}},
			test: func(t *testing.T, cfg *Config, err error) {
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
		cfg  *Config
		read readFunc
		test func(t *testing.T, cfg *Config, err error)
	}{
		{
			name: "empty should not error",
			read: func(cfg *dynamic.Config) error {
				return nil
			},
			cfg: &Config{},
			test: func(t *testing.T, cfg *Config, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "no error when channel value is nil",
			read: func(cfg *dynamic.Config) error {
				return nil
			},
			cfg: &Config{Channels: map[string]*ChannelRef{"foo": nil}},
			test: func(t *testing.T, cfg *Config, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "channel with reference",
			read: func(cfg *dynamic.Config) error {
				require.Equal(t, "/foo.yml", cfg.Info.Url.String())
				target := &Channel{Description: "reference"}
				config := &Config{Channels: map[string]*ChannelRef{
					"foo": {Value: target},
				}}
				cfg.Data = config
				return nil
			},
			cfg: &Config{Channels: map[string]*ChannelRef{
				"foo": {Ref: "foo.yml#/channels/foo"},
			}},
			test: func(t *testing.T, cfg *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "reference", cfg.Channels["foo"].Value.Description)
			},
		},
		{
			name: "file reference but nil",
			read: func(cfg *dynamic.Config) error {
				require.Equal(t, "/foo.yml", cfg.Info.Url.String())
				config := &Config{Channels: map[string]*ChannelRef{
					"foo": {},
				}}
				cfg.Data = config
				return nil
			},
			cfg: &Config{Channels: map[string]*ChannelRef{
				"foo": {Ref: "foo.yml#/channels/foo"},
			}},
			test: func(t *testing.T, cfg *Config, err error) {
				require.NoError(t, err)
				require.Nil(t, cfg.Channels["foo"].Value)
			},
		},
		{
			name: "reader returns error",
			read: func(cfg *dynamic.Config) error {
				return fmt.Errorf("TEST ERROR")
			},
			cfg: &Config{Channels: map[string]*ChannelRef{
				"foo": {Ref: "foo.yml#/channels/foo"},
			}},
			test: func(t *testing.T, cfg *Config, err error) {
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

func TestMessageResolve(t *testing.T) {
	testcases := []struct {
		name string
		cfg  *Config
		read readFunc
		test func(t *testing.T, cfg *Config, err error)
	}{
		{
			name: "local subscribe message reference",
			read: func(cfg *dynamic.Config) error {
				return nil
			},
			cfg: &Config{Channels: map[string]*ChannelRef{
				"foo": {Value: &Channel{Subscribe: &Operation{Message: &MessageRef{Ref: "#/components/messages/foo"}}}},
			}, Components: &Components{
				Messages: map[string]*Message{"foo": {Description: "foo"}},
			}},
			test: func(t *testing.T, cfg *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cfg.Channels["foo"].Value.Subscribe.Message.Value.Description)
			},
		},
		{
			name: "local publish message reference",
			read: func(cfg *dynamic.Config) error {
				return nil
			},
			cfg: &Config{Channels: map[string]*ChannelRef{
				"foo": {Value: &Channel{Publish: &Operation{Message: &MessageRef{Ref: "#/components/messages/foo"}}}},
			}, Components: &Components{
				Messages: map[string]*Message{"foo": {Description: "foo"}},
			}},
			test: func(t *testing.T, cfg *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cfg.Channels["foo"].Value.Publish.Message.Value.Description)
			},
		},
		{
			name: "subscribe message file reference",
			read: func(cfg *dynamic.Config) error {
				require.Equal(t, "/foo.yml", cfg.Info.Url.String())
				config := &Config{Components: &Components{
					Messages: map[string]*Message{"foo": {Description: "foo"}},
				}}
				cfg.Data = config
				return nil
			},
			cfg: &Config{Channels: map[string]*ChannelRef{
				"foo": {Value: &Channel{Subscribe: &Operation{Message: &MessageRef{Ref: "foo.yml#/components/messages/foo"}}}},
			}},
			test: func(t *testing.T, cfg *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cfg.Channels["foo"].Value.Subscribe.Message.Value.Description)
			},
		},
		{
			name: "publish message file reference",
			read: func(cfg *dynamic.Config) error {
				require.Equal(t, "/foo.yml", cfg.Info.Url.String())
				config := &Config{Components: &Components{
					Messages: map[string]*Message{"foo": {Description: "foo"}},
				}}
				cfg.Data = config
				return nil
			},
			cfg: &Config{Channels: map[string]*ChannelRef{
				"foo": {Value: &Channel{Publish: &Operation{Message: &MessageRef{Ref: "foo.yml#/components/messages/foo"}}}},
			}},
			test: func(t *testing.T, cfg *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cfg.Channels["foo"].Value.Publish.Message.Value.Description)
			},
		},
		{
			name: "reader returns error",
			read: func(cfg *dynamic.Config) error {
				return fmt.Errorf("TEST ERROR")
			},
			cfg: &Config{Channels: map[string]*ChannelRef{
				"foo": {Value: &Channel{Subscribe: &Operation{Message: &MessageRef{Ref: "foo.yml#/components/messages/foo"}}}},
			}},
			test: func(t *testing.T, cfg *Config, err error) {
				require.EqualError(t, err, "resolve reference 'foo.yml#/components/messages/foo' failed: TEST ERROR")
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
		target := &Message{}
		reader := &testReader{readFunc: func(cfg *dynamic.Config) error { return nil }}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Value: &Channel{Publish: &Operation{Message: &MessageRef{Ref: "#/components/messages/foo"}}}},
		}, Components: &Components{
			Messages: map[string]*Message{"foo": {}},
		}}
		file := &dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}
		err := config.Parse(file, reader)

		// modify file
		file.Data.(*Config).Components.Messages["foo"] = target

		require.NoError(t, err)
		require.Equal(t, target, config.Channels["foo"].Value.Publish.Message.Value)
	})
}

func TestModifyFileResolve(t *testing.T) {
	target := &Channel{}
	var fooConfig *dynamic.Config
	reader := &testReader{readFunc: func(cfg *dynamic.Config) error {
		require.Equal(t, "/foo.yml", cfg.Info.Url.String())
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Value: &Channel{}},
		}}
		cfg.Data = config
		fooConfig = cfg
		return nil
	}}
	config := &Config{Channels: map[string]*ChannelRef{
		"foo": {Ref: "foo.yml#/channels/foo"},
	}}
	err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
	require.NoError(t, err)

	fooConfig.Data.(*Config).Channels["foo"].Value = target
	err = fooConfig.Data.(dynamic.Parser).Parse(fooConfig, reader)

	require.NoError(t, err)
	require.Equal(t, target, config.Channels["foo"].Value)
}

func TestSchema(t *testing.T) {
	message := &Message{}
	config := &Config{
		Channels: map[string]*ChannelRef{
			"foo": {Value: &Channel{
				Publish: &Operation{
					Message: &MessageRef{
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
		schemas := &schema.Schemas{}
		schemas.Set("foo", &schema.Ref{Value: target})
		config.Components = &Components{Schemas: schemas}
		message.Payload = &schema.Ref{Reference: ref.Reference{Ref: "#/components/Schemas/foo"}}
		reader := &testReader{readFunc: func(cfg *dynamic.Config) error { return nil }}

		err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
		require.NoError(t, err)
		require.Equal(t, target, message.Payload.Value)
	})
	t.Run("file reference direct", func(t *testing.T) {
		target := &schema.Schema{}
		message.Payload = &schema.Ref{Reference: ref.Reference{Ref: "foo.yml"}}
		reader := &testReader{readFunc: func(cfg *dynamic.Config) error {
			cfg.Data = target
			return nil
		}}

		err := config.Parse(&dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
		require.NoError(t, err)
		require.Equal(t, target, message.Payload.Value)
	})
	t.Run("modify file reference direct", func(t *testing.T) {
		target := &schema.Schema{}
		message.Payload = &schema.Ref{Reference: ref.Reference{Ref: "foo.yml"}}
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
		require.Equal(t, target, message.Payload.Value)
	})
}
