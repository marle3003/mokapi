package cli

import (
	"fmt"
	"regexp"
)

var regexIndex = regexp.MustCompile(`\[<.*>]`)
var regexKey = regexp.MustCompile(`<.*>`)

func (fs *FlagSet) DynamicInt(name string, usage string) {
	f := &DynamicFlag{
		Usage:   usage,
		pattern: convertToPattern(name),
		setValue: func(name string, value []string) error {
			f, ok := fs.flags[name]
			if !ok {
				v := &intFlag{}
				if err := v.Set(value); err != nil {
					return err
				}
				f = &Flag{Name: name, Value: v}
				fs.setFlag(f)
				return nil
			} else {
				return f.Value.Set(value)
			}
		},
	}
	fs.dynamic = append(fs.dynamic, f)
}

func (fs *FlagSet) DynamicString(name string, usage string) {
	f := &DynamicFlag{
		Usage:   usage,
		pattern: convertToPattern(name),
		setValue: func(name string, value []string) error {
			f, ok := fs.flags[name]
			if !ok {
				v := &stringFlag{}
				if err := v.Set(value); err != nil {
					return err
				}
				f = &Flag{Name: name, Value: v}
				fs.setFlag(f)
				return nil
			} else {
				return f.Value.Set(value)
			}
		},
	}
	fs.dynamic = append(fs.dynamic, f)
}

func (fs *FlagSet) DynamicStringSlice(name string, usage string, explode bool) {
	f := &DynamicFlag{
		Usage:   usage,
		pattern: convertToPattern(name),
		setValue: func(name string, value []string) error {
			f, ok := fs.flags[name]
			if !ok {
				v := &stringSliceFlag{explode: explode}
				if err := v.Set(value); err != nil {
					return err
				}
				f = &Flag{Name: name, Value: v}
				fs.setFlag(f)
				return nil
			} else {
				return f.Value.Set(value)
			}
		},
	}
	fs.dynamic = append(fs.dynamic, f)
}

func convertToPattern(s string) *regexp.Regexp {
	pattern := regexIndex.ReplaceAllString(s, "\\[[0-9]+]")
	pattern = regexKey.ReplaceAllString(pattern, "[a-zA-Z]+")
	regex, err := regexp.Compile(fmt.Sprintf("^%s$", pattern))
	if err != nil {
		panic(fmt.Errorf("invalid regex pattern: %s", pattern))
	}
	return regex
}
