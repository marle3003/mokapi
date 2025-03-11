package compiler

import (
	"fmt"
	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
	"path/filepath"
	"unicode"
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
	if src != "" {
		opts := api.TransformOptions{
			Loader:     getLoader(filename),
			Target:     api.DefaultTarget,
			Platform:   api.PlatformDefault,
			Format:     api.FormatCommonJS,
			Sourcefile: filename,
			Sourcemap:  api.SourceMapInline,
		}

		result := api.Transform(src, opts)
		if len(result.Errors) > 0 {
			m := result.Errors[0]
			text := m.Text
			if len(text) > 0 {
				text = string(unicode.ToLower(rune(text[0]))) + text[1:]
			}
			return nil, fmt.Errorf("script error: %s: %s:%d:%d", text, filename, m.Location.Line, m.Location.Column)
		}
		src = string(result.Code)
	}

	src = fmt.Sprintf("(function(exports, module, require) {%s\n})", src)
	return goja.Compile(filename, src, false)
}

func (c *Compiler) Compile(filename, src string) (*goja.Program, error) {
	if src != "" {
		opts := api.TransformOptions{
			Loader:     getLoader(filename),
			Target:     api.DefaultTarget,
			Platform:   api.PlatformDefault,
			Format:     api.FormatCommonJS,
			Sourcefile: filename,
			Sourcemap:  api.SourceMapInline,
		}

		result := api.Transform(src, opts)
		if len(result.Errors) > 0 {
			m := result.Errors[0]
			text := m.Text
			if len(text) > 0 {
				text = string(unicode.ToLower(rune(text[0]))) + text[1:]
			}
			return nil, fmt.Errorf("script error: %s: %s:%d:%d", text, filename, m.Location.Line, m.Location.Column)
		}
		src = string(result.Code)
	}

	return goja.Compile(filename, src, false)
}

func getLoader(filename string) api.Loader {
	switch filepath.Ext(filename) {
	case ".ts", ".tsx":
		return api.LoaderTS
	default:
		return api.LoaderJS
	}
}
