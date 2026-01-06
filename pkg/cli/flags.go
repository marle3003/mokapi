package cli

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

type FlagSet struct {
	flags         map[string]*Flag
	dynamic       []*DynamicFlag
	orderedFlags  map[string]int
	setConfigFile func(string)
}

type Flag struct {
	Name         string
	Shorthand    string
	Usage        string
	Value        Value
	DefaultValue any
	Aliases      []string
}

type DynamicFlag struct {
	Usage    string
	pattern  *regexp.Regexp
	setValue func(name string, value []string) error
}

type FlagNotFound struct {
	Name string
}

type Value interface {
	Set([]string) error
	Value() any
	String() string
	IsSet() bool
	Type() string
}

func (fs *FlagSet) setFlag(f *Flag) {
	if fs.flags == nil {
		fs.flags = make(map[string]*Flag)
	}
	fs.flags[f.Name] = f
	if f.Shorthand != "" {
		fs.flags[f.Shorthand] = f
	}
	fs.orderedFlags[f.Name] = len(fs.flags)
}

func (fs *FlagSet) setValue(name string, value []string) error {
	// backwards compatibility
	name = strings.ReplaceAll(name, ".", "-")
	if fs.flags != nil {
		f, ok := fs.flags[name]
		if ok {
			err := f.Value.Set(value)
			if err != nil {
				return fmt.Errorf("failed to set flag %s: %w", name, err)
			}
			fs.orderedFlags[name] = len(fs.orderedFlags)
			return nil
		}
	}
	for _, flag := range fs.dynamic {
		if flag.isValidFlag(name) {
			err := flag.setValue(name, value)
			if err != nil {
				return fmt.Errorf("failed to set flag %s: %w", name, err)
			}
			fs.orderedFlags[name] = len(fs.orderedFlags)
			return nil
		}
	}
	return fmt.Errorf("unknown flag '%v'", name)
}

func (fs *FlagSet) IsValidFlag(name string) bool {
	if fs.flags == nil {
		return false
	}
	if _, ok := fs.flags[name]; ok {
		return true
	}
	for _, flag := range fs.dynamic {
		if flag.isValidFlag(name) {
			return true
		}
	}
	return false
}

func (fs *FlagSet) GetValue(name string) (any, bool) {
	f, ok := fs.flags[name]
	if !ok {
		return nil, false
	}
	return f.Value.Value(), true
}

func (fs *FlagSet) Alias(name, alias string) {
	f, ok := fs.flags[name]
	if !ok {
		panic(fmt.Sprintf("flag '%v' does not exist", name))
	}
	fs.flags[alias] = f
	f.Aliases = append(f.Aliases, alias)
}

func (fs *FlagSet) Visit(fn func(*Flag) error) error {
	flags := make([]*Flag, 0, len(fs.flags))
	for _, flag := range fs.flags {
		flags = append(flags, flag)
	}
	// 1. unused flags first, 2. flags in order of set
	slices.SortStableFunc(flags, func(f1, f2 *Flag) int {
		if f1.Value.IsSet() != f2.Value.IsSet() {
			if !f1.Value.IsSet() {
				return -1
			}
			return 1
		}
		if d := fs.orderedFlags[f1.Name] - fs.orderedFlags[f2.Name]; d != 0 {
			return d
		}
		return strings.Compare(f1.Name, f2.Name)
	})
	for _, flag := range flags {
		err := fn(flag)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DynamicFlag) isValidFlag(flag string) bool {
	return d.pattern.MatchString(flag)
}

func (e *FlagNotFound) Error() string {
	return fmt.Sprintf("flag '%s' not found", e.Name)
}
