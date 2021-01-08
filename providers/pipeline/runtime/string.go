package runtime

import (
	"mokapi/providers/pipeline/lang"
	"regexp"
	"strings"
)

func format(s string, scope *Scope) (string, error) {
	pat := regexp.MustCompile(`[^\\]((\${(?P<exp>.*)})|(\$(?P<var>[^{^\s]*)))`)
	matches := pat.FindAllStringSubmatch(s, -1) // matches is [][]string
	groupNames := pat.SubexpNames()
	for _, match := range matches {
		for groupIndex, group := range match {
			if len(group) == 0 {
				continue
			}
			name := groupNames[groupIndex]
			if name == "exp" || name == "var" {
				expr, err := lang.ParseExpr([]byte(group))
				if err != nil {
					return "", err
				}
				obj, err := runExpr(expr, scope)
				if err != nil {
					return "", err
				}
				r := "NULL"
				if obj != nil {
					r = obj.String()
				}
				s = strings.ReplaceAll(s, match[1], r)
			}
		}
	}
	return s, nil
}
