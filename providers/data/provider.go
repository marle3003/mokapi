package data

type Provider interface {
	Provide(name string, schema *Schema) (interface{}, error)
	Close()
}

type Schema struct {
	Type                 string
	Format               string
	Description          string
	Properties           map[string]*Schema
	Faker                string
	Items                *Schema
	Xml                  *XmlEncoding
	AdditionalProperties string
	Reference            string
}

type XmlEncoding struct {
	Wrapped   bool
	Name      string
	Attribute bool
	Prefix    string
	Namespace string
	CData     bool
}
