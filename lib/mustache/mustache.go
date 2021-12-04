package mustache

import (
	"fmt"
	"mokapi/objectpath"
	"regexp"
	"strings"
)

var pattern = regexp.MustCompile(`{{\s*([\w\.]+)\s*}}`)

func Render(template string, data interface{}) (string, error) {
	matches := pattern.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		path := strings.TrimSpace(match[1])
		var value interface{}
		var err error
		if path == "." {
			value = data
		} else {
			value, err = objectpath.Resolve(path, data)
			if err != nil {
				return "", err
			}
		}

		template = strings.ReplaceAll(template, match[0], fmt.Sprintf("%v", value))
	}

	return template, nil
}
