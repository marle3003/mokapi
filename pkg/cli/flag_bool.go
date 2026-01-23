package cli

import (
	"fmt"
)

type boolFlag struct {
	value  bool
	isSet  bool
	source Source
}

func (b *boolFlag) Set(values []string, source Source) error {
	if b.source > source {
		return nil
	}

	if len(values) == 0 {
		b.value = true
	} else {
		s := values[0]
		if s == "" || s == "true" {
			b.value = true
		} else if s == "false" {
			b.value = false
		} else {
			return fmt.Errorf("flag '%s' is not a bool", s)
		}
	}
	b.isSet = true
	b.source = source
	return nil
}

func (b *boolFlag) Value() any {
	return b.value
}

func (b *boolFlag) IsSet() bool {
	return b.isSet
}

func (b *boolFlag) String() string {
	return fmt.Sprintf("%v", b.value)
}

func (fs *FlagSet) Bool(name string, defaultValue bool, doc FlagDoc) *FlagBuilder {
	return fs.BoolShort(name, "", defaultValue, doc)
}

func (fs *FlagSet) BoolShort(name string, short string, defaultValue bool, doc FlagDoc) *FlagBuilder {
	v := &boolFlag{value: defaultValue}
	f := &Flag{Value: v, Name: name, Shorthand: short, DefaultValue: defaultValue, FlagDoc: doc, Type: "bool"}
	fs.setFlag(f)
	return &FlagBuilder{flag: f}
}

func (fs *FlagSet) GetBool(name string) bool {
	v, ok := fs.GetValue(name)
	if !ok {
		panic(FlagNotFound{Name: name})
	}
	b, ok := v.(bool)
	if !ok {
		panic(fmt.Sprintf("flag '%s' is not a bool", name))
	}
	return b
}
