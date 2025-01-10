package dynamic

import "fmt"

type Scope struct {
	name   string
	scopes []*ScopeItem
	parent *Scope
}

type ScopeItem struct {
	name    string
	lexical map[string]interface{}
	dynamic map[string]interface{}
}

func NewScope(name string) *Scope {
	s := &Scope{}
	s.Open(name)
	return s
}

func (s *Scope) SetParent(parent Scope) {
	s.parent = &parent
}

func (s *Scope) Name() string {
	current := s.top()
	if current == nil {
		return s.name
	}
	return current.name
}

func (s *Scope) SetName(name string) {
	current := s.top()
	if current == nil {
		s.name = name
	} else {
		current.name = name
	}
}

func (s *Scope) GetLexical(name string) (interface{}, error) {
	if s == nil {
		return nil, fmt.Errorf("name '%s' not found: no scope present", name)
	}

	current := s.top()

	if current.lexical != nil {
		if v, ok := current.lexical[name]; ok {
			return v, nil
		}
	}

	return nil, fmt.Errorf("name '%s' not found in scope '%s'", name, current.name)
}

func (s *Scope) SetLexical(name string, value interface{}) error {
	if s == nil {
		return fmt.Errorf("set name '%s' failed: no scope present", name)
	}

	current := s.top()
	if current == nil {
		return fmt.Errorf("set name '%s' failed: no scope present", name)
	}

	if current.lexical == nil {
		current.lexical = map[string]interface{}{}
	}

	if _, ok := current.lexical[name]; ok {
		return fmt.Errorf("name '%s' already defined in scope '%s'", name, current.name)
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

	current := s.top()
	if current == nil {
		return fmt.Errorf("set name '%s' failed: no scope present", name)
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

	if s.parent != nil {
		return s.parent.GetDynamic(name)
	} else {
		return nil, fmt.Errorf("name '%s' not found", name)
	}
}

func (s *Scope) Open(name string) {
	s.scopes = append(s.scopes, &ScopeItem{name: name})
}

func (s *Scope) Close() {
	// we don't close first scope to be able to get anchors on root level
	if len(s.scopes) > 1 {
		s.scopes = s.scopes[:len(s.scopes)-1]
	}
}

func (s *Scope) top() *ScopeItem {
	if len(s.scopes) > 0 {
		return s.scopes[len(s.scopes)-1]
	}
	return nil
}

func (s *Scope) OpenIfNeeded(name string) {
	if len(s.scopes) == 0 {
		s.Open(name)
	}
}
