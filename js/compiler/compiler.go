package compiler

import (
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

func (c *Compiler) Compile(filename, src string) (*goja.Program, error) {
	prg, err := goja.Compile(filename, src, false)
	if err != nil {
		src, err = c.babel.Transform(src)
		if err != nil {
			return nil, err
		}
		prg, err = goja.Compile(filename, src, false)
		if err != nil {
			return nil, err
		}
	}

	return prg, nil
}
