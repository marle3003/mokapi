package runtime

import "mokapi/providers/pipeline/lang/types"

type stack struct {
	values []types.Object
}

func newStack() *stack {
	return &stack{}
}

func (s *stack) Pop() (val types.Object) {
	n := len(s.values)
	if n == 0 {
		return
	}
	val = s.values[n-1]
	s.values = s.values[:n-1]
	return
}

func (s *stack) Push(val types.Object) {
	s.values = append(s.values, val)
}

func (s *stack) Reset() {
	s.values = nil
}
