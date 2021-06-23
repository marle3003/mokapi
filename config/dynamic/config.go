package dynamic

import "reflect"

var (
	configTypes []configType
)

type ChangeEventHandler func(path string, c Config, r ConfigReader) (bool, Config)

type Config interface {
}

type ConfigReader interface {
	Read(path string, c Config, h ChangeEventHandler) error
}

type configType struct {
	header       string
	config       reflect.Type
	eventHandler ChangeEventHandler
}

type configItem struct {
	eventHandler ChangeEventHandler
	item         Config
}

func NewEmptyEventHandler(parent Config) ChangeEventHandler {
	return func(path string, c Config, r ConfigReader) (bool, Config) { return true, parent }
}

func Register(header string, c Config, h ChangeEventHandler) {
	val := reflect.ValueOf(c).Elem()
	configTypes = append(configTypes, configType{header, val.Type(), h})
}

func (ci *configItem) UnmarshalYAML(unmarshal func(interface{}) error) error {
	data := make(map[string]string)
	_ = unmarshal(data)

	for _, c := range configTypes {
		if _, ok := data[c.header]; ok {
			ci.item = reflect.New(c.config).Interface()
			ci.eventHandler = c.eventHandler
			err := unmarshal(ci.item)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
