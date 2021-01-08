package basics

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"mokapi/providers/pipeline/lang/types"
	"mokapi/providers/pipeline/runtime"
	"path/filepath"
)

type YamlStep struct {
	types.AbstractStep
}

type YamlExecution struct {
	File string `step:"file,required"`
}

func (e *YamlStep) Start() types.StepExecution {
	return &YamlExecution{}
}

func (e *YamlExecution) Run(ctx types.StepContext) (interface{}, error) {
	env, ok := ctx.Get(runtime.EnvVarsType).(runtime.EnvVars)
	if !ok {
		return nil, errors.Errorf("env not defined")
	}

	dir, _ := env["WORKING_DIRECTORY"]

	file := filepath.Join(dir, fmt.Sprintf("%v", e.File))

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(data, &m)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing YAML file %s", e.File)
	}

	return convertObject(m), nil
}

func convertObject(o interface{}) types.Object {
	if m, ok := o.(map[interface{}]interface{}); ok {
		obj := types.NewExpando()
		for k, v := range m {
			propertyName := fmt.Sprint(k)
			v := convertObject(v)
			obj.Set(propertyName, v)
		}
		return obj
	}
	if a, ok := o.([]interface{}); ok {
		array := types.NewArray()
		for _, e := range a {
			array.Add(convertObject(e))
		}
		return array
	} else {
		if s, ok := o.(string); ok {
			return types.NewString(s)
		} else if f, ok := o.(float64); ok {
			return types.NewNumber(f)
		} else if i, ok := o.(int); ok {
			return types.NewNumber(float64(i))
		}
		return types.NewString(fmt.Sprintf("%v", o))
	}
}
