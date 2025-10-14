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
	f := &Flag{Value: v, Usage: usage, DefaultValue: v.String()}
	fs.setFlag(name, f)
	if short != "" {
		fs.setFlag(short, f)
	}
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

type sliceFlag struct {
	value   []string
	explode bool
}

func (f *sliceFlag) Set(values []string) error {
	return nil
}

func (f *sliceFlag) Value() any {
	return f.value
}

func (f *sliceFlag) String() string {
	return ""
}

func (fs *FlagSet) SliceShort(name string, short string, defaultValue []string, usage string, explode bool) {
	v := &sliceFlag{value: defaultValue, explode: explode}
	f := &Flag{Value: v, Usage: usage, DefaultValue: ""}
	fs.setFlag(name, f)
	if short != "" {
		fs.setFlag(short, f)
	}
}
