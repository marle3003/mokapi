package mail

import (
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
)

func (c *Config) Parse(config *dynamic.Config, _ dynamic.Reader) error {
	if c == nil {
		return nil
	}

	log.Warnf("Deprecated mail configuration in %s. This format is deprecated and will be removed in future versions. Please migrate to the new format. More info: https://mokapi.io/docs/email/overview", config.Info.Path())

	converted := c.Convert()
	config.Data = converted
	return nil
}
