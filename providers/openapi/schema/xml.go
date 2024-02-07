package schema

type Xml struct {
	Wrapped   bool   `yaml:"wrapped" json:"wrapped"`
	Name      string `yaml:"name" json:"name"`
	Attribute bool   `yaml:"attribute" json:"attribute"`
	Prefix    string `yaml:"prefix" json:"prefix"`
	Namespace string `yaml:"namespace" json:"namespace"`
}
