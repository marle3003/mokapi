package decoders

type ConfigDecoder interface {
	Decode(flags map[string]string, element interface{}) error
}

func Load(decoders []ConfigDecoder, config interface{}) error {
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
