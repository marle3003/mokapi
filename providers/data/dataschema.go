package data

type DataSchema struct {
	Type        string
	Format      string
	Description string
	Properties  map[string]*DataSchema
	Faker       string `yaml:"x-faker"`
	Resource    string `yaml:"x-resource"`
	Items       *DataSchema
	XmlEncoding *XmlEncoding
}

type XmlEncoding struct {
	Wrapped   bool
	Name      string
	Attribute bool
	Prefix    string
	Namespace string
	CData     bool
}
