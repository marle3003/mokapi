package swagger

import (
	"mokapi/config/dynamic/common"
)

func (c *Config) Parse(config *common.Config, reader common.Reader) error {
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
