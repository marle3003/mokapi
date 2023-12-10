package asyncApi

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi/ref"
	"mokapi/config/dynamic/openapi/schema"
	"net/url"
	"testing"
)

type testReader struct {
	readFunc func(cfg *common.Config) error
}

func (tr *testReader) Read(u *url.URL, opts ...common.ConfigOptions) (*common.Config, error) {
	cfg := common.NewConfig(common.ConfigInfo{Url: u}, opts...)
	for _, opt := range opts {
		opt(cfg, true)
	}
	if err := tr.readFunc(cfg); err != nil {
		return cfg, err
	}
	if p, ok := cfg.Data.(common.Parser); ok {
		return cfg, p.Parse(cfg, tr)
	}
	return cfg, nil
}

func (tr *testReader) Close() {}

func TestResolve(t *testing.T) {
	t.Run("empty should not error", func(t *testing.T) {
		reader := &testReader{readFunc: func(cfg *common.Config) error { return nil }}
		config := &Config{}
		err := config.Parse(&common.Config{Info: common.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
		require.NoError(t, err)
	})
}

func TestChannelResolve(t *testing.T) {
	t.Run("nil should not error", func(t *testing.T) {
		reader := &testReader{readFunc: func(cfg *common.Config) error { return nil }}
		config := &Config{Channels: map[string]*ChannelRef{"foo": nil}}
		err := config.Parse(common.NewConfig(common.ConfigInfo{Url: &url.URL{}}, common.WithData(config)), reader)
		require.NoError(t, err)
	})
	t.Run("file reference", func(t *testing.T) {
		target := &Channel{}
		reader := &testReader{readFunc: func(cfg *common.Config) error {
			require.Equal(t, "/foo.yml", cfg.Info.Url.String())
			config := &Config{Channels: map[string]*ChannelRef{
				"foo": {Value: target},
			}}
			cfg.Data = config
			return nil
		}}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Ref: "foo.yml#/channels/foo"},
		}}
		err := config.Parse(common.NewConfig(common.ConfigInfo{Url: &url.URL{}}, common.WithData(config)), reader)
		require.NoError(t, err)
		require.Equal(t, target, config.Channels["foo"].Value)
	})
	t.Run("file reference but nil", func(t *testing.T) {
		reader := &testReader{readFunc: func(file *common.Config) error {
			require.Equal(t, "/foo.yml", file.Info.Url.String())
			config := &Config{Channels: map[string]*ChannelRef{
				"foo": {},
			}}
			file.Data = config
			return nil
		}}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Ref: "foo.yml#/channels/foo"},
		}}
		err := config.Parse(common.NewConfig(common.ConfigInfo{Url: &url.URL{}}, common.WithData(config)), reader)
		require.NoError(t, err)
		require.Nil(t, config.Channels["foo"].Value)
	})
	t.Run("reader returns error", func(t *testing.T) {
		reader := &testReader{readFunc: func(cfg *common.Config) error {
			return fmt.Errorf("TEST ERROR")
		}}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Ref: "foo.yml#/channels/foo"},
		}}
		err := config.Parse(common.NewConfig(common.ConfigInfo{Url: &url.URL{}}, common.WithData(config)), reader)
		require.EqualError(t, err, "resolve reference 'foo.yml#/channels/foo' failed: TEST ERROR")
	})
}

