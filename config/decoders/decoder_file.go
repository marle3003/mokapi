package decoders

import (
	"encoding/json"
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"reflect"
)

var (
	searchPaths = []string{".", "/etc/mokapi"}
	fileNames   = []string{"mokapi.yaml", "mokapi.yml"}
)

type ReadFileFS func(path string) ([]byte, error)

type FileDecoder struct {
	filename string
	readFile ReadFileFS
}

func NewDefaultFileDecoder() *FileDecoder {
	return NewFileDecoder(os.ReadFile)
}

func NewFileDecoder(readFile ReadFileFS) *FileDecoder {
	return &FileDecoder{readFile: readFile}
}

func (f *FileDecoder) Decode(flags map[string][]string, element interface{}) error {
	if len(f.filename) == 0 {
		if val, ok := flags["configfile"]; ok {
			delete(flags, "configfile")
			f.filename = val[0]
		} else if val, ok := flags["config-file"]; ok {
			delete(flags, "config-file")
			f.filename = val[0]
		} else if val, ok := flags["cli-input"]; ok {
			delete(flags, "cli-input")
			f.filename = val[0]
		}
	}

	if len(f.filename) > 0 {
		return f.read(f.filename, element)
	}

	for _, dir := range searchPaths {
		for _, name := range fileNames {
			path := filepath.Join(dir, name)
			if err := f.read(path, element); err == nil {
				return nil
			} else if !os.IsNotExist(err) {
				return err
			}
		}
	}

	return nil
}

func (f *FileDecoder) read(path string, element interface{}) error {
	data, err := f.readFile(path)
	if err != nil {
		return err
	}
	switch filepath.Ext(path) {
	case ".yaml", ".yml":
		err = unmarshalYaml(data, reflect.ValueOf(element))
	case ".json":
		err = unmarshalJson(data, reflect.ValueOf(element))
	default:
		err = fmt.Errorf("unsupported file extension: %v", filepath.Ext(path))
	}

	if err != nil {
		return fmt.Errorf("parse file '%v' failed: %w", path, err)
	}
	return nil
}

func unmarshalYaml(b []byte, element reflect.Value) error {
	m := map[string]interface{}{}
	err := yaml.Unmarshal(b, m)
	if err != nil {
		return err
	}

	return mapConfig(m, element, "yaml")
}

func unmarshalJson(b []byte, element reflect.Value) error {
	m := map[string]interface{}{}
	err := yaml.Unmarshal(b, m)
	if err != nil {
		return err
	}

	return mapConfig(m, element, "json")
}

var caser = cases.Title(language.English)

func mapConfig(value interface{}, element reflect.Value, format string) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("cannot unmarshal %v into %v", toTypeName(reflect.ValueOf(value)), toTypeName(element))
		}
	}()

	switch element.Type().Kind() {
	case reflect.Pointer:
		if element.IsNil() {
			element.Set(reflect.New(element.Type().Elem()))
		}
		return mapConfig(value, element.Elem(), format)
	case reflect.Bool, reflect.Int, reflect.Float64:
		element.Set(reflect.ValueOf(value))
	case reflect.Int64:
		switch i := value.(type) {
		case int:
			element.SetInt(int64(i))
		case int64:
			element.SetInt(i)
		default:
			return fmt.Errorf("cannot unmarshal %v into %v", toTypeName(reflect.ValueOf(value)), toTypeName(element))
		}

	case reflect.String:
		if _, ok := value.(string); ok {
			t := element.Type()
			if !reflect.TypeOf(value).AssignableTo(t) {
				element.Set(reflect.ValueOf(value).Convert(t))
			} else {
				element.Set(reflect.ValueOf(value))
			}
		} else {
			var b []byte
			b, err = json.Marshal(value)
			if err != nil {
				return
			}
			element.Set(reflect.ValueOf(string(b)))
		}
	case reflect.Slice:
		v := reflect.ValueOf(value)
		if v.Type().Kind() != reflect.Slice {
			ptr := reflect.New(element.Type().Elem())
			err = mapConfig(value, ptr.Elem(), format)
			if err != nil {
				return
			}
			element.Set(reflect.Append(element, ptr.Elem()))
		} else {
			arr, ok := value.([]interface{})
			if !ok {
				return fmt.Errorf("expected array, got: %v", value)
			}
			for _, item := range arr {
				err = mapConfig(item, element, format)
				if err != nil {
					return
				}
			}
		}
	case reflect.Struct:
		m, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected object structure, got: %v", value)
		}
		for k, v := range m {
			f := getFieldByTag(element, k, format)
			if f.IsValid() {
				err = mapConfig(v, f, format)
				if err != nil {
					return
				}
				continue
			}
			name := caser.String(k)
			f = element.FieldByNameFunc(func(f string) bool { return f == name })
			if f.IsValid() {
				err = mapConfig(v, f, format)
				if err != nil {
					return
				}
				continue
			}
			f = getFieldByTag(element, k, "explode")
			if f.IsValid() {
				err = mapConfig(v, f, format)
				if err != nil {
					return
				}
			}
		}
	case reflect.Map:
		m, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected object structure, got: %v", value)
		}
		if element.IsNil() {
			element.Set(reflect.MakeMap(element.Type()))
		}
		for k, v := range m {
			ptr := reflect.New(element.Type().Elem())
			err = mapConfig(v, ptr.Elem(), format)
			if err != nil {
				return
			}
			element.SetMapIndex(reflect.ValueOf(k), ptr.Elem())
		}
	default:
		return fmt.Errorf("type not supported: %v", element.Type().Kind())
	}

	return
}

func toTypeName(v reflect.Value) string {
	switch v.Type().Kind() {
	case reflect.Slice:
		return "array"
	case reflect.Struct, reflect.Map:
		return "object"
	default:
		return v.Type().Kind().String()
	}
}
