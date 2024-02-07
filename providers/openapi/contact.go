package openapi

type Contact struct {
	Name  string `yaml:"name" json:"name"`
	Url   string `yaml:"url" json:"url"`
	Email string `yaml:"email" json:"email"`
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
