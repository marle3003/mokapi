package openapi

func (c *Config) Patch(patch *Config) {
	c.Info.patch(patch.Info)
	c.patchServers(patch.Servers)
	c.Paths.patch(patch.Paths)
	c.patchComponents(patch)
}

func (c *Info) patch(patch Info) {
	if len(patch.Description) > 0 {
		c.Description = patch.Description
	}
	if c.Contact == nil {
		c.Contact = patch.Contact
	} else {
		c.Contact.patch(patch.Contact)
	}
	if len(patch.Version) > 0 {
		c.Version = patch.Version
	}
}

func (c *Contact) patch(patch *Contact) {
	if patch == nil {
		return
	}
	if len(patch.Name) > 0 {
		c.Name = patch.Name
	}
	if len(patch.Url) > 0 {
		c.Url = patch.Url
	}
	if len(patch.Email) > 0 {
		c.Email = patch.Email
	}
}

func (c *Config) patchServers(patch []*Server) {
	if len(patch) == 0 {
		return
	}
	if len(c.Servers) == 0 {
		c.Servers = patch
	}

LoopPatch:
	for _, p := range patch {
		for _, s := range c.Servers {
			if s.Url == p.Url {
				if len(p.Description) > 0 {
					s.Description = p.Description
				}
				continue LoopPatch
			}
		}
		c.Servers = append(c.Servers, p)
	}
}

func (r *EndpointsRef) patch(patch EndpointsRef) {
	if patch.Value == nil {
		return
	}
	if r.Value == nil {
		r.Value = patch.Value
		return
	}

	for path, v := range patch.Value {
		if e, ok := r.Value[path]; ok {
			e.patch(v)
		} else {
			r.Value[path] = v
		}
	}
}

func (r *EndpointRef) patch(patch *EndpointRef) {
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

	if len(patch.Value.Description) > 0 {
		r.Value.Description = patch.Value.Description
	}

	if r.Value.Get == nil {
		r.Value.Get = patch.Value.Get
	} else {
		r.Value.Get.patch(patch.Value.Get)
	}

	if r.Value.Post == nil {
		r.Value.Post = patch.Value.Post
	} else {
		r.Value.Post.patch(patch.Value.Post)
	}

	if r.Value.Put == nil {
		r.Value.Put = patch.Value.Put
	} else {
		r.Value.Put.patch(patch.Value.Put)
	}

	if r.Value.Patch == nil {
		r.Value.Patch = patch.Value.Patch
	} else {
		r.Value.Patch.patch(patch.Value.Patch)
	}

	if r.Value.Delete == nil {
		r.Value.Delete = patch.Value.Delete
	} else {
		r.Value.Delete.patch(patch.Value.Delete)
	}

	if r.Value.Head == nil {
		r.Value.Head = patch.Value.Head
	} else {
		r.Value.Head.patch(patch.Value.Head)
	}

	if r.Value.Options == nil {
		r.Value.Options = patch.Value.Options
	} else {
		r.Value.Options.patch(patch.Value.Options)
	}

	if r.Value.Trace == nil {
		r.Value.Trace = patch.Value.Trace
	} else {
		r.Value.Trace.patch(patch.Value.Trace)
	}

	r.Value.Parameters.Patch(patch.Value.Parameters)
}

func (op *Operation) patch(patch *Operation) {
	if len(patch.Summary) > 0 {
		op.Summary = patch.Summary
	}
	if len(patch.Description) > 0 {
		op.Description = patch.Description
	}
	if len(patch.OperationId) > 0 {
		op.OperationId = patch.OperationId
	}
	op.Deprecated = patch.Deprecated

	if op.RequestBody == nil {
		op.RequestBody = patch.RequestBody
	} else {
		op.RequestBody.patch(patch.RequestBody)
	}

	if op.Responses == nil {
		op.Responses = patch.Responses
	} else {
		for it := patch.Responses.Iter(); it.Next(); {
			r := it.Value().(*ResponseRef)
			if r.Value == nil {
				continue
			}
			statusCode := it.Key().(int)
			if v := op.Responses.GetResponse(statusCode); v != nil {
				v.patch(r.Value)
			} else {
				op.Responses.Set(statusCode, r)
			}
		}
	}
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
