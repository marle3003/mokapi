package file

import (
	"encoding/json"
	"fmt"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/provider/file"
	"mokapi/engine/common"
	"net/url"
	"os"
	"path/filepath"

	"github.com/dop251/goja"
)

type Module struct {
	host   common.Host
	rt     *goja.Runtime
	parent *dynamic.Config
}

func Enable(rt *goja.Runtime, host common.Host, parent *dynamic.Config) {
	r := &Module{
		host:   host,
		rt:     rt,
		parent: parent,
	}
	_ = rt.Set("open", r.open)
}

func Require(vm *goja.Runtime, module *goja.Object) {
	o := vm.Get("mokapi/internal").(*goja.Object)
	host := o.Get("host").Export().(common.Host)
	m := &Module{
		rt:   vm,
		host: host,
	}
	obj := module.Get("exports").(*goja.Object)
	_ = obj.Set("read", m.Read)
	_ = obj.Set("writeString", m.WriteString)
	_ = obj.Set("appendString", m.AppendString)
}

func (m *Module) open(file string, args map[string]interface{}) (any, error) {
	f, err := m.host.OpenFile(file, "")
	if err != nil {
		return "", err
	}
	dynamic.AddRef(m.parent, f)
	switch args["as"] {
	case "binary":
		return f.Raw, nil
	case "resolved":
		return m.resolve(f)
	case "string":
		fallthrough
	default:
		return string(f.Raw), nil
	}
}

func (m *Module) resolve(f *dynamic.Config) (any, error) {
	b, err := json.Marshal(f.Data)
	if err != nil {
		return nil, err
	}
	var v any
	err = json.Unmarshal(b, &v)
	return v, err
}

func (m *Module) Read(path string) (string, error) {
	p, err := m.resolvePath(path)
	if err != nil {
		return "", err
	}
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (m *Module) WriteString(path, s string) error {
	p, err := m.resolvePath(path)
	if err != nil {
		panic(fmt.Sprintf("failed to write to file: %s", err))
	}
	f, err := os.Create(p)
	if err != nil {
		return fmt.Errorf("failed to write to file: %s", err)
	}
	defer func() {
		_ = f.Close()
	}()

	_, err = f.WriteString(s)
	if err != nil {
		return fmt.Errorf("failed to write to file: %s", err)
	}
	return nil
}

func (m *Module) AppendString(path, s string) error {
	p, err := m.resolvePath(path)
	if err != nil {
		panic(fmt.Sprintf("failed to write to file: %s", err))
	}
	f, err := os.OpenFile(p,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to append to file: %s", err)
	}
	defer func() {
		_ = f.Close()
	}()

	_, err = f.WriteString(s)
	if err != nil {
		return fmt.Errorf("failed to append string: %v", err)
	}
	return nil
}

func (m *Module) resolvePath(path string) (string, error) {
	u, err := url.Parse(path)
	if err != nil || len(u.Scheme) == 0 || len(u.Opaque) > 0 {
		if !filepath.IsAbs(path) {
			cwd := m.host.Cwd()
			path = filepath.Join(cwd, path)
		}

		u, err = file.ParseUrl(path)
		if err != nil {
			return "", err
		}
	}

	if u.Scheme != "file" {
		return "", fmt.Errorf("file access only allowed from local scripts")
	}

	p := u.Path
	if len(u.Opaque) > 0 {
		p = u.Opaque
	}
	return p, nil
}
