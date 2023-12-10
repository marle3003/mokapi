package common

type ConfigOptions func(config *Config, init bool)

func WithRaw(raw []byte) ConfigOptions {
	return func(config *Config, init bool) {
		config.Raw = raw
	}
}

func WithData(data interface{}) ConfigOptions {
	return func(file *Config, init bool) {
		if !init {
			return
		}
		file.Data = data
	}
}

func WithParent(parent *Config) ConfigOptions {
	return func(file *Config, init bool) {
		file.AddListener(parent.Info.Url.String(), func(_ *Config) {
			parent.Changed()
		})
	}
}

func WithChecksum(b []byte) ConfigOptions {
	return func(config *Config, init bool) {
		config.Checksum = b
	}
}
