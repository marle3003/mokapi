package openapi

type Info struct {
	// The title of the service
	Name string `yaml:"title" json:"title"`

	Summary string `yaml:"summary,omitempty" json:"summary,omitempty"`

	// A short description of the API. CommonMark syntax MAY be
	// used for rich text representation.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	Contact *Contact `yaml:"contact,omitempty" json:"contact,omitempty"`

	// The version of the service
	Version string `yaml:"version" json:"version"`

	TermsOfService string `yaml:"termsOfService,omitempty" json:"termsOfService,omitempty"`

	License *License `yaml:"license,omitempty" json:"license,omitempty"`
}

type License struct {
	Name string `yaml:"name" json:"name"`
	Url  string `yaml:"url" json:"url"`
}

func (l *License) patch(patch *License) {
	if patch == nil {
		return
	}
	if len(patch.Name) > 0 {
		l.Name = patch.Name
	}
	if len(patch.Url) > 0 {
		l.Url = patch.Url
	}
}

func (c *Info) patch(patch Info) {
	if len(patch.Name) > 0 {
		c.Name = patch.Name
	}
	if len(patch.Summary) > 0 {
		c.Summary = patch.Summary
	}
	if len(patch.Description) > 0 {
		c.Description = patch.Description
	}
	if c.Contact == nil {
		c.Contact = clonePointer(patch.Contact)
	} else {
		c.Contact.patch(patch.Contact)
	}
	if len(patch.Version) > 0 {
		c.Version = patch.Version
	}
	if len(patch.TermsOfService) > 0 {
		c.TermsOfService = patch.TermsOfService
	}
	if c.License == nil {
		c.License = clonePointer(patch.License)
	} else {
		c.License.patch(patch.License)
	}
}
