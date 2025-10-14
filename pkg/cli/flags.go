package cli

import (
	"fmt"
	"regexp"
)

type FlagSet struct {
	flags   map[string]*Flag
	dynamic []*DynamicFlag
}

type Flag struct {
	Usage        string
	Value        Value
	DefaultValue string
}

type DynamicFlag struct {
	Flag
	pattern *regexp.Regexp
}

type FlagNotFound struct {
	Name string
}

type Value interface {
	Set([]string) error
	Value() any
	String() string
}

func (fs *FlagSet) setFlag(name string, f *Flag) {
	if fs.flags == nil {
		fs.flags = make(map[string]*Flag)
	}
	fs.flags[name] = f
}

func (fs *FlagSet) setValue(name string, value []string) error {
	if fs.flags != nil {
		f, ok := fs.flags[name]
		if ok {
			return f.Value.Set(value)
		}
		for _, flag := range fs.dynamic {
			if flag.isValidFlag(name) {
				return flag.Value.Set(value)
			}
		}
	}
	return fmt.Errorf("unknown flag '%v'", name)
}

func (fs *FlagSet) GetValue(name string) (any, error) {
	f, ok := fs.flags[name]
	if !ok {
		return nil, &FlagNotFound{Name: name}
	}
	return f.Value.Value(), nil
}

func (d *DynamicFlag) isValidFlag(flag string) bool {
	return d.pattern.MatchString(flag)
}

func (e *FlagNotFound) Error() string {
	return fmt.Sprintf("flag '%s' not found", e.Name)
}
