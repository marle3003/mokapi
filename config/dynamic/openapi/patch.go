package openapi

func (c *Config) Patch(patch *Config) {
	c.Info.patch(patch.Info)
	c.patchServers(patch.Servers)
	c.Paths.patch(patch.Paths)
	c.patchComponents(patch)
}

func (r *RequestBodyRef) patch(patch *RequestBodyRef) {
	if patch == nil || patch.Value == nil {
		return
	}
	if r.Value == nil {
		r.Value = patch.Value
	} else {
		r.Value.patch(patch.Value)
	}
}

func (r *RequestBody) patch(patch *RequestBody) {
	if len(patch.Description) > 0 {
		r.Description = patch.Description
	}
	r.Required = patch.Required

	if len(r.Content) == 0 {
		r.Content = patch.Content
		return
	}

	r.Content.patch(patch.Content)
}

func (c *Content) patch(patch Content) {
	for k, v := range patch {
		if con, ok := (*c)[k]; ok {
			con.patch(v)
		} else {
			(*c)[k] = v
		}
	}
}

func (c *MediaType) patch(patch *MediaType) {
	if c.Schema == nil {
		c.Schema = patch.Schema
	} else {
		c.Schema.Patch(patch.Schema)
	}

	if c.Example == nil {
		c.Example = patch.Example
	}

	for k, v := range patch.Examples {
		if e, ok := c.Examples[k]; ok {
			e.patch(v)
		} else {
			c.Examples[k] = v
		}
	}
}

func (r *ExampleRef) patch(patch *ExampleRef) {
	if patch == nil || patch.Value == nil {
		return
	}
	if r.Value == nil {
		r.Value = patch.Value
		return
	}

	if len(patch.Value.Summary) > 0 {
		r.Value.Summary = patch.Value.Summary
	}

	if r.Value.Value == nil {
		r.Value.Value = patch.Value.Value
	}

	if len(patch.Value.Description) > 0 {
		r.Value.Description = patch.Value.Description
	}
}

func (r *Response) patch(patch *Response) {
	if patch == nil {
		return
	}

	if len(patch.Description) > 0 {
		r.Description = patch.Description
	}

	if r.Content == nil {
		r.Content = patch.Content
	} else {
		r.Content.patch(patch.Content)
	}

	if len(r.Headers) == 0 {
		r.Headers = patch.Headers
	} else {

	}
}

func (r *HeaderRef) patch(patch *HeaderRef) {
	if patch == nil || patch.Value == nil {
		return
	}

	if r.Value == nil {
		r.Value = patch.Value
	} else {
		r.Value.patch(patch.Value)
	}
}

func (h *Header) patch(patch *Header) {
	if len(patch.Name) > 0 {
		h.Name = patch.Name
	}
	if len(patch.Description) > 0 {
		h.Description = patch.Description
	}
	if h.Schema == nil {
		h.Schema = patch.Schema
	} else {
		h.Schema.Patch(patch.Schema)
	}
}

func (c *Config) patchComponents(patch *Config) {
	if c.Components.Schemas == nil {
		c.Components.Schemas = patch.Components.Schemas
	} else {
		c.Components.Schemas.Patch(patch.Components.Schemas)
	}
	if c.Components.Responses == nil {
		c.Components.Responses = patch.Components.Responses
	} else {
		c.Components.Responses.patch(patch.Components.Responses)
	}
	if c.Components.RequestBodies == nil {
		c.Components.RequestBodies = patch.Components.RequestBodies
	} else {
		c.Components.RequestBodies.patch(patch.Components.RequestBodies)
	}
	if c.Components.Parameters == nil {
		c.Components.Parameters = patch.Components.Parameters
	} else {
		c.Components.Parameters.Patch(patch.Components.Parameters)
	}
	if c.Components.Examples == nil {
		c.Components.Examples = patch.Components.Examples
	} else {
		c.Components.Examples.patch(patch.Components.Examples)
	}
	if c.Components.Headers == nil {
		c.Components.Headers = patch.Components.Headers
	} else {
		c.Components.Headers.patch(patch.Components.Headers)
	}
}

func (r *NamedResponses) patch(patch *NamedResponses) {
	if patch == nil || patch.Value == nil {
		return
	}
	if r.Value == nil {
		r.Value = patch.Value
	}
	for k, p := range patch.Value {
		if p.Value == nil {
			continue
		}
		if v, ok := r.Value[k]; ok {
			if v.Value == nil {
				v.Value = p.Value
			} else {
				v.Value.patch(p.Value)
			}
		} else {
			r.Value[k] = p
		}
	}
}

func (r *RequestBodies) patch(patch *RequestBodies) {
	if patch == nil || patch.Value == nil {
		return
	}
	if r.Value == nil {
		r.Value = patch.Value
	}
	for k, p := range patch.Value {
		if p.Value == nil {
			continue
		}
		if v, ok := r.Value[k]; ok {
			if v.Value == nil {
				v.Value = p.Value
			} else {
				v.Value.patch(p.Value)
			}
		} else {
			r.Value[k] = p
		}
	}
}

func (r *Examples) patch(patch *Examples) {
	if patch == nil || patch.Value == nil {
		return
	}
	if r.Value == nil {
		r.Value = patch.Value
	}
	for k, p := range patch.Value {
		if p.Value == nil {
			continue
		}
		if v, ok := r.Value[k]; ok {
			if v.Value == nil {
				v.Value = p.Value
			} else {
				v.patch(p)
			}
		} else {
			r.Value[k] = p
		}
	}
}

func (r *NamedHeaders) patch(patch *NamedHeaders) {
	if patch == nil || patch.Value == nil {
		return
	}
	if r.Value == nil {
		r.Value = patch.Value
	}
	for k, p := range patch.Value {
		if p.Value == nil {
			continue
		}
		if v, ok := r.Value[k]; ok {
			if v.Value == nil {
				v.Value = p.Value
			} else {
				v.patch(p)
			}
		} else {
			r.Value[k] = p
		}
	}
}
