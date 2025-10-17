package cli

import (
	"fmt"
)

type boolFlag struct {
	value bool
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
	return nil
}

func (b *boolFlag) Value() any {
	return b.value
}

func (b *boolFlag) String() string {
	return fmt.Sprintf("%v", b.value)
}

func (fs *FlagSet) Bool(name string, defaultValue bool, usage string) {
	fs.BoolShort(name, "", defaultValue, usage)
}

func (fs *FlagSet) BoolShort(name string, short string, defaultValue bool, usage string) {
	v := &boolFlag{value: defaultValue}
	f := &Flag{Value: &boolFlag{}, Usage: usage, DefaultValue: v.String()}
	fs.setFlag(name, f)
	if short != "" {
		fs.setFlag(short, f)
	}
}

func (fs *FlagSet) GetBool(name string) bool {
	v, err := fs.GetValue(name)
	if err != nil {
		panic(err)
	}
	b, ok := v.(bool)
	if !ok {
		panic(fmt.Sprintf("flag '%s' is not a bool", name))
	}
	return b
}
