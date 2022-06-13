package swagger

import (
	"mokapi/config/dynamic/common"
)

func (c *Config) Parse(config *common.Config, reader common.Reader) error {
	if c == nil {
		return nil
	}

	//for _, p := range c.Paths {
	//	p.Parse(config, reader)
	//}

	converted, err := Convert(c)
	if err != nil {
		return nil
	}
	config.Data = converted
	return converted.Parse(config, reader)
}

func (p *PathItem) Parse(config *common.Config, reader common.Reader) error {
	if len(p.Ref) > 0 {
		p2 := &PathItem{}
		if err := common.Resolve(p.Ref, &p2, config, reader); err != nil {
			return err
		}
		*p = *p2
		return nil
	}

	for _, o := range p.Operations() {
		o.Parse(config, reader)
	}

	return nil
}

func (o *Operation) Parse(config *common.Config, reader common.Reader) error {
	for _, p := range o.Parameters {
		p.Parse(config, reader)
	}
	for _, r := range o.Responses {
		r.Parse(config, reader)
	}

	return nil
}

func (p *Parameter) Parse(config *common.Config, reader common.Reader) error {
	if err := p.Schema.Parse(config, reader); err != nil {
		return err
	}
	if err := p.Items.Parse(config, reader); err != nil {
		return err
	}

	return nil
}

func (r *Response) Parse(config *common.Config, reader common.Reader) error {
	if err := r.Schema.Parse(config, reader); err != nil {
		return err
	}

	return nil
}
