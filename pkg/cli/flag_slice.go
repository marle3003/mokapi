package cli

import (
	"fmt"
	"strings"
)

type stringSliceFlag struct {
	value   []string
	explode bool
	isSet   bool
}

func (f *stringSliceFlag) Set(values []string) error {
	if len(values) == 1 {
		if !f.explode {
			values = splitArrayItems(values[0])
		}
		if len(values) == 1 {
			f.value = append(f.value, values[0])
		} else {
			f.value = values
		}
	} else {
		f.value = values
	}
	f.isSet = true
	return nil
}

func (f *stringSliceFlag) Value() any {
	return f.value
}

func (f *stringSliceFlag) IsSet() bool {
	return f.isSet
}

func (f *stringSliceFlag) Type() string {
	return "list"
}

func (f *stringSliceFlag) String() string {
	return strings.Join(f.value, ",")
}

func (fs *FlagSet) StringSlice(name string, defaultValue []string, usage string, explode bool) *FlagBuilder {
	return fs.StringSliceShort(name, "", defaultValue, usage, explode)
}

func (fs *FlagSet) StringSliceShort(name string, short string, defaultValue []string, usage string, explode bool) *FlagBuilder {
	v := &stringSliceFlag{value: defaultValue, explode: explode}
	f := &Flag{Name: name, Shorthand: short, Value: v, Usage: usage, DefaultValue: defaultValue}
	fs.setFlag(f)
	return &FlagBuilder{flag: f}
}

func (fs *FlagSet) GetStringSlice(name string) []string {
	v, ok := fs.GetValue(name)
	if !ok {
		panic(FlagNotFound{Name: name})
	}
	s, ok := v.([]string)
	if !ok {
		panic(fmt.Sprintf("flag '%s' is not a string slice", name))
	}
	return s
}
