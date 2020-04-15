package file

import (
	"errors"
	"mokapi/config"
	"mokapi/config/decoders"
)

type Provider struct {
	Filename string
}

func (p *Provider) Provide(element interface{}) error {
	return p.loadConfig(element)
}

func (p *Provider) loadConfig(element interface{}) error {
	if len(p.Filename) > 0 {
		error := p.loadFileConfig(p.Filename, element)
		if error != nil {
			return error
		}

		return nil
	}

	return errors.New("error using file configuration provider, but no filename defined")
}

func (p *Provider) loadFileConfig(filename string, element interface{}) error {
	configDecoders := []decoders.ConfigDecoder{decoders.NewFileDecoder(p.Filename)}
	error := config.Load(configDecoders, element)
	if error != nil {
		return error
	}
	return nil
}
