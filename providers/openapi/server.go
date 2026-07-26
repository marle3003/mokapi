package openapi

type Server struct {
	Url string `yaml:"url,omitempty" json:"url,omitempty"`

	// An optional string describing the host designated by the URL.
	// CommonMark syntax MAY be used for rich text representation.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	IsDefault bool `yaml:"-" json:"-"`
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
