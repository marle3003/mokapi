package jstest

import (
	"mokapi/config/dynamic"
	"mokapi/js"
	"mokapi/js/require"
	"net/url"
	"time"
)

type Script struct {
	*js.Script
}

func New(opts ...js.Option) (*Script, error) {
	registry, err := require.NewRegistry(func(file, hint string) (*dynamic.Config, error) {
		return nil, require.ModuleFileNotFound
	})
	js.RegisterNativeModules(registry)

	opts = append([]js.Option{js.WithRegistry(registry)}, opts...)
	s, err := js.NewScript(opts...)
	if err != nil {
		return nil, err
	}
	return &Script{Script: s}, nil
}

func WithSource(source string) js.Option {
	return WithPathSource("test.js", source)
}

func WithPathSource(path, source string) js.Option {
	file := &dynamic.Config{
		Info: dynamic.ConfigInfo{
			Provider: "test",
			Url:      mustParse(path),
			Checksum: nil,
			Time:     time.Time{},
		},
		Raw:       []byte(source),
		Data:      nil,
		Refs:      dynamic.Refs{},
		Listeners: dynamic.Listeners{},
	}
	return js.WithFile(file)
}

func mustParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
