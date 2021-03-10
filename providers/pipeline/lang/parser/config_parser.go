package parser

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/config/dynamic/mokapi"
	"mokapi/providers/pipeline/lang/ast"
	"strings"
)

type configParser struct {
	errors []error
	scope  *ast.Scope
}

func ParseConfig(config *mokapi.Config, scope *ast.Scope) (f *ast.File, err error) {
	parser := &configParser{
		errors: make([]error, 0),
		scope:  scope,
	}
	f = parser.parse(config)
	err = parser.err()
	return
}

func (p *configParser) parse(config *mokapi.Config) (f *ast.File) {
	f = &ast.File{Scope: p.scope}
	for _, pipeline := range config.Pipelines {
		f.Pipelines = append(f.Pipelines, p.parsePipeline(pipeline))
	}
	return
}

func (p *configParser) parsePipeline(c mokapi.Pipeline) *ast.Pipeline {
	pipeline := &ast.Pipeline{}
	pipeline.Name = c.Name

	p.openScope()
	defer p.closeScope()
	pipeline.Scope = p.scope
	if c.Stages != nil {
		s, _ := p.parseStages(c.Stages)
		pipeline.Stages = s
	} else if c.Stage != nil {
		s := p.parseStage(c.Stage)
		pipeline.Stages = append(pipeline.Stages, s)
	} else {
		s := &ast.Stage{Scope: p.scope}
		s.Steps = p.parseSteps(c.Steps)
		pipeline.Stages = append(pipeline.Stages, s)
	}

	return pipeline
}

func (p *configParser) parseStages(s []*mokapi.Stage) (stages []*ast.Stage, vars *ast.VarsBlock) {
	for _, stage := range s {
		stages = append(stages, p.parseStage(stage))
	}

	return
}

func (p *configParser) parseStage(s *mokapi.Stage) *ast.Stage {
	src := fmt.Sprintf("stage('%v') { when { %v } steps { %v }", s.Name, s.Condition, s.Steps)
	parser := newParser([]byte(src), p.scope)
	return parser.parseStage()
}

func (p *configParser) parseSteps(steps string) *ast.StepBlock {
	parser := newParser([]byte(fmt.Sprintf("steps { %v }", steps)), p.scope)
	return parser.parseSteps()
}

func (p *configParser) openScope() {
	p.scope = ast.OpenScope(p.scope)
}

func (p *configParser) closeScope() {
	p.scope = p.scope.Outer
}

func (p configParser) err() error {
	if len(p.errors) == 0 {
		return nil
	}
	sb := strings.Builder{}
	for i, e := range p.errors {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(e.Error())
	}
	return errors.New(sb.String())
}
