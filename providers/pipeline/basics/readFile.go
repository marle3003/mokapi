package basics

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"
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

	bytes, err := readFile(file, e.AsTemplate)
	if err != nil {
		return nil, err
	}

	if e.AsString {
		return string(bytes), nil
	}

	return bytes, nil
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
