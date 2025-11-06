package cli

import (
	"fmt"
	"strings"
)

type stringSliceFlag struct {
	value   []string
	explode bool
}

func (f *stringSliceFlag) Set(values []string) error {
	if !f.explode && len(values) == 1 {
		f.value = splitArrayItems(values[0])
	} else {
		f.value = values
	}
	return nil
}

func (f *stringSliceFlag) Value() any {
	return f.value
}

func (f *stringSliceFlag) String() string {
	return strings.Join(f.value, ",")
}

func (fs *FlagSet) StringSlice(name string, defaultValue []string, usage string, explode bool) {
	fs.StringSliceShort(name, "", defaultValue, usage, explode)
}

func (fs *FlagSet) StringSliceShort(name string, short string, defaultValue []string, usage string, explode bool) {
	v := &stringSliceFlag{value: defaultValue, explode: explode}
	f := &Flag{Name: name, Shorthand: short, Value: v, Usage: usage, DefaultValue: v.String()}
	fs.setFlag(f)
}

func (fs *FlagSet) GetStringSlice(name string) []string {
	v, err := fs.GetValue(name)
	if err != nil {
		panic(err)
	}
	s, ok := v.([]string)
	if !ok {
		panic(fmt.Sprintf("flag '%s' is not a string slice", name))
	}
	return s
}
