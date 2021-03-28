package basics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"mokapi/providers/pipeline/lang/runtime"
	"mokapi/providers/pipeline/lang/types"
	"path/filepath"
	"strings"
	"text/template"
)

type ReadFileStep struct {
	types.AbstractStep
}

type ReadFileStepExecution struct {
	File       string `step:"file,position=0,required"`
	AsString   bool
	AsTemplate bool
	AsJson     bool
	AsYml      bool
}

func (e *ReadFileStep) Start() types.StepExecution {
	return &ReadFileStepExecution{AsString: true}
}

func (e *ReadFileStepExecution) Run(ctx types.StepContext) (interface{}, error) {
	env, ok := ctx.Get(runtime.EnvVarsType).(runtime.EnvVars)
	if !ok {
		return nil, errors.Errorf("env not defined")
	}

	dir, _ := env["WORKING_DIRECTORY"]

	file := filepath.Join(dir, fmt.Sprintf("%v", e.File))

	b, err := readFile(file, e.AsTemplate)
	if err != nil {
		return nil, err
	}

	if e.AsJson {
		return toJson(b)
	} else if e.AsYml {
		return toYml(b)
	} else if e.AsString {
		return string(b), nil
	}

	return b, nil
}

func readFile(p string, asTemplate bool) ([]byte, error) {
	data, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}

	if !asTemplate {
		return data, nil
	}

	content := string(data)

	funcMap := sprig.TxtFuncMap()
	funcMap["extractUsername"] = extractUsername
	tmpl := template.New(p).Funcs(funcMap)

	_, err = tmpl.Parse(content)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, false)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), err
}

func extractUsername(s string) string {
	slice := strings.Split(s, "\\")
	return slice[len(slice)-1]
}

func toJson(b []byte) (interface{}, error) {
	var m interface{}
	err := json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}

	return convertObject(m), nil
}

func toYml(b []byte) (interface{}, error) {
	m := make(map[interface{}]interface{})
	err := yaml.Unmarshal(b, &m)
	if err != nil {
		return nil, err
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
