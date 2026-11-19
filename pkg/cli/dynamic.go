package cli

import (
	"fmt"
	"regexp"
)

var regexIndex = regexp.MustCompile(`\[<.*>]`)
var regexKey = regexp.MustCompile(`<.*>`)

func (fs *FlagSet) DynamicInt(name string, doc FlagDoc) *FlagBuilder {
	var d *DynamicFlag
	d = &DynamicFlag{
		Name:    name,
		Flag:    fs.stamped(Flag{FlagDoc: doc, Type: "int"}),
		pattern: convertToPattern(name),
		setValue: func(name string, value []string, source Source) error {
			f, ok := fs.flags[name]
			if !ok {
				v := &intFlag{}
				if err := v.Set(value, source); err != nil {
					return err
				}
				f = &Flag{Name: name, Value: v, Type: "int", FlagDoc: doc, Group: d.Group, Order: d.Order}
				fs.setFlag(f)
				return nil
			} else {
				return f.Value.Set(value, source)
			}
		},
	}
	fs.dynamic = append(fs.dynamic, d)
	return &FlagBuilder{flag: &d.Flag}
}

func (fs *FlagSet) DynamicString(name string, doc FlagDoc) *FlagBuilder {
	var d *DynamicFlag
	d = &DynamicFlag{
		Name:    name,
		Flag:    fs.stamped(Flag{FlagDoc: doc, Type: "string"}),
		pattern: convertToPattern(name),
		setValue: func(name string, value []string, source Source) error {
			f, ok := fs.flags[name]
			if !ok {
				v := &stringFlag{}
				if err := v.Set(value, source); err != nil {
					return err
				}
				f = &Flag{Name: name, Value: v, Type: "string", FlagDoc: doc, Group: d.Group, Order: d.Order}
				fs.setFlag(f)
				return nil
			} else {
				return f.Value.Set(value, source)
			}
		},
	}
	fs.dynamic = append(fs.dynamic, d)
	return &FlagBuilder{flag: &d.Flag}
}

func (fs *FlagSet) DynamicStringSlice(name string, explode bool, doc FlagDoc) *FlagBuilder {
	var d *DynamicFlag
	d = &DynamicFlag{
		Name:    name,
		Flag:    fs.stamped(Flag{FlagDoc: doc, Type: "list"}),
		pattern: convertToPattern(name),
		setValue: func(name string, value []string, source Source) error {
			f, ok := fs.flags[name]
			if !ok {
				v := &stringSliceFlag{explode: explode}
				if err := v.Set(value, source); err != nil {
					return err
				}
				f = &Flag{Name: name, Value: v, Type: "list", FlagDoc: doc, Group: d.Group, Order: d.Order}
				fs.setFlag(f)
				return nil
			} else {
				return f.Value.Set(value, source)
			}
		},
	}
	fs.dynamic = append(fs.dynamic, d)
	return &FlagBuilder{flag: &d.Flag}
}

func convertToPattern(s string) *regexp.Regexp {
	// index is either [0] or _0_. The latter is the old version.
	pattern := regexIndex.ReplaceAllString(s, "((\\[[0-9]+])|(-[0-9]+-?))")
	pattern = regexKey.ReplaceAllString(pattern, "[a-zA-Z]+")
	regex, err := regexp.Compile(fmt.Sprintf("^%s$", pattern))
	if err != nil {
		panic(fmt.Errorf("invalid regex pattern: %s", pattern))
	}
	return regex
}
