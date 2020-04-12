package file

import (
	"errors"
	"mokapi/config"
	"mokapi/config/decoders"
)

type Provider struct {
	Filename string
}

func (p *Provider) Provide() (*config.Api, error) {
	return p.LoadConfig()
}

func (p *Provider) LoadConfig() (*config.Api, error) {
	if len(p.Filename) > 0 {
		api, error := p.loadFileConfig(p.Filename)
		if error != nil {
			return nil, error
		}

		return api, nil
	}

	return nil, errors.New("error using file configuration provider, but no filename defined")
}

func (p *Provider) loadFileConfig(filename string) (*config.Api, error) {
	configDecoders := []decoders.ConfigDecoder{decoders.NewFileDecoder(p.Filename)}
	api := &config.Api{}
	error := config.Load(configDecoders, api)
	if error != nil {
		return nil, error
	}
	return api, nil
}
