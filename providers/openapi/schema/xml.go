package schema

type Xml struct {
	Wrapped   *bool  `yaml:"wrapped,omitempty" json:"wrapped,omitempty"`
	Name      string `yaml:"name" json:"name"`
	Attribute *bool  `yaml:"attribute,omitempty" json:"attribute,omitempty"`
	Prefix    string `yaml:"prefix" json:"prefix"`
	Namespace string `yaml:"namespace" json:"namespace"`
}

func (x *Xml) IsWrapped() bool {
	return isTrue(x.Wrapped)
}

func (x *Xml) IsAttribute() bool {
	return isTrue(x.Attribute)
}

// isTrue reads an optional flag: a flag that is not set is false.
func isTrue(b *bool) bool {
	return b != nil && *b
}
