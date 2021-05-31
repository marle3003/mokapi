package runtime

import "fmt"

type Env struct {
	parent *Env
	env    map[string]interface{}
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

func (e *Env) Get(name string) interface{} {
	if val, ok := e.env[name]; ok {
		return val
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil
}

func (e *Env) Set(name string, value interface{}) {
	e.env[name] = value
}

func (e *Env) envStrings() []string {
	r := make([]string, 0)
	if e.parent != nil {
		r = append(r, e.parent.envStrings()...)
	}

	for k, v := range e.env {
		r = append(r, fmt.Sprintf("%v=%v", k, v))
	}
	return r
}
