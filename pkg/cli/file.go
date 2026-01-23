package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

var (
	supportedExts = []string{".yaml", ".yml", ".json"}
)

type FSReader interface {
	ReadFile(name string) ([]byte, error)
	FileExists(name string) bool
}

type FileReader struct{}

func (r *FileReader) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (r *FileReader) FileExists(name string) bool {
	_, err := os.Stat(name)
	if err != nil {
		return false
	}
	return true
}

var fileReader FSReader = &FileReader{}

func SetFileReader(r FSReader) {
	fileReader = r
}

func (c *Command) readConfigFile() error {
	file := c.configFile
	if file == "" {
		file = c.findConfigFile()
	}

	if file == "" {
		return nil
	}

	err := readConfigFile(file, c.Config)
	if err != nil {
		return fmt.Errorf("read config file '%s' failed: %w", file, err)
	}

	return mapConfigToFlags(c.Config, c.flags)
}

func (c *Command) findConfigFile() string {
	name := c.configFileName
	if name == "" {
		name = strings.ToLower(c.Name)
	}

	for _, dir := range c.configPaths {
		for _, ext := range supportedExts {
			path := filepath.Join(dir, fmt.Sprintf("%s%s", name, ext))
			if fileReader.FileExists(path) {
				return path
			}
		}
	}
	return ""
}

func readConfigFile(path string, config any) error {
	data, err := fileReader.ReadFile(path)
	if err != nil {
		return err
	}
	switch filepath.Ext(path) {
	case ".yaml", ".yml":
		err = unmarshalYaml(data, config)
	case ".json":
		err = unmarshalJson(data, config)
	default:
		err = fmt.Errorf("unsupported file extension: %v", filepath.Ext(path))
	}

	if err != nil {
		return fmt.Errorf("parse file '%v' failed: %w", path, err)
	}
	return nil
}

func mapConfigToFlags(config any, flags *FlagSet) error {
	return mapValueToFlags(reflect.ValueOf(config), "", flags)
}

func mapValueToFlags(v reflect.Value, key string, flags *FlagSet) error {
	switch v.Kind() {
	case reflect.Ptr:
		return mapValueToFlags(v.Elem(), key, flags)
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}

			name := strings.ToLower(field.Name)
			tag := field.Tag.Get("name")
			if tag != "" {
				name = strings.Split(tag, ",")[0]
			} else {
				tag = field.Tag.Get("flag")
				if tag != "" {
					name = strings.Split(tag, ",")[0]
				}
			}
			if name == "-" {
				continue
			}
			if key != "" {
				name = key + "-" + name
			}

			err := mapValueToFlags(v.Field(i), name, flags)
			if err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice:
		if _, ok := flags.GetValue(key); ok {
			var values []string
			for i := 0; i < v.Len(); i++ {
				values = append(values, fmt.Sprintf("%v", v.Index(i)))
			}
			return flags.setValue(key, values, SourceFile)
		}
		for i := 0; i < v.Len(); i++ {
			err := mapValueToFlags(v.Index(i), fmt.Sprintf("%s[%v]", key, i), flags)
			if err != nil {
				return err
			}
		}
		return nil
	case reflect.Map:
		for _, k := range v.MapKeys() {
			err := mapValueToFlags(v.MapIndex(k), fmt.Sprintf("%s-%v", key, k.Interface()), flags)
			if err != nil {
				return err
			}
		}

		return nil
	default:
		if canBeNil(v) && v.IsNil() {
			return nil
		}
		return flags.setValue(key, []string{fmt.Sprintf("%v", v.Interface())}, SourceFile)
	}
}

func canBeNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return true
	default:
		return false
	}
}

func unmarshalYaml(b []byte, config any) error {
	m := map[string]interface{}{}
	err := yaml.Unmarshal(b, m)
	if err != nil {
		return err
	}

	return mapValueToConfig(m, reflect.ValueOf(config), "yaml")
}

func unmarshalJson(b []byte, config any) error {
	m := map[string]interface{}{}
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	return mapValueToConfig(m, reflect.ValueOf(config), "json")
}

var caser = cases.Title(language.English)

func mapValueToConfig(value interface{}, configElement reflect.Value, format string) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("cannot unmarshal %v into %v", toTypeName(reflect.ValueOf(value)), toTypeName(configElement))
		}
	}()

	switch configElement.Type().Kind() {
	case reflect.Pointer:
		if configElement.IsNil() {
			configElement.Set(reflect.New(configElement.Type().Elem()))
		}
		return mapValueToConfig(value, configElement.Elem(), format)
	case reflect.Bool, reflect.Int, reflect.Float64:
		configElement.Set(reflect.ValueOf(value))
	case reflect.Int64:
		switch i := value.(type) {
		case int:
			configElement.SetInt(int64(i))
		case int64:
			configElement.SetInt(i)
		default:
			return fmt.Errorf("cannot unmarshal %v into %v", toTypeName(reflect.ValueOf(value)), toTypeName(configElement))
		}

	case reflect.String:
		if _, ok := value.(string); ok {
			t := configElement.Type()
			if !reflect.TypeOf(value).AssignableTo(t) {
				configElement.Set(reflect.ValueOf(value).Convert(t))
			} else {
				configElement.Set(reflect.ValueOf(value))
			}
		} else {
			var b []byte
			b, err = json.Marshal(value)
			if err != nil {
				return
			}
			configElement.Set(reflect.ValueOf(string(b)))
		}
	case reflect.Slice:
		v := reflect.ValueOf(value)
		if v.Type().Kind() != reflect.Slice {
			ptr := reflect.New(configElement.Type().Elem())
			err = mapValueToConfig(value, ptr.Elem(), format)
			if err != nil {
				return
			}
			configElement.Set(reflect.Append(configElement, ptr.Elem()))
		} else {
			arr, ok := value.([]any)
			if !ok {
				return fmt.Errorf("expected array, got: %v", value)
			}
			configElement.Set(reflect.Zero(configElement.Type()))
			for _, item := range arr {
				err = mapValueToConfig(item, configElement, format)
				if err != nil {
					return
				}
			}
		}
	case reflect.Struct:
		m, ok := value.(map[string]any)
		if !ok {
			i := configElement.Interface()
			_ = i
			return fmt.Errorf("expected object structure, got: %v", value)
		}
		for k, v := range m {
			f := getFieldByTag(configElement, k, format)
			if f.IsValid() {
				err = mapValueToConfig(v, f, format)
				if err != nil {
					return
				}
				continue
			}
			name := caser.String(k)
			f = configElement.FieldByNameFunc(func(f string) bool { return f == name })
			if f.IsValid() {
				err = mapValueToConfig(v, f, format)
				if err != nil {
					return
				}
				continue
			}
			f = getFieldByTag(configElement, k, "explode")
			if f.IsValid() {
				err = mapValueToConfig(v, f, format)
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
		if configElement.IsNil() {
			configElement.Set(reflect.MakeMap(configElement.Type()))
		}
		for k, v := range m {
			ptr := reflect.New(configElement.Type().Elem())
			err = mapValueToConfig(v, ptr.Elem(), format)
			if err != nil {
				return
			}
			configElement.SetMapIndex(reflect.ValueOf(k), ptr.Elem())
		}
	default:
		return fmt.Errorf("type not supported: %v", configElement.Type().Kind())
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
