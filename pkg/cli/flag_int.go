package cli

import (
	"fmt"
	"strconv"
)

type intFlag struct {
	value int
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
		return err
	}
	f.value = v
	return nil
}

func (f *intFlag) Value() any {
	return f.value
}

func (f *intFlag) String() string {
	return fmt.Sprintf("%d", f.value)
}

func (fs *FlagSet) Int(name string, defaultValue int, usage string) {
	fs.IntShort(name, "", defaultValue, usage)
}

func (fs *FlagSet) IntShort(name string, short string, defaultValue int, usage string) {
	v := &intFlag{value: defaultValue}
	f := &Flag{Name: name, Shorthand: short, Value: v, Usage: usage, DefaultValue: v.String()}
	fs.setFlag(f)
}

func (fs *FlagSet) GetInt(name string) int {
	v, err := fs.GetValue(name)
	if err != nil {
		panic(err)
	}
	b, ok := v.(int)
	if !ok {
		panic(fmt.Sprintf("flag '%s' is not a int", name))
	}
	return b
}
