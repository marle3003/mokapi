package dynamic

import (
	"encoding/json"
	"mokapi/config/dynamic/asyncApi"
)

type ConfigMessage struct {
	ProviderName string
	Config       *ConfigurationItem
	Key          string
}

type Configuration struct {
	OpenApi  map[string]*OpenApi
	Ldap     map[string]*Ldap
	AsyncApi map[string]*asyncApi.Config
}

func NewConfiguration() *Configuration {
	return &Configuration{OpenApi: make(map[string]*OpenApi), Ldap: make(map[string]*Ldap), AsyncApi: make(map[string]*asyncApi.Config)}
}

type ConfigurationItem struct {
	OpenApi  *OpenApi
	Ldap     *Ldap
	AsyncApi *asyncApi.Config
}

func NewConfigurationItem() *ConfigurationItem {
	return &ConfigurationItem{}
}

func (c *ConfigurationItem) UnmarshalYAML(unmarshal func(interface{}) error) error {
	data := make(map[string]string)
	unmarshal(data)

	if _, ok := data["openapi"]; ok {
		openapi := &OpenApi{}
		err := unmarshal(openapi)
		if err != nil {
			return err
		}

		c.OpenApi = openapi
	} else if _, ok := data["ldap"]; ok {
		ldap := &Ldap{}
		err := unmarshal(ldap)
		if err != nil {
			return err
		}
		c.Ldap = ldap
	} else if _, ok := data["asyncapi"]; ok {
		config := &asyncApi.Config{}
		err := unmarshal(config)
		if err != nil {
			return err
		}
		c.AsyncApi = config
	}

	return nil
}

func (c *ConfigurationItem) UnmarshalJSON(b []byte) error {
	data := make(map[string]string)
	json.Unmarshal(b, &data)

	if _, ok := data["openapi"]; ok {
		openapi := &OpenApi{}
		err := json.Unmarshal(b, openapi)
		if err != nil {
			return err
		}

		c.OpenApi = openapi
	} else if _, ok := data["ldap"]; ok {
		ldap := &Ldap{}
		err := json.Unmarshal(b, ldap)
		if err != nil {
			return err
		}
		c.Ldap = ldap
	} else if _, ok := data["asyncapi"]; ok {
		config := &asyncApi.Config{}
		err := json.Unmarshal(b, config)
		if err != nil {
			return err
		}
		c.AsyncApi = config
	}

	return nil
}

func (a *AdditionalProperties) UnmarshalJSON(b []byte) error {
	switch string(b) {
	case "false":
		return nil
	case "true":
		a.Schema = &Schema{}
	default:
		a.Schema = &Schema{}
		return json.Unmarshal(b, a.Schema)
	}

	return nil
}
