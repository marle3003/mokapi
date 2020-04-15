package data

import (
	"fmt"
	"mokapi/config"
)

type StaticDataProvider struct {
	data map[interface{}]interface{}
}

func NewStaticDataProvider(data map[interface{}]interface{}) *StaticDataProvider {
	return &StaticDataProvider{data: data}
}

func (provider *StaticDataProvider) Provide(parameters map[string]string, schema *config.Schema) (interface{}, error) {
	data := provider.getData(schema.Resource)
	data = filterData(data, parameters)
	data = selectData(data, schema)
	return data, nil
}

func (provider *StaticDataProvider) getData(resource string) interface{} {
	if resource != "" {
		return convertData(provider.data[resource])
	}
	return convertData(provider.data)
}

func convertData(o interface{}) interface{} {
	if a, ok := o.([]interface{}); ok {
		var result []interface{}
		result = make([]interface{}, len(a))
		for i, e := range a {
			result[i] = convertData(e)
		}
		return result
	} else {
		return convertObject(o)
	}
}

func convertObject(o interface{}) interface{} {
	if m, ok := o.(map[interface{}]interface{}); ok {
		result := make(map[string]interface{}, len(m))
		for k, v := range m {
			propertyName := fmt.Sprint(k)
			result[propertyName] = convertData(v)
		}
		return result
	}
	return o
}

func filterData(data interface{}, parameters map[string]string) interface{} {
	if parameters == nil || len(parameters) == 0 {
		return data
	}

	if list, ok := data.([]interface{}); ok {
		result := make([]interface{}, 0)
		for _, d := range list {
			match := true
			o := d.(map[string]interface{})
			for p, v := range parameters {
				if o[p] != v {
					match = false
					break
				}
			}
			if match {
				result = append(result, o)
			}
		}
		return result
	}
	return data
}

func selectData(data interface{}, schema *config.Schema) interface{} {
	if schema.Type == "array" {
		if list, ok := data.([]interface{}); ok {
			for i, e := range list {
				list[i] = selectData(e, schema.Items)
			}
			return list
		}
		// todo error handling
		return nil
	} else if schema.Type == "object" {
		o := data.(map[string]interface{})
		selectedData := make(map[string]interface{}, 5)
		for propertyName, propertySchema := range schema.Properties {
			if p, ok := o[propertyName]; ok {
				selectedData[propertyName] = selectData(p, propertySchema)
			}
		}
		return selectedData
	}
	return data
}
