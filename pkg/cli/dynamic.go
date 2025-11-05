package cli

import (
	"fmt"
	"regexp"
)

var regexIndex = regexp.MustCompile(`\[<.*>]`)
var regexKey = regexp.MustCompile(`<.*>`)

func (fs *FlagSet) DynamicInt(name string, defaultValue int, usage string) {
	v := &intFlag{value: defaultValue}
	f := &DynamicFlag{
		Flag: Flag{
			Usage:        usage,
			Value:        v,
			DefaultValue: v.String(),
		},
		pattern: convertToPattern(name),
	}
	fs.dynamic = append(fs.dynamic, f)
}

func (fs *FlagSet) DynamicString(name string, defaultValue string, usage string) {
	v := &stringFlag{value: defaultValue}
	f := &DynamicFlag{
		Flag: Flag{
			Usage:        usage,
			Value:        v,
			DefaultValue: v.String(),
		},
		pattern: convertToPattern(name),
	}
	fs.dynamic = append(fs.dynamic, f)
}

func (fs *FlagSet) DynamicStringSlice(name string, defaultValue []string, usage string, explode bool) {
	v := &stringSliceFlag{value: defaultValue, explode: explode}
	f := &DynamicFlag{
		Flag: Flag{
			Usage:        usage,
			Value:        v,
			DefaultValue: v.String(),
		},
		pattern: convertToPattern(name),
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
