package cli

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"reflect"
	"strings"
)

type Command struct {
	Name     string
	Short    string
	Long     string
	Example  string
	Config   any
	Commands []*Command
	Run      func(cmd *Command, args []string) error

	envPrefix string
	args      []string
	flags     *FlagSet
}

func (c *Command) Execute() error {
	args := c.args
	if args == nil {
		args = os.Args[1:]
	}

	cmd := c
	envPrefix := c.envPrefix

	if len(args) > 0 {
		for _, child := range c.Commands {
			if child.Name == args[0] {
				cmd = child
				args = args[1:]
				if cmd.envPrefix != "" {
					envPrefix = cmd.envPrefix
				}
			}
		}
	}

	m, err := parseFlags(args, envPrefix)
	if err != nil {
		return err
	}

	if cmd.Config != nil {
		var file string
		file, err = readConfigFileFromFlags(m, cmd.Config)
		if err != nil {
			return err
		}
		if file != "" {
			var valuesFromConfig map[string][]string
			valuesFromConfig, err = getMapFromConfig(cmd.Config, cmd.Flags())
			if err != nil {
				return fmt.Errorf("reading config file '%s' failed: %w", file, err)
			}
			for k, v := range valuesFromConfig {
				if _, ok := m[k]; !ok {
					m[k] = v
				}
			}
		}
	}

	var positional []string
	for k, v := range m {
		switch k {
		case "args":
			positional = v
		default:
			err = cmd.Flags().setValue(k, v)
			if err != nil {
				return fmt.Errorf("failed to set flag '%s': %w", k, err)
			}
		}
	}

	for k, v := range cmd.flags.flags {
		if _, ok := m[k]; !ok && v.DefaultValue != "" {
			m[k] = []string{v.DefaultValue}
		}
	}

	if cmd.Config != nil {
		// reset configs, because values or now in flag set
		clearConfig(cmd.Config)
		b := flagConfigBinder{}
		err = b.Decode(m, cmd.Config)
		if err != nil {
			return fmt.Errorf("failed to bind flags to config: %w", err)
		}
	}

	if cmd.Run != nil {
		return cmd.Run(cmd, positional)
	} else {
		return fmt.Errorf("no command run specified")
	}
}

func (c *Command) SetArgs(args []string) {
	c.args = args
}

func (c *Command) Flags() *FlagSet {
	if c.flags == nil {
		c.flags = &FlagSet{}
	}
	return c.flags
}

// SetEnvPrefix defines prefix of the environment variables to be considered
// With the prefix "mokapi", only environment variables with MOKAPI_ are considered.
func (c *Command) SetEnvPrefix(in string) {
	if in != "" {
		in = in + "_"
	}
	c.envPrefix = strings.ToUpper(in)
}

func getMapFromConfig(cfg any, flags *FlagSet) (map[string][]string, error) {
	return getMapFrom(reflect.ValueOf(cfg), "", flags)
}

func getMapFrom(v reflect.Value, key string, flags *FlagSet) (map[string][]string, error) {
	switch v.Kind() {
	case reflect.Ptr:
		return getMapFrom(v.Elem(), key, flags)
	case reflect.Struct:
		t := v.Type()
		result := map[string][]string{}
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

			m, err := getMapFrom(v.Field(i), name, flags)
			if err != nil {
				return nil, err
			} else if m == nil {
				continue
			}
			for k, val := range m {
				result[k] = val
			}
		}
		return result, nil
	case reflect.Slice:
		if _, err := flags.GetValue(key); err != nil {
			var notFound *FlagNotFound
			if errors.As(err, &notFound) {
				return nil, nil
			}
			return nil, err
		}
		var values []string
		for i := 0; i < v.Len(); i++ {
			values = append(values, fmt.Sprintf("%v", v.Index(i)))
		}
		return map[string][]string{key: values}, nil
	default:
		if canBeNil(v) && v.IsNil() {
			return nil, nil
		}
		return map[string][]string{key: {fmt.Sprintf("%v", v.Interface())}}, nil
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

func clearConfig(v interface{}) {
	p := reflect.ValueOf(v).Elem()
	p.Set(reflect.Zero(p.Type()))
}
