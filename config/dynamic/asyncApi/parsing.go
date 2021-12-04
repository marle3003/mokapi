package asyncApi

import (
	"mokapi/config/dynamic/common"
)

func (c *Config) Parse(file *common.File, reader common.Reader) error {
	for _, ch := range c.Channels {
		if ch == nil {
			continue
		}
		if err := ch.Parse(file, reader); err != nil {
			return err
		}
	}

	return nil
}

func (c *ChannelRef) Parse(file *common.File, reader common.Reader) error {
	if len(c.Ref) > 0 && c.Value == nil {
		if err := common.Resolve(c.Ref, &c.Value, file, reader); err != nil {
			return err
		}
	}

	if c.Value == nil {
		return nil
	}

	if c.Value.Publish != nil {
		if err := c.Value.Publish.Parse(file, reader); err != nil {
			return err
		}
	}

	if c.Value.Subscribe != nil {
		if err := c.Value.Subscribe.Parse(file, reader); err != nil {
			return err
		}
	}

	return nil
}

func (o *Operation) Parse(file *common.File, reader common.Reader) error {
	if o.Message != nil {
		return o.Message.Parse(file, reader)
	}
	return nil
}

func (m *MessageRef) Parse(file *common.File, reader common.Reader) error {
	if len(m.Ref) > 0 && m.Value == nil {
		if err := common.Resolve(m.Ref, &m.Value, file, reader); err != nil {
			return err
		}
	}

	if m.Value == nil {
		return nil
	}

	if m.Value.Payload != nil {
		if err := m.Value.Payload.Parse(file, reader); err != nil {
			return err
		}
	}

	return nil
}

// -------

//type ReferenceResolver struct {
//	reader common.Reader
//	file   *dynamic.File
//	config *Config
//}
//
//func resolve(file *dynamic.File, reader dynamic.Reader) error {
//	config, ok := file.Data.(*Config)
//	if !ok {
//		return fmt.Errorf("unexpected config %v", reflect.TypeOf(file.Data).String())
//	}
//
//	r := &ReferenceResolver{
//		reader: reader,
//		config: config,
//		file:   file,
//	}
//
//	return r.Resolve()
//}
//
//func (r ReferenceResolver) Resolve() error {
//	for _, ch := range r.config.Channels {
//		if err := r.resolveChannelRef(ch); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (r ReferenceResolver) resolveChannelRef(m *ChannelRef) error {
//	if m == nil {
//		return nil
//	}
//
//	if len(m.Ref) > 0 && m.Value == nil {
//		if err := r.resolve(m.Ref, &m.Value); err != nil {
//			return err
//		}
//	}
//
//	if m.Value == nil {
//		return nil
//	}
//
//	if m.Value.Publish != nil {
//		if err := r.resolveMessageRef(m.Value.Publish.Message); err != nil {
//			return err
//		}
//	}
//
//	if m.Value.Subscribe != nil {
//		if err := r.resolveMessageRef(m.Value.Subscribe.Message); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (r ReferenceResolver) resolveMessageRef(m *MessageRef) error {
//	if m == nil {
//		return nil
//	}
//
//	if len(m.Ref) > 0 && m.Value == nil {
//		if err := r.resolve(m.Ref, &m.Value); err != nil {
//			return err
//		}
//	}
//
//	if m.Value == nil {
//		return nil
//	}
//
//	return r.resolveSchemaRef(m.Value.Payload)
//}
//
//func (r ReferenceResolver) resolveSchemas(s *openapi.Schemas) error {
//	if s == nil {
//		return nil
//	}
//
//	if len(s.Ref) > 0 && s.Value == nil {
//		if err := r.resolve(s.Ref, &s.Value); err != nil {
//			return err
//		}
//	}
//
//	if s.Value == nil {
//		return nil
//	}
//
//	for _, child := range s.Value {
//		if err := r.resolveSchemaRef(child); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (r ReferenceResolver) resolveSchemaRef(s *openapi.SchemaRef) error {
//	if s == nil {
//		return nil
//	}
//
//	if len(s.Ref) > 0 && s.Value == nil {
//		if err := r.resolve(s.Ref, &s.Value); err != nil {
//			return err
//		}
//	}
//
//	if s.Value == nil {
//		return nil
//	}
//
//	if err := r.resolveSchemaRef(s.Value.Items); err != nil {
//		return err
//	}
//
//	if err := r.resolveSchemas(s.Value.Properties); err != nil {
//		return err
//	}
//
//	if err := r.resolveSchemaRef(s.Value.AdditionalProperties); err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func (r ReferenceResolver) resolve(ref string, val interface{}) (err error) {
//	u, err := url.Parse(ref)
//	if err != nil {
//		return
//	}
//
//	var i interface{}
//	if len(u.Path) > 0 {
//		if !u.IsAbs() {
//			u, err = r.file.Url.Parse(ref)
//		}
//
//		var f *dynamic.File
//		switch s := strings.ToLower(u.Fragment); {
//		case strings.HasPrefix(s, "/components"), strings.HasPrefix(s, "/channels"):
//			f, err = r.reader.Read(u, dynamic.WithData(&Config{}), dynamic.WithParent(r.file), dynamic.WithInitializer(func(file *dynamic.File, reader dynamic.Reader) error {
//				err := resolve(file, reader)
//				if err != nil {
//					return err
//				}
//				return common.ResolvePath(u.Fragment, file.Data, val)
//			}))
//			return err
//		case len(s) == 0:
//			f, err = r.reader.Read(u, dynamic.WithData(val), dynamic.WithParent(r.file), dynamic.WithInitializer(func(file *dynamic.File, reader dynamic.Reader) error {
//				// TODO
//				//err := resolve(file, reader)
//				//if err != nil {
//				//	return err
//				//}
//				return common.ResolvePath(u.Fragment, file.Data, val)
//			}))
//		default:
//			switch val.(type) {
//			case **openapi.SchemaRef:
//				schemas := &openapi.Schemas{}
//				f, err = r.reader.Read(u, dynamic.WithData(schemas.Value))
//			}
//		}
//
//		if err != nil {
//			return
//		}
//		i = f.Data
//	} else {
//		i = r.config
//	}
//
//	return common.ResolvePath(u.Fragment, i, val)
//}
//
//func (r *MessageRef) Resolve(token string) (interface{}, error) {
//	return common.Get(token, r.Value)
//}
//
//func (c *ChannelRef) Resolve(token string) (interface{}, error) {
//	return common.Get(token, c.Value)
//}
