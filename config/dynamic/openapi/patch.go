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

func (r RequestBodies) patch(patch RequestBodies) {
	for k, p := range patch {
		if p == nil || p.Value == nil {
			continue
		}
		if v, ok := r[k]; ok && v != nil {
			v.patch(p)
		} else {
			r[k] = p
		}
	}
}
