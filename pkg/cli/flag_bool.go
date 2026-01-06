package cli

import (
	"fmt"
)

type boolFlag struct {
	value bool
	isSet bool
}

func (b *boolFlag) Set(values []string) error {
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

func (fs *FlagSet) Bool(name string, defaultValue bool, usage string) {
	fs.BoolShort(name, "", defaultValue, usage)
}

func (fs *FlagSet) BoolShort(name string, short string, defaultValue bool, usage string) {
	v := &boolFlag{value: defaultValue}
	f := &Flag{Value: v, Name: name, Shorthand: short, Usage: usage, DefaultValue: defaultValue}
	fs.setFlag(f)
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
