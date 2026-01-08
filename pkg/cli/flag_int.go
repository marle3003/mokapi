package cli

import (
	"errors"
	"fmt"
	"strconv"
)

type intFlag struct {
	value int
	isSet bool
}

func (f *intFlag) Set(values []string) error {
	if len(values) != 1 {
		return fmt.Errorf("expected 1 value, got %d", len(values))
	}
	if values[0] == "" {
		return nil
	}

	s := values[0]
	v, err := strconv.Atoi(s)
	if err != nil {
		err = errors.Unwrap(err)
		return fmt.Errorf("parsing %s: %w", s, err)
	}
	f.value = v
	f.isSet = true
	return nil
}

func (f *intFlag) Value() any {
	return f.value
}

func (f *intFlag) IsSet() bool {
	return f.isSet
}

func (f *intFlag) Type() string {
	return "int"
}

func (f *intFlag) String() string {
	return fmt.Sprintf("%d", f.value)
}

func (fs *FlagSet) Int(name string, defaultValue int, usage string) *FlagBuilder {
	return fs.IntShort(name, "", defaultValue, usage)
}

func (fs *FlagSet) IntShort(name string, short string, defaultValue int, usage string) *FlagBuilder {
	v := &intFlag{value: defaultValue}
	f := &Flag{Name: name, Shorthand: short, Value: v, Usage: usage, DefaultValue: defaultValue}
	fs.setFlag(f)
	return &FlagBuilder{flag: f}
}

func (fs *FlagSet) GetInt(name string) int {
	v, ok := fs.GetValue(name)
	if !ok {
		panic(FlagNotFound{Name: name})
	}
	b, ok := v.(int)
	if !ok {
		panic(fmt.Sprintf("flag '%s' is not a int", name))
	}
	return b
}
