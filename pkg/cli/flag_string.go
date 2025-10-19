package cli

import (
	"fmt"
)

type stringFlag struct {
	value string
}

func (f *stringFlag) Set(values []string) error {
	if len(values) != 1 {
		return fmt.Errorf("expected 1 value, got %d", len(values))
	}

	f.value = values[0]
	return nil
}

func (f *stringFlag) Value() any {
	return f.value
}

func (f *stringFlag) String() string {
	return f.value
}

func (fs *FlagSet) String(name string, defaultValue string, usage string) {
	fs.StringShort(name, "", defaultValue, usage)
}

func (fs *FlagSet) StringShort(name string, short string, defaultValue string, usage string) {
	v := &stringFlag{value: defaultValue}
	f := &Flag{Value: v, Usage: usage, DefaultValue: defaultValue}
	fs.setFlag(name, f)
	if short != "" {
		fs.setFlag(short, f)
	}
}

func (fs *FlagSet) GetString(name string) string {
	v, err := fs.GetValue(name)
	if err != nil {
		panic(err)
	}
	s, ok := v.(string)
	if !ok {
		panic(fmt.Sprintf("flag '%s' is not a string", name))
	}
	return s
}
