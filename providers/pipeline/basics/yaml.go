package basics

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"mokapi/providers/pipeline/lang/runtime"
	"mokapi/providers/pipeline/lang/types"
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

func convertObject(i interface{}) types.Object {
	switch o := i.(type) {
	case map[interface{}]interface{}:
		obj := types.NewExpando()
		for k, v := range o {
			propertyName := fmt.Sprint(k)
			v := convertObject(v)
			obj.SetField(propertyName, v)
		}
		return obj
	case map[string]interface{}:
		obj := types.NewExpando()
		for k, v := range o {
			v := convertObject(v)
			obj.SetField(k, v)
		}
		return obj
	case []interface{}:
		array := types.NewArray()
		for _, e := range o {
			array.Add(convertObject(e))
		}
		return array
	case string:
		return types.NewString(o)
	case float64:
		return types.NewNumber(o)
	case int:
		return types.NewNumber(float64(o))
	}
	return types.NewString(fmt.Sprintf("%v", i))
}
