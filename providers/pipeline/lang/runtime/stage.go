package runtime

import (
	log "github.com/sirupsen/logrus"
	"mokapi/providers/pipeline/lang/ast"
	"mokapi/providers/pipeline/lang/types"
)

type stageVisitor struct {
	outer visitor
	stage *ast.Stage
}

func newStageVisitor(stage *ast.Stage, outer visitor) *stageVisitor {
	return &stageVisitor{stage: stage, outer: outer}
}

func (v *stageVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil && !v.outer.HasErrors() {
		switch node.(type) {
		case *ast.StepBlock:
			if v.stage.When != nil {
				val := v.outer.Stack().Pop()
				if b, ok := val.(*types.Bool); ok && !b.Val() {
					log.Infof("skipping stage '%v'", v.stage.Name)
					return nil
				}
			}
		}
		return v.outer.Visit(node)
	}

	v.outer.CloseScope()

	return nil
}