func TestMessageResolve(t *testing.T) {
	t.Run("subscribe message", func(t *testing.T) {
		target := &Message{}
		reader := &testReader{readFunc: func(cfg *common.Config) error { return nil }}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Value: &Channel{Subscribe: &Operation{Message: &MessageRef{Ref: "#/components/messages/foo"}}}},
		}, Components: &Components{
			Messages: map[string]*Message{"foo": target},
		}}
		err := config.Parse(common.NewConfig(common.ConfigInfo{Url: &url.URL{}}, common.WithData(config)), reader)
		require.NoError(t, err)
		require.Equal(t, target, config.Channels["foo"].Value.Subscribe.Message.Value)
	})
	t.Run("publish message", func(t *testing.T) {
		target := &Message{}
		reader := &testReader{readFunc: func(cfg *common.Config) error { return nil }}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Value: &Channel{Publish: &Operation{Message: &MessageRef{Ref: "#/components/messages/foo"}}}},
		}, Components: &Components{
			Messages: map[string]*Message{"foo": target},
		}}
		err := config.Parse(common.NewConfig(common.ConfigInfo{Url: &url.URL{}}, common.WithData(config)), reader)
		require.NoError(t, err)
		require.Equal(t, target, config.Channels["foo"].Value.Publish.Message.Value)
	})
	t.Run("modify file", func(t *testing.T) {
		target := &Message{}
		reader := &testReader{readFunc: func(cfg *common.Config) error { return nil }}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Value: &Channel{Publish: &Operation{Message: &MessageRef{Ref: "#/components/messages/foo"}}}},
		}, Components: &Components{
			Messages: map[string]*Message{"foo": {}},
		}}
		file := common.NewConfig(common.ConfigInfo{Url: &url.URL{}}, common.WithData(config))
		err := config.Parse(file, reader)

		// modify file
		file.Data.(*Config).Components.Messages["foo"] = target

		require.NoError(t, err)
		require.Equal(t, target, config.Channels["foo"].Value.Publish.Message.Value)
	})
	t.Run("subscribe message", func(t *testing.T) {
		target := &Message{}
		reader := &testReader{readFunc: func(file *common.Config) error {
			require.Equal(t, "/foo.yml", file.Info.Url.String())
			config := &Config{Components: &Components{
				Messages: map[string]*Message{"foo": target},
			}}
			file.Data = config
			return nil
		}}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Value: &Channel{Subscribe: &Operation{Message: &MessageRef{Ref: "foo.yml#/components/messages/foo"}}}},
		}}
		err := config.Parse(common.NewConfig(common.ConfigInfo{Url: &url.URL{}}, common.WithData(config)), reader)
		require.NoError(t, err)
		require.Equal(t, target, config.Channels["foo"].Value.Subscribe.Message.Value)
	})
	t.Run("publish message", func(t *testing.T) {
		target := &Message{}
		reader := &testReader{readFunc: func(file *common.Config) error {
			require.Equal(t, "/foo.yml", file.Info.Url.String())
			config := &Config{Components: &Components{
				Messages: map[string]*Message{"foo": target},
			}}
			file.Data = config
			return nil
		}}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Value: &Channel{Publish: &Operation{Message: &MessageRef{Ref: "foo.yml#/components/messages/foo"}}}},
		}}
		err := config.Parse(common.NewConfig(common.ConfigInfo{Url: &url.URL{}}, common.WithData(config)), reader)
		require.NoError(t, err)
		require.Equal(t, target, config.Channels["foo"].Value.Publish.Message.Value)
	})
	t.Run("subscribe reader returns error", func(t *testing.T) {
		reader := &testReader{readFunc: func(file *common.Config) error {
			return fmt.Errorf("TEST ERROR")
		}}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Value: &Channel{Subscribe: &Operation{Message: &MessageRef{Ref: "foo.yml#/components/messages/foo"}}}},
		}}
		err := config.Parse(common.NewConfig(common.ConfigInfo{Url: &url.URL{}}, common.WithData(config)), reader)
		require.EqualError(t, err, "resolve reference 'foo.yml#/components/messages/foo' failed: TEST ERROR")
	})
	t.Run("publisher reader returns error", func(t *testing.T) {
		reader := &testReader{readFunc: func(cfg *common.Config) error {
			return fmt.Errorf("TEST ERROR")
		}}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Value: &Channel{Publish: &Operation{Message: &MessageRef{Ref: "foo.yml#/components/messages/foo"}}}},
		}}
		err := config.Parse(common.NewConfig(common.ConfigInfo{Url: &url.URL{}}, common.WithData(config)), reader)
		require.EqualError(t, err, "resolve reference 'foo.yml#/components/messages/foo' failed: TEST ERROR")
	})
}

func TestFileResolve(t *testing.T) {
	t.Run("modify file", func(t *testing.T) {
		target := &Channel{}
		var fooConfig *common.Config
		reader := &testReader{readFunc: func(cfg *common.Config) error {
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
		err := config.Parse(common.NewConfig(common.ConfigInfo{Url: &url.URL{}}, common.WithData(config)), reader)
		require.NoError(t, err)

		fooConfig.Data.(*Config).Channels["foo"].Value = target
		err = fooConfig.Data.(common.Parser).Parse(fooConfig, reader)

		require.NoError(t, err)
		require.Equal(t, target, config.Channels["foo"].Value)
	})

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
		reader := &testReader{readFunc: func(file *common.Config) error { return nil }}
		err := config.Parse(&common.Config{Info: common.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
		require.NoError(t, err)
	})
	t.Run("reference inside", func(t *testing.T) {
		target := &schema.Schema{}
		schemas := &schema.Schemas{}
		schemas.Set("foo", &schema.Ref{Value: target})
		config.Components = &Components{Schemas: schemas}
		message.Payload = &schema.Ref{Reference: ref.Reference{Ref: "#/components/Schemas/foo"}}
		reader := &testReader{readFunc: func(cfg *common.Config) error { return nil }}

		err := config.Parse(&common.Config{Info: common.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
		require.NoError(t, err)
		require.Equal(t, target, message.Payload.Value)
	})
	t.Run("file reference direct", func(t *testing.T) {
		target := &schema.Schema{}
		message.Payload = &schema.Ref{Reference: ref.Reference{Ref: "foo.yml"}}
		reader := &testReader{readFunc: func(cfg *common.Config) error {
			cfg.Data = target
			return nil
		}}

		err := config.Parse(&common.Config{Info: common.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
		require.NoError(t, err)
		require.Equal(t, target, message.Payload.Value)
	})
	t.Run("modify file reference direct", func(t *testing.T) {
		target := &schema.Schema{}
		message.Payload = &schema.Ref{Reference: ref.Reference{Ref: "foo.yml"}}
		var fooConfig *common.Config
		reader := &testReader{readFunc: func(file *common.Config) error {
			file.Data = &schema.Schema{}
			fooConfig = file
			return nil
		}}

		err := config.Parse(&common.Config{Info: common.ConfigInfo{Url: &url.URL{}}, Data: config}, reader)
		require.NoError(t, err)

		// modify
		fooConfig.Data = target
		err = fooConfig.Data.(common.Parser).Parse(fooConfig, reader)

		require.NoError(t, err)
		require.Equal(t, target, message.Payload.Value)
	})
}
