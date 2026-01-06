package cli

import (
	"fmt"
	"strconv"
)

type floatFlag struct {
	value float64
	isSet bool
}

func (f *floatFlag) Set(values []string) error {
	if len(values) != 1 {
		return fmt.Errorf("expected 1 value, got %d", len(values))
	}

	s := values[0]
	v, err := strconv.ParseFloat(s, 10)
	if err != nil {
		return err
	}
	f.value = v
	f.isSet = true
	return nil
}

func (f *floatFlag) Value() any {
	return f.value
}

func (f *floatFlag) IsSet() bool {
	return f.isSet
}

func (f *floatFlag) String() string {
	return fmt.Sprintf("%f", f.value)
}

func (fs *FlagSet) Float(name string, defaultValue float64, usage string) {
	fs.FloatShort(name, "", defaultValue, usage)
}

func (fs *FlagSet) FloatShort(name string, short string, defaultValue float64, usage string) {
	v := &floatFlag{value: defaultValue}
	f := &Flag{Name: name, Shorthand: short, Value: v, Usage: usage, DefaultValue: defaultValue}
	fs.setFlag(f)
}

func (fs *FlagSet) GetFloat(name string) float64 {
	v, ok := fs.GetValue(name)
	if !ok {
		panic(FlagNotFound{Name: name})
	}
	b, ok := v.(float64)
	if !ok {
		panic(fmt.Sprintf("flag '%s' is not a int", name))
	}
	return b
}
