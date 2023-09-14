package openapi

type Info struct {
	// The title of the service
	Name string `yaml:"title" json:"title"`

	// A short description of the API. CommonMark syntax MAY be
	// used for rich text representation.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	Contact *Contact `yaml:"contact,omitempty" json:"contact,omitempty"`

	// The version of the service
	Version string `yaml:"version" json:"version"`
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
