package pipeline

import (
	"fmt"
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/types"
	"reflect"
	"strings"
)

type evaluatable interface {
	eval(*context) (types.Object, error)
}

func (p *pipeline) eval(ctx *context) (types.Object, error) {
	return p.Block.eval(ctx)
}

func (b *block) eval(ctx *context) (types.Object, error) {
	var result types.Object
	var err error
	line := 0
	for _, s := range b.Statements {
		line++
		result, err = s.eval(ctx)
		if err != nil {
			if ctx.outer != nil {
				return nil, err
			} else {
				return nil, errors.Wrapf(err, "Line %v", line)
			}
		}
	}
	return result, nil
}

func (s *statement) eval(ctx *context) (types.Object, error) {
	if s.Assignment != nil {
		return s.Assignment.eval(ctx)
	} else if s.Expression != nil {
		return s.Expression.eval(ctx)
	}

	return nil, nil
}

func (a *assignment) eval(ctx *context) (types.Object, error) {
	value, err := a.Expression.eval(ctx)
	if err != nil {
		return nil, err
	}

	return a.Variable.set(a.Operator, value, ctx)
}

func (v *variable) set(operator string, data interface{}, ctx *context) (types.Object, error) {
	value, ok := data.(types.Object)
	if data != nil && !ok {
		return nil, fmt.Errorf("syntax error: invalid type %v", reflect.TypeOf(data))
	}

	if len(v.Identifier) > 0 {
		_, varExists := ctx.getVar(v.Identifier)
		if operator == ":=" && varExists {
			return nil, fmt.Errorf("syntax error: identifier %v has already been declared", v.Identifier)
		} else if operator == "=" && !varExists {
			return nil, fmt.Errorf("syntax error: undefined identifier '%v'", v.Identifier)
		}
		ctx.setVar(v.Identifier, value)
		return nil, nil
	} else if len(v.Member) > 0 {
		if operator == ":=" {
			return nil, fmt.Errorf("syntax error: unexpected declaration")
		}
		paths := strings.Split(v.Member, ".")
		obj, _ := ctx.getVar(paths[0])
		err := types.SetField(obj, paths[1:], value)
		return nil, err
	}

	return nil, fmt.Errorf("syntax error")
}

func (e *expression) eval(ctx *context) (types.Object, error) {
	if e.Equality != nil {
		return e.Equality.eval(ctx)
	}

	return nil, nil
}

func (e *equality) eval(ctx *context) (types.Object, error) {
	return process(e.Operator, e.Left, e.Right, ctx)
}

func (e *relational) eval(ctx *context) (types.Object, error) {
	return process(e.Operator, e.Left, e.Right, ctx)
}

func (e *additive) eval(ctx *context) (types.Object, error) {
	return process(e.Operator, e.Left, e.Right, ctx)
}

func (p *primary) eval(ctx *context) (types.Object, error) {
	if p.MemberAccess != nil {
		return p.MemberAccess.eval(ctx)
	} else if p.Step != nil {
		return p.Step.eval(ctx)
	} else if p.Literal != nil {
		return p.Literal.eval(ctx)
	}

	return nil, nil
}

func (l *literal) eval(ctx *context) (types.Object, error) {
	if l.Number != nil {
		return types.NewNumber(*l.Number), nil
	} else if l.String != nil {
		s := *l.String
		return types.NewString(s[1 : len(s)-1]), nil
	} else if l.Bool != nil {
		return types.NewBool(bool(*l.Bool)), nil
	}

	return nil, fmt.Errorf("not implemented literal")
}

func (m *memberAccess) eval(ctx *context) (types.Object, error) {
	path := strings.Split(m.Name, ".")
	obj, ok := ctx.getVar(path[0])
	if !ok {
		return nil, fmt.Errorf("undefined identifier '%v'", path[0])
	}

	args := make([]types.Object, len(m.Args))
	for i, arg := range m.Args {
		v, err := arg.Value.eval(ctx)
		if err != nil {
			return nil, err
		}
		args[i] = v
	}

	return types.InvokeMember(obj, path[1:], args)
}

