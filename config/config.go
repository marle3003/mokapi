package config

import "mokapi/config/decoders"

func Load(decoders []decoders.ConfigDecoder, config interface{}) error {
	flags, error := parseFlags()
	if error != nil {
		return error
	}

	for _, decoder := range decoders {
		error := decoder.Decode(flags, config)
		if error != nil {
			return error
		}
	}

	return nil
}
