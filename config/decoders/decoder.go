package decoders

type ConfigDecoder interface {
	Decode(flags map[string]string, element interface{}) error
}

func Load(decoders []ConfigDecoder, config interface{}) error {
	flags, err := parseFlags()
	if err != nil {
		return err
	}

	for _, decoder := range decoders {
		err := decoder.Decode(flags, config)
		if err != nil {
			return err
		}
	}

	return nil
}
