package asyncApi

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi/ref"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/test"
	"net/url"
	"testing"
)

type testReader struct {
	readFunc func(file *common.File) error
}

func (tr *testReader) Read(u *url.URL, opts ...common.FileOptions) (*common.File, error) {
	file := &common.File{Url: u}
	for _, opt := range opts {
		opt(file, true)
	}
	if err := tr.readFunc(file); err != nil {
		return file, err
	}
	if p, ok := file.Data.(common.Parser); ok {
		return file, p.Parse(file, tr)
	}
	return file, nil
}

func (tr *testReader) Close() {}

func TestResolve(t *testing.T) {
	t.Run("empty should not error", func(t *testing.T) {
		reader := &testReader{readFunc: func(file *common.File) error { return nil }}
		config := &Config{}
		err := config.Parse(&common.File{Url: &url.URL{}, Data: config}, reader)
		test.Ok(t, err)
	})
}

func TestChannelResolve(t *testing.T) {
	t.Run("nil should not error", func(t *testing.T) {
		reader := &testReader{readFunc: func(file *common.File) error { return nil }}
		config := &Config{Channels: map[string]*ChannelRef{"foo": nil}}
		err := config.Parse(&common.File{Url: &url.URL{}, Data: config}, reader)
		test.Ok(t, err)
	})
	t.Run("file reference", func(t *testing.T) {
		target := &Channel{}
		reader := &testReader{readFunc: func(file *common.File) error {
			test.Equals(t, "/foo.yml#/channels/foo", file.Url.String())
			config := &Config{Channels: map[string]*ChannelRef{
				"foo": {Value: target},
			}}
			file.Data = config
			return nil
		}}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Ref: "foo.yml#/channels/foo"},
		}}
		err := config.Parse(&common.File{Url: &url.URL{}, Data: config}, reader)
		test.Ok(t, err)
		test.Equals(t, target, config.Channels["foo"].Value)
	})
	t.Run("file reference but nil", func(t *testing.T) {
		reader := &testReader{readFunc: func(file *common.File) error {
			test.Equals(t, "/foo.yml#/channels/foo", file.Url.String())
			config := &Config{Channels: map[string]*ChannelRef{
				"foo": {},
			}}
			file.Data = config
			return nil
		}}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Ref: "foo.yml#/channels/foo"},
		}}
		err := config.Parse(&common.File{Url: &url.URL{}, Data: config}, reader)
		test.Ok(t, err)
		test.Equals(t, nil, config.Channels["foo"].Value)
	})
	t.Run("reader returns error", func(t *testing.T) {
		reader := &testReader{readFunc: func(file *common.File) error {
			return test.TestError
		}}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Ref: "foo.yml#/channels/foo"},
		}}
		err := config.Parse(&common.File{Url: &url.URL{}, Data: config}, reader)
		test.Error(t, err)
		require.Equal(t, fmt.Sprintf("unable to read /foo.yml#/channels/foo: %v", test.TestError), err.Error())
	})
}

