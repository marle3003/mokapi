package directory

import "slices"

func (c *Config) Patch(patch *Config) {
	if patch == nil {
		return
	}

	c.patchInfo(patch)

	if len(patch.Address) > 0 {
		c.Address = patch.Address
	}

	if patch.SizeLimit > 0 {
		c.SizeLimit = patch.SizeLimit
	}

	c.patchEntries(patch)
}

func (c *Config) patchInfo(patch *Config) {
	if len(patch.Info.Description) > 0 {
		c.Info.Description = patch.Info.Description
	}
	if len(patch.Info.Version) > 0 {
		c.Info.Version = patch.Info.Version
	}
}

func (e *Entry) patch(patch Entry) {
	if len(e.Dn) == 0 {
		e.Dn = patch.Dn
	}
	for k, p := range patch.Attributes {
		if v, ok := e.Attributes[k]; ok {
			v = append(v, p...)
			e.Attributes[k] = v
		} else {
			if e.Attributes == nil {
				e.Attributes = map[string][]string{}
			}
			e.Attributes[k] = slices.Clone(p)
		}
	}
}

func (c *Config) patchEntries(patch *Config) {
	if c.Entries == nil {
		c.Entries = NewEntryStore()
	}
	c.Entries.Merge(patch.Entries)
}
