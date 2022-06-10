package compiler

import (
	"fmt"
	"github.com/dop251/goja"
)

type Compiler struct {
	babel *babel
}

func New() (*Compiler, error) {
	c := &Compiler{}

	var err error
	c.babel, err = newBabel()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Compiler) CompileModule(filename, src string) (*goja.Program, error) {
	source, err := c.babel.Transform(src)
	if err != nil {
		return nil, err
	}
	source = fmt.Sprintf("(function(exports, module) {%s\n})", source)
	return goja.Compile(filename, source, false)
}

func (c *Compiler) Compile(filename, src string) (*goja.Program, error) {
	source, err := c.babel.Transform(src)
	if err != nil {
		return nil, err
	}
	return goja.Compile(filename, source, false)
}
