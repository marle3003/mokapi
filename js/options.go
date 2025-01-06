package js

import (
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/js/require"
)

type Option func(*Script)

func WithHost(host common.Host) Option {
	return func(script *Script) {
		script.host = host
	}
}

func WithRegistry(registry *require.Registry) Option {
	return func(script *Script) {
		script.registry = registry
	}
}

func WithFile(file *dynamic.Config) Option {
	return func(script *Script) {
		script.file = file
	}
}
