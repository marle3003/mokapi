package runtime

import (
	log "github.com/sirupsen/logrus"
	"mokapi/providers/pipeline/lang"
	"mokapi/providers/pipeline/lang/types"
)

type stageVisitor struct {
	stack *stack
	outer visitor
	stage *lang.Stage
}

func (v *stageVisitor) Visit(node lang.Node) lang.Visitor {
	if node != nil && !v.outer.hasErrors() {
		switch node.(type) {
		case *lang.StepBlock:
			if v.stage.When != nil {
				val := v.stack.Pop()
				if b, ok := val.(*types.Bool); ok && !b.Val() {
					log.Infof("skipping stage '%v'", v.stage.Name)
					return nil
				}
			}
		}
		return v.outer.Visit(node)
	}

	return nil
}
