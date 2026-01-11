package cli

import (
	"fmt"
)

type stringFlag struct {
	value  string
	isSet  bool
	source Source
}

func (f *stringFlag) Set(values []string, source Source) error {
	if f.source > source {
		return nil
	}

	if len(values) != 1 {
		return fmt.Errorf("expected 1 value, got %d", len(values))
	}

	f.value = values[0]
	f.isSet = true
	f.source = source
	return nil
}

func (f *stringFlag) Value() any {
	return f.value
}

func (f *stringFlag) String() string {
	return f.value
}

func (f *stringFlag) IsSet() bool { return f.isSet }

func (fs *FlagSet) String(name string, defaultValue string, doc FlagDoc) *FlagBuilder {
	return fs.StringShort(name, "", defaultValue, doc)
}

func (fs *FlagSet) StringShort(name string, short string, defaultValue string, doc FlagDoc) *FlagBuilder {
	v := &stringFlag{value: defaultValue}
	f := &Flag{Name: name, Shorthand: short, Value: v, DefaultValue: defaultValue, FlagDoc: doc, Type: "string"}
	fs.setFlag(f)
	return &FlagBuilder{flag: f}
}

func (fs *FlagSet) GetString(name string) string {
	v, ok := fs.GetValue(name)
	if !ok {
		panic(FlagNotFound{Name: name})
	}
	s, ok := v.(string)
	if !ok {
		panic(fmt.Sprintf("flag '%s' is not a string", name))
	}
	return s
}
