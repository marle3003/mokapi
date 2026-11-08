package openapi

type Encoding struct {
	ContentType   string            `yaml:"contentType,omitempty" json:"contentType,omitempty"`
	Headers       map[string]Header `yaml:"headers,omitempty" json:"headers,omitempty"`
	Style         string            `yaml:"style,omitempty" json:"style,omitempty"`
	Explode       bool              `yaml:"explode,omitempty" json:"explode,omitempty"`
	AllowReserved bool              `yaml:"allowReserved,omitempty" json:"allowReserved,omitempty"`
}

func (e *Encoding) patch(patch *Encoding) {
	if patch == nil {
		return
	}
	if len(patch.ContentType) > 0 {
		e.ContentType = patch.ContentType
	}
	if len(patch.Style) > 0 {
		e.Style = patch.Style
	}
	e.Explode = patch.Explode
	e.AllowReserved = patch.AllowReserved

	if e.Headers == nil {
		e.Headers = cloneMap(patch.Headers, cloneHeader)
		return
	}
	for name, h := range patch.Headers {
		// map values are not addressable
		existing, ok := e.Headers[name]
		if ok {
			existing.patch(&h)
			h = existing
		} else {
			h = cloneHeader(h)
		}
		e.Headers[name] = h
	}
}

func (e *Encoding) clone() *Encoding {
	if e == nil {
		return nil
	}
	c := *e
	c.Headers = cloneMap(e.Headers, cloneHeader)
	return &c
}

func cloneHeader(h Header) Header {
	return *h.clone()
}
