package mail

import "mokapi/config/dynamic"

func (c *Config) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if c == nil {
		return nil
	}

	converted := c.Convert()
	config.Data = converted
	return nil
}
