package dynamic

import "encoding/json"

type ConfigMessage struct {
	ProviderName string
	Config       *ConfigurationItem
	Key          string
}

type Configuration struct {
	OpenApi map[string]*OpenApi
	Ldap    map[string]*Ldap
}

func NewConfiguration() *Configuration {
	return &Configuration{OpenApi: make(map[string]*OpenApi), Ldap: make(map[string]*Ldap)}
}

type ConfigurationItem struct {
	OpenApi *OpenApi
	Ldap    *Ldap
}

func NewConfigurationItem() *ConfigurationItem {
	return &ConfigurationItem{}
}

func (c *ConfigurationItem) UnmarshalYAML(unmarshal func(interface{}) error) error {
	data := make(map[string]string)
	unmarshal(data)

	if _, ok := data["openapi"]; ok {
		openapi := &OpenApi{}
		error := unmarshal(openapi)
		if error != nil {
			return error
		}

		c.OpenApi = openapi
	} else if _, ok := data["ldap"]; ok {
		ldap := &Ldap{}
		error := unmarshal(ldap)
		if error != nil {
			return error
		}
		c.Ldap = ldap
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
