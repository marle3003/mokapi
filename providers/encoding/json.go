package encoding

import (
	"encoding/json"
	"mokapi/service"
)

func MarshalJSON(obj interface{}, schema *service.Schema) ([]byte, error) {
	data := selectData(obj, schema)
	return json.Marshal(data)
}

func selectData(data interface{}, schema *service.Schema) interface{} {
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
