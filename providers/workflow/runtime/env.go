package runtime

import "fmt"

type Env struct {
	parent *Env
	env    map[string]string
}

func (e *Env) Resolve(name string) (interface{}, error) {
	if val, ok := e.env[name]; ok {
		return val, nil
	}
	if e.parent != nil {
		return e.parent.Resolve(name)
	}
	return nil, fmt.Errorf("unknown env variable '%q'", name)
}

func (e *Env) get(name string) string {
	if val, ok := e.env[name]; ok {
		return val
	}
	if e.parent != nil {
		return e.parent.get(name)
	}
	return ""
}

func (e *Env) set(name, value string) {
	e.env[name] = value
}

func (e *Env) environ() []string {
	r := make([]string, 0)
	if e.parent != nil {
		r = append(r, e.parent.environ()...)
	}

	for k, v := range e.env {
		r = append(r, fmt.Sprintf("%v=%v", k, v))
	}
	return r
}
