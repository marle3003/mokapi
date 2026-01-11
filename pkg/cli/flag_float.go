package cli

import (
	"fmt"
	"strconv"
)

type floatFlag struct {
	value  float64
	isSet  bool
	source Source
}

func (f *floatFlag) Set(values []string, source Source) error {
	if f.source > source {
		return nil
	}

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
	f.source = source
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

func (fs *FlagSet) Float(name string, defaultValue float64, doc FlagDoc) *FlagBuilder {
	return fs.FloatShort(name, "", defaultValue, doc)
}

func (fs *FlagSet) FloatShort(name string, short string, defaultValue float64, doc FlagDoc) *FlagBuilder {
	v := &floatFlag{value: defaultValue}
	f := &Flag{Name: name, Shorthand: short, Value: v, DefaultValue: defaultValue, FlagDoc: doc, Type: "float"}
	fs.setFlag(f)
	return &FlagBuilder{flag: f}
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
