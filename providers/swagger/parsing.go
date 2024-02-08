package swagger

import (
	"mokapi/config/dynamic"
)

func (c *Config) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if c == nil {
		return nil
	}

	converted, err := Convert(c)
	if err != nil {
		return nil
	}
	config.Data = converted
	return converted.Parse(config, reader)
}
