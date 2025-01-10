package dynamic

import "fmt"

type Scope struct {
	Name    string
	lexical map[string]interface{}
	dynamic map[string]interface{}

	scopes []*Scope
}

func NewScope(name string) *Scope {
	return &Scope{Name: name}
}

func (s *Scope) GetLexical(name string) (interface{}, error) {
	if s == nil {
		return nil, fmt.Errorf("name '%s' not found: no scope present", name)
	}

	current := s
	if len(s.scopes) > 0 {
		// get top
		current = s.scopes[len(s.scopes)-1]
	}

	if current.lexical != nil {
		if v, ok := current.lexical[name]; ok {
			return v, nil
		}
	}

	return nil, fmt.Errorf("name '%s' not found in scope '%s'", name, current.Name)
}

func (s *Scope) SetLexical(name string, value interface{}) error {
	if s == nil {
		return fmt.Errorf("set name '%s' failed: no scope present", name)
	}

	current := s
	if len(s.scopes) > 0 {
		current = s.scopes[len(s.scopes)-1]
	}

	if current.lexical == nil {
		current.lexical = map[string]interface{}{}
	}

	if _, ok := current.lexical[name]; ok {
		return fmt.Errorf("name '%s' already defined in scope '%s'", name, current.Name)
	}

	current.lexical[name] = value
	return nil
}

func (s *Scope) SetDynamic(name string, value interface{}) error {
	if s == nil {
		return fmt.Errorf("set name '%s' failed: no scope present", name)
	}

	_, err := s.GetDynamic(name)
	if err == nil {
		return nil
	}

	current := s
	if len(s.scopes) > 0 {
		current = s.scopes[len(s.scopes)-1]
	}

	if current.dynamic == nil {
		current.dynamic = map[string]interface{}{}
	}

	current.dynamic[name] = value
	return nil
}

func (s *Scope) GetDynamic(name string) (interface{}, error) {
	for i := len(s.scopes) - 1; i >= 0; i-- {
		current := s.scopes[i]
		if v, ok := current.dynamic[name]; ok {
			return v, nil
		}
	}

	if v, ok := s.dynamic[name]; ok {
		return v, nil
	}

	return nil, fmt.Errorf("name '%s' not found", name)
}

func (s *Scope) Open(name string) {
	s.scopes = append(s.scopes, &Scope{Name: name})
}

func (s *Scope) Close() {
	if len(s.scopes) > 0 {
		s.scopes = s.scopes[:len(s.scopes)-1]
	}
}
