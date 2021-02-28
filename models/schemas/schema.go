package schemas

type Schema struct {
	Name                 string
	Type                 string
	Format               string
	Description          string
	Properties           map[string]*Schema
	Faker                string
	Items                *Schema
	Xml                  *XmlEncoding
	AdditionalProperties *Schema
	Reference            string
	Required             []string
	IsResolved           bool
	Nullable             bool
}

func (s *Schema) IsPropertyRequired(name string) bool {
	if s.Required == nil {
		return false
	}
	for _, p := range s.Required {
		if p == name {
			return true
		}
	}
	return false
}

type XmlEncoding struct {
	Wrapped   bool
	Name      string
	Attribute bool
	Prefix    string
	Namespace string
	CData     bool
}