func (s *step) eval(ctx *context) (types.Object, error) {
	if step, ok := ctx.getStep(s.Name); ok {
		execution := step.Start()

		args := map[string]types.Object{}
		for i, arg := range s.Args {
			value, err := arg.Value.eval(ctx)
			if err != nil {
				return nil, err
			}
			argName := arg.Name
			if len(argName) == 0 {
				argName = fmt.Sprintf("%v", i)
			}
			args[argName] = value
		}

		result, err := eval(execution, args, ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "error in step %v", s.Name)
		}
		return result, nil
	}

	return nil, fmt.Errorf("unknown step %v", s.Name)
}

func (o *argumentValue) eval(ctx *context) (types.Object, error) {
	if o.Literal != nil {
		return o.Literal.eval(ctx)
	} else if o.Member != nil {
		return o.Member.eval(ctx)
	} else if len(o.Identifier) > 0 {
		value, ok := ctx.getVar(o.Identifier)
		if !ok {
			return nil, fmt.Errorf("syntax error: undefined identifier '%v'", o.Identifier)
		}
		return value, nil
	} else if o.Closure != nil {
		return o.Closure.eval(ctx)
	}

	return nil, fmt.Errorf("error in operand")
}

func (c *closure) eval(ctx *context) (types.Object, error) {
	f := func(args []types.Object) (types.Object, error) {
		parameters := map[string]types.Object{}
		for i, a := range c.Args {
			if i > len(args)-1 {
				return nil, fmt.Errorf("index out of range of arguments")
			}
			parameters[a] = args[i]
		}
		return c.Block.eval(newContext(withVars(parameters), withOuter(ctx)))
	}

	return types.NewClosure(f), nil
}

func process(operator string, leftTerm evaluatable, rightTerm evaluatable, ctx *context) (types.Object, error) {
	if leftTerm == nil {
		return nil, fmt.Errorf("syntax error: missing operand")
	}
	left, err := leftTerm.eval(ctx)
	if err != nil {
		return nil, err
	} else if reflect.ValueOf(rightTerm).IsNil() {
		return left, nil
	}

	right, err := rightTerm.eval(ctx)
	if err != nil {
		return nil, err
	}

	switch operator {
	case "==":
		return types.NewBool(left.Equals(right)), nil
	case "!=":
		return types.NewBool(!left.Equals(right)), nil
	default:
		if value, ok := left.(types.ValueType); ok {
			return value.Operator(types.ArithmeticOperator(operator), right)
		}
	}

	return nil, errors.Errorf("invalid operation '%v' on type %v", left.GetType())
}

type tagOptions map[string]string

func eval(step StepExecution, args map[string]types.Object, ctx StepContext) (types.Object, error) {
	v := reflect.ValueOf(step).Elem()
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := getTag(f)
		if tag == "-" {
			continue
		}
		alias, options := parseOption(tag)
		if len(alias) == 0 {
			alias = strings.ToLower(string(f.Name[0])) + f.Name[1:]
		}

		var value types.Object
		if arg, ok := args[alias]; ok {
			value = arg
		} else {
			position, positionExists := options["position"]
			if arg, ok := args[position]; positionExists && ok {
				value = arg
			} else if _, required := options["required"]; required {
				return nil, fmt.Errorf("missing required argument '%v'", alias)
			} else {
				continue
			}
		}

		field := v.FieldByName(f.Name)
		fieldValue, err := types.ConvertFrom(value, field.Type())
		if err != nil {
			return nil, err
		}
		field.Set(fieldValue)
	}

	result, err := step.Run(ctx)
	if err != nil {
		return nil, err
	}

	return types.Convert(result)
}

func parseOption(tag string) (string, tagOptions) {
	s := strings.Split(tag, ",")
	options := tagOptions{}
	for _, opt := range s[1:] {
		kv := strings.Split(opt, "=")
		if len(kv) > 1 {
			options[kv[0]] = kv[1]
		} else {
			options[kv[0]] = ""
		}
	}
	return s[0], options
}

func getTag(f reflect.StructField) string {
	tag := f.Tag.Get("step")
	if len(tag) > 0 {
		return tag
	}
	tag = string(f.Tag)
	if strings.Index(tag, ":") < 0 {
		return tag
	}
	return ""
}
