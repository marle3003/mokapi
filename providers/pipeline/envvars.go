package pipeline

import (
	"fmt"
	"os"
	"strings"
)

type EnvVarsModifier func(EnvVars)

type EnvVars map[string]string

func NewEnvVars(modifiers ...EnvVarsModifier) EnvVars {
	e := map[string]string{}
	for _, m := range modifiers {
		m(e)
	}
	return e
}

func With(envVars map[string]string) EnvVarsModifier {
	return func(original EnvVars) {
		for k, v := range envVars {
			original[k] = v
		}
	}
}

func fromOS() EnvVarsModifier {
	return func(original EnvVars) {
		for _, v := range os.Environ() {
			kv := strings.SplitN(v, "=", 2)
			original[kv[0]] = kv[1]
		}
	}
}

func (e EnvVars) String() string {
	sb := strings.Builder{}
	sb.WriteString("{")
	counter := 0
	for k, v := range e {
		if counter > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%v=%v", k, v))
		counter++
	}
	sb.WriteString("}")
	return sb.String()
}
