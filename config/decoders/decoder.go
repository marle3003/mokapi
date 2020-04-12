package decoders

type ConfigDecoder interface {
	Decode(flags map[string]string, element interface{}) error
}
