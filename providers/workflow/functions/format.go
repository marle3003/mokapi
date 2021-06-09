package functions

import (
	"fmt"
	"mokapi/providers/utils"
	"regexp"
	"strconv"
	"strings"
)

func Format(args ...interface{}) (interface{}, error) {
	p := regexp.MustCompile(`{(\d+)}`)
	s := utils.ToString(args[0])
	matches := p.FindAllStringSubmatch(s, -1)
	for _, m := range matches {
		i, err := strconv.Atoi(m[1])
		if err != nil {
			return "", err
		}
		if i+1 >= len(args) {
			return "", fmt.Errorf("index out of range")
		}
		v := utils.ToString(args[i+1])
		s = strings.Replace(s, m[0], v, 1)
	}

	return s, nil
}
