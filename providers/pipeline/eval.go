package pipeline

//
//import (
//	"fmt"
//	"github.com/pkg/errors"
//	log "github.com/sirupsen/logrus"
//	"mokapi/providers/pipeline/types"
//	"reflect"
//	"regexp"
//	"strings"
//)
//
//type evaluatable interface {
//	eval(*context) (types.Object, error)
//}
//
//func (p *pipeline) eval(ctx *context) (types.Object, error) {
//	var result types.Object = nil
//	var err error = nil
//	for _, s := range p.Stages {
//		result, err = s.eval(ctx)
//	}
//	return result, err
//}
//
//func (s *stage) eval(ctx *context) (types.Object, error) {
//	if s.When != nil {
//		obj, err := s.When.eval(ctx)
//		if err != nil {
//			return nil, errors.Wrapf(err, "stage '%v': when", s.DisplayName())
//		}
//		if b, ok := obj.(*types.Bool); ok {
//			if !b.Value() {
//				log.Debugf("skipping stage '%v'", s.DisplayName())
//				return nil, nil
//			}
//		} else {
//			return nil, errors.Errorf("expected bool but received value '%v'", obj.GetType())
//		}
//	}
//	result, err := s.Steps.eval(ctx)
//	if err != nil {
//		return nil, errors.Wrapf(err, "stage '%v'", s.DisplayName())
//	}
//
//	return result, nil
//}
//
//func (b *block) eval(ctx *context) (types.Object, error) {
//	var result types.Object
//	var err error
//	line := 0
//	for _, s := range b.Statements {
//		line++
//		result, err = s.eval(ctx)
//		if err != nil {
//			if ctx.outer != nil {
//				return nil, err
//			} else {
//				return nil, errors.Wrapf(err, "Line %v", line)
//			}
//		}
//	}
//	return result, nil
//}
//
//func (s *statement) eval(ctx *context) (types.Object, error) {
//	if s.Assignment != nil {
//		return s.Assignment.eval(ctx)
//	} else if s.Expression != nil {
//		return s.Expression.eval(ctx)
//	}
//
//	return nil, nil
//}
//
//func (a *assignment) eval(ctx *context) (types.Object, error) {
//	value, err := a.Expression.eval(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	return a.Variable.set(a.Operator, value, ctx)
//}
//
//func (v *variable) set(operator string, data interface{}, ctx *context) (types.Object, error) {
//	value, ok := data.(types.Object)
//	if data != nil && !ok {
//		return nil, fmt.Errorf("syntax error: invalid type %v", reflect.TypeOf(data))
//	}
//
//	if len(v.Identifier) > 0 {
//		_, varExists := ctx.getVar(v.Identifier)
//		if operator == ":=" && varExists {
//			return nil, fmt.Errorf("syntax error: identifier %v has already been declared", v.Identifier)
//		} else if operator == "=" && !varExists {
//			return nil, fmt.Errorf("syntax error: undefined identifier '%v'", v.Identifier)
//		}
//		ctx.setVar(v.Identifier, value)
//		return nil, nil
//	} else if len(v.Member) > 0 {
//		if operator == ":=" {
//			return nil, fmt.Errorf("syntax error: unexpected declaration")
//		}
//		paths := strings.Split(v.Member, ".")
//		obj, _ := ctx.getVar(paths[0])
//		err := types.SetField(obj, paths[1:], value)
//		return nil, err
//	}
//
//	return nil, fmt.Errorf("syntax error")
//}
//
//func (e *expression) eval(ctx *context) (types.Object, error) {
//	if e.OrCondition != nil {
//		return e.OrCondition.eval(ctx)
//	}
//
//	return nil, nil
//}
//
//func (e *orCondition) eval(ctx *context) (types.Object, error) {
//	return process("||", e.Left, e.Right, ctx)
//}
//
//func (e *andCondition) eval(ctx *context) (types.Object, error) {
//	return process("&&", e.Left, e.Right, ctx)
//}
//
//func (e *equality) eval(ctx *context) (types.Object, error) {
//	return process(e.Operator, e.Left, e.Right, ctx)
//}
//
//func (e *relational) eval(ctx *context) (types.Object, error) {
//	return process(e.Operator, e.Left, e.Right, ctx)
//}
//
//func (e *additive) eval(ctx *context) (types.Object, error) {
//	return process(e.Operator, e.Left, e.Right, ctx)
//}
//
//func (e *multiplicative) eval(ctx *context) (types.Object, error) {
//	return process(e.Operator, e.Left, e.Right, ctx)
//}
//
//func (e *unary) eval(ctx *context) (types.Object, error) {
//	switch e.Operator {
//	case "!":
//		right := &literal{Bool: NewBoolean(true)}
//		return process("!=", e.Primary, right, ctx)
//	default:
//		return e.Primary.eval(ctx)
//	}
//}
//
//func (p *primary) eval(ctx *context) (types.Object, error) {
//	if p.MemberAccess != nil {
//		return p.MemberAccess.eval(ctx)
//	} else if p.Step != nil {
//		return p.Step.eval(ctx)
//	} else if p.Literal != nil {
//		return p.Literal.eval(ctx)
//	} else if p.Closure != nil {
//		return p.Closure.eval(ctx)
//	} else if p.Expression != nil {
//		return p.Expression.eval(ctx)
//	}
//
//	return nil, nil
//}
//
//func (l *literal) eval(ctx *context) (types.Object, error) {
//	if l.Number != nil {
//		return types.NewNumber(*l.Number), nil
//	} else if l.String != nil {
//		s := *l.String
//
//		if s[0] == '"' {
//			pat := regexp.MustCompile(`[^\\]((\${(?P<exp>.*)})|(\$(?P<var>[^{^\s]*)))`)
//			matches := pat.FindAllStringSubmatch(s, -1) // matches is [][]string
//			groupNames := pat.SubexpNames()
//			for _, match := range matches {
//				for groupIndex, group := range match {
//					if len(group) == 0 {
//						continue
//					}
//					name := groupNames[groupIndex]
//					if name == "exp" {
//						exp, err := getExpr(group)
//						if err != nil {
//							return nil, err
//						}
//						v, err := exp.eval(ctx)
//						if err != nil {
//							return nil, err
//						}
//						s = strings.ReplaceAll(s, match[1], v.String())
//					} else if name == "var" {
//						v, err := accessMember(group, make([]types.Object, 0), ctx)
//						if err != nil {
//							return nil, err
//						}
//						s = strings.ReplaceAll(s, match[1], v.String())
//					}
//				}
//			}
//		}
//
//		return types.NewString(s[1 : len(s)-1]), nil
//	} else if l.Bool != nil {
//		return types.NewBool(bool(*l.Bool)), nil
//	}
//
//	return nil, fmt.Errorf("not implemented literal")
//}
//
//func (m *memberAccess) eval(ctx *context) (types.Object, error) {
//	args := make([]types.Object, len(m.Args))
//	for i, arg := range m.Args {
//		v, err := arg.Value.eval(ctx)
//		if err != nil {
//			return nil, err
//		}
//		args[i] = v
//	}
//
//	return accessMember(m.Name, args, ctx)
//}
//
//func accessMember(name string, args []types.Object, ctx *context) (types.Object, error) {
//	segments := strings.SplitN(name, ".", 2)
//	obj, ok := ctx.getVar(segments[0])
//	if !ok {
//		return nil, fmt.Errorf("undefined identifier '%v'", segments[0])
//	}
//
//	path := types.ParsePath(segments[1])
//	return path.Resolve(obj, args)
//}
//
//func (s *step) eval(ctx *context) (types.Object, error) {
//	if v, ok := ctx.getVar(s.Name); ok {
//		if len(s.Args) > 0 {
//			return nil, errors.Errorf("closure variable not supported")
//		}
//		return v, nil
//	} else if step, ok := ctx.getStep(s.Name); ok {
//		execution := step.Start()
//
//		args := map[string]types.Object{}
//		for i, arg := range s.Args {
//			value, err := arg.Value.eval(ctx)
//			if err != nil {
//				return nil, err
//			}
//			argName := arg.Name
//			if len(argName) == 0 {
//				argName = fmt.Sprintf("%v", i)
//			}
//			args[argName] = value
//		}
//
//		result, err := eval(execution, args, ctx)
//		if err != nil {
//			return nil, errors.Wrapf(err, "error in step %v", s.Name)
//		}
//		return result, nil
//	}
//
//	return nil, fmt.Errorf("unknown identifier %v", s.Name)
//}
//
//func (v *argumentValue) eval(ctx *context) (types.Object, error) {
//	return v.Expression.eval(ctx)
//}
//
//func (c *closure) eval(ctx *context) (types.Object, error) {
//	f := func(args []types.Object) (types.Object, error) {
//		parameters := map[string]types.Object{}
//		for i, a := range c.Args {
//			if i > len(args)-1 {
//				return nil, fmt.Errorf("index out of range of arguments")
//			}
//			parameters[a] = args[i]
//		}
//		return c.Block.eval(newContext(withVars(parameters), withOuter(ctx)))
//	}
//
//	return types.NewClosure(f), nil
//}
//
//func process(operator string, leftTerm evaluatable, rightTerm evaluatable, ctx *context) (types.Object, error) {
//	if leftTerm == nil {
//		return nil, fmt.Errorf("syntax error: missing operand")
//	}
//	left, err := leftTerm.eval(ctx)
//	if err != nil {
//		return nil, err
//	} else if reflect.ValueOf(rightTerm).IsNil() {
//		return left, nil
//	}
//
//	right, err := rightTerm.eval(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	switch operator {
//	//case "==":
//	//	return types.NewBool(left.Equals(right)), nil
//	//case "!=":
//	//	return types.NewBool(!left.Equals(right)), nil
//	default:
//		if value, ok := left.(types.ValueType); ok {
//			return value.Operator(types.Operator(operator), right)
//		}
//	}
//
//	return nil, errors.Errorf("invalid operation '%v' on type %v", operator, left.GetType())
//}
//
//type tagOptions map[string]string
//
//func eval(step StepExecution, args map[string]types.Object, ctx StepContext) (types.Object, error) {
//	v := reflect.ValueOf(step).Elem()
//	t := v.Type()
//	for i := 0; i < t.NumField(); i++ {
//		f := t.Field(i)
//		tag := getTag(f)
//		if tag == "-" {
//			continue
//		}
//		alias, options := parseOption(tag)
//		if len(alias) == 0 {
//			alias = strings.ToLower(string(f.Name[0])) + f.Name[1:]
//		}
//
//		var value types.Object
//		if arg, ok := args[alias]; ok {
//			value = arg
//		} else {
//			position, positionExists := options["position"]
//			if arg, ok := args[position]; positionExists && ok {
//				value = arg
//			} else if _, required := options["required"]; required {
//				return nil, fmt.Errorf("missing required argument '%v'", alias)
//			} else {
//				continue
//			}
//		}
//
//		field := v.FieldByName(f.Name)
//		fieldValue, err := types.ConvertFrom(value, field.Type())
//		if err != nil {
//			return nil, err
//		}
//		field.Set(fieldValue)
//	}
//
//	result, err := step.Run(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	return types.Convert(result)
//}
//
//func parseOption(tag string) (string, tagOptions) {
//	s := strings.Split(tag, ",")
//	options := tagOptions{}
//	for _, opt := range s[1:] {
//		kv := strings.Split(opt, "=")
//		if len(kv) > 1 {
//			options[kv[0]] = kv[1]
//		} else {
//			options[kv[0]] = ""
//		}
//	}
//	return s[0], options
//}
//
//func getTag(f reflect.StructField) string {
//	tag := f.Tag.Get("step")
//	if len(tag) > 0 {
//		return tag
//	}
//	tag = string(f.Tag)
//	if strings.Index(tag, ":") < 0 {
//		return tag
//	}
//	return ""
//}