func TestMessageResolve(t *testing.T) {
	t.Run("subscribe message", func(t *testing.T) {
		target := &Message{}
		reader := &testReader{readFunc: func(file *common.File) error { return nil }}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Value: &Channel{Subscribe: &Operation{Message: &MessageRef{Ref: "#/components/messages/foo"}}}},
		}, Components: &Components{
			Messages: map[string]*Message{"foo": target},
		}}
		err := config.Parse(&common.File{Url: &url.URL{}, Data: config}, reader)
		test.Ok(t, err)
		test.Equals(t, target, config.Channels["foo"].Value.Subscribe.Message.Value)
	})
	t.Run("publish message", func(t *testing.T) {
		target := &Message{}
		reader := &testReader{readFunc: func(file *common.File) error { return nil }}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Value: &Channel{Publish: &Operation{Message: &MessageRef{Ref: "#/components/messages/foo"}}}},
		}, Components: &Components{
			Messages: map[string]*Message{"foo": target},
		}}
		err := config.Parse(&common.File{Url: &url.URL{}, Data: config}, reader)
		test.Ok(t, err)
		test.Equals(t, target, config.Channels["foo"].Value.Publish.Message.Value)
	})
	t.Run("modify file", func(t *testing.T) {
		target := &Message{}
		reader := &testReader{readFunc: func(file *common.File) error { return nil }}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Value: &Channel{Publish: &Operation{Message: &MessageRef{Ref: "#/components/messages/foo"}}}},
		}, Components: &Components{
			Messages: map[string]*Message{"foo": {}},
		}}
		file := &common.File{Url: &url.URL{}, Data: config}
		err := config.Parse(file, reader)

		// modify file
		file.Data.(*Config).Components.Messages["foo"] = target

		test.Ok(t, err)
		test.Equals(t, target, config.Channels["foo"].Value.Publish.Message.Value)
	})
	t.Run("subscribe message", func(t *testing.T) {
		target := &Message{}
		reader := &testReader{readFunc: func(file *common.File) error {
			test.Equals(t, "/foo.yml#/components/messages/foo", file.Url.String())
			config := &Config{Components: &Components{
				Messages: map[string]*Message{"foo": target},
			}}
			file.Data = config
			return nil
		}}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Value: &Channel{Subscribe: &Operation{Message: &MessageRef{Ref: "foo.yml#/components/messages/foo"}}}},
		}}
		err := config.Parse(&common.File{Url: &url.URL{}, Data: config}, reader)
		test.Ok(t, err)
		test.Equals(t, target, config.Channels["foo"].Value.Subscribe.Message.Value)
	})
	t.Run("publish message", func(t *testing.T) {
		target := &Message{}
		reader := &testReader{readFunc: func(file *common.File) error {
			test.Equals(t, "/foo.yml#/components/messages/foo", file.Url.String())
			config := &Config{Components: &Components{
				Messages: map[string]*Message{"foo": target},
			}}
			file.Data = config
			return nil
		}}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Value: &Channel{Publish: &Operation{Message: &MessageRef{Ref: "foo.yml#/components/messages/foo"}}}},
		}}
		err := config.Parse(&common.File{Url: &url.URL{}, Data: config}, reader)
		test.Ok(t, err)
		test.Equals(t, target, config.Channels["foo"].Value.Publish.Message.Value)
	})
	t.Run("subscribe reader returns error", func(t *testing.T) {
		reader := &testReader{readFunc: func(file *common.File) error {
			return test.TestError
		}}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Value: &Channel{Subscribe: &Operation{Message: &MessageRef{Ref: "foo.yml#/components/messages/foo"}}}},
		}}
		err := config.Parse(&common.File{Url: &url.URL{}, Data: config}, reader)
		test.Error(t, err)
		require.Equal(t, fmt.Sprintf("unable to read /foo.yml#/components/messages/foo: %v", test.TestError), err.Error())
	})
	t.Run("publisher reader returns error", func(t *testing.T) {
		reader := &testReader{readFunc: func(file *common.File) error {
			return test.TestError
		}}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Value: &Channel{Publish: &Operation{Message: &MessageRef{Ref: "foo.yml#/components/messages/foo"}}}},
		}}
		err := config.Parse(&common.File{Url: &url.URL{}, Data: config}, reader)
		test.Error(t, err)
		require.Equal(t, fmt.Sprintf("unable to read /foo.yml#/components/messages/foo: %v", test.TestError), err.Error())
	})
}

func TestFileResolve(t *testing.T) {
	t.Run("modify file", func(t *testing.T) {
		target := &Channel{}
		var fooFile *common.File
		reader := &testReader{readFunc: func(file *common.File) error {
			test.Equals(t, "/foo.yml#/channels/foo", file.Url.String())
			config := &Config{Channels: map[string]*ChannelRef{
				"foo": {Value: &Channel{}},
			}}
			file.Data = config
			fooFile = file
			return nil
		}}
		config := &Config{Channels: map[string]*ChannelRef{
			"foo": {Ref: "foo.yml#/channels/foo"},
		}}
		err := config.Parse(&common.File{Url: &url.URL{}, Data: config}, reader)
		test.Ok(t, err)

		fooFile.Data.(*Config).Channels["foo"].Value = target
		err = fooFile.Data.(common.Parser).Parse(fooFile, reader)

		test.Ok(t, err)
		test.Equals(t, target, config.Channels["foo"].Value)
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
		reader := &testReader{readFunc: func(file *common.File) error { return nil }}
		err := config.Parse(&common.File{Url: &url.URL{}, Data: config}, reader)
		test.Ok(t, err)
	})
	t.Run("reference inside", func(t *testing.T) {
		target := &schema.Schema{}
		schemas := &schema.Schemas{}
		schemas.Set("foo", &schema.Ref{Value: target})
		config.Components = &Components{Schemas: schemas}
		message.Payload = &schema.Ref{Reference: ref.Reference{Value: "#/components/Schemas/foo"}}
		reader := &testReader{readFunc: func(file *common.File) error { return nil }}

		err := config.Parse(&common.File{Url: &url.URL{}, Data: config}, reader)
		test.Ok(t, err)
		test.Equals(t, target, message.Payload.Value)
	})
	t.Run("file reference direct", func(t *testing.T) {
		target := &schema.Schema{}
		message.Payload = &schema.Ref{Reference: ref.Reference{Value: "foo.yml"}}
		reader := &testReader{readFunc: func(file *common.File) error {
			file.Data = target
			return nil
		}}

		err := config.Parse(&common.File{Url: &url.URL{}, Data: config}, reader)
		test.Ok(t, err)
		test.Equals(t, target, message.Payload.Value)
	})
	t.Run("modify file reference direct", func(t *testing.T) {
		target := &schema.Schema{}
		message.Payload = &schema.Ref{Reference: ref.Reference{Value: "foo.yml"}}
		var fooFile *common.File
		reader := &testReader{readFunc: func(file *common.File) error {
			file.Data = &schema.Schema{}
			fooFile = file
			return nil
		}}

		err := config.Parse(&common.File{Url: &url.URL{}, Data: config}, reader)
		test.Ok(t, err)

		// modify
		fooFile.Data = target
		err = fooFile.Data.(common.Parser).Parse(fooFile, reader)

		test.Ok(t, err)
		test.Equals(t, target, message.Payload.Value)
	})
}
