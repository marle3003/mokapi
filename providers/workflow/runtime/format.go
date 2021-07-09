package runtime

import (
	"mokapi/providers/utils"
	"regexp"
	"strings"
)

func parse(s string, ctx *WorkflowContext) (interface{}, error) {
	s = strings.TrimRight(s, "\n")
	if strings.HasPrefix(s, "${{") && strings.HasSuffix(s, "}}") {
		return RunExpression(s[3:len(s)-2], ctx)
	}

	p := regexp.MustCompile(`\${{(?P<exp>[^}]*)}}`)
	matches := p.FindAllStringSubmatch(s, -1)
	for _, m := range matches {
		i, err := RunExpression(m[1], ctx)
		if err != nil {
			return s, err
		}
		v := utils.ToString(i)
		s = strings.Replace(s, m[0], v, 1)
	}

	return s, nil
}

func sPrint(s string, ctx *WorkflowContext) (string, error) {
	i, err := parse(s, ctx)
	if err != nil {
		return "", err
	}
	return utils.ToString(i), nil
}
