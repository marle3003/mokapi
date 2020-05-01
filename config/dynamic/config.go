package dynamic

type ConfigMessage struct {
	ProviderName string
	Config       *Configuration
	Key          string
}

type Configuration struct {
	OpenApi *OpenApiPart
	Ldap    *Ldap
}

type ConfigurationType struct {
}

func NewConfiguration() *Configuration {
	return &Configuration{}
}

func (c *Configuration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	data := make(map[string]string)
	unmarshal(data)

	if _, ok := data["openapi"]; ok {
		part := &OpenApiPart{}
		error := unmarshal(part)
		if error != nil {
			return error
		}

		c.OpenApi = part
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

// type Message struct {
// 	ProviderName  string
// 	Configuration *Configuration
// }

// type Configuration struct {
// 	Providers []*Provider
// }

// type Provider struct {
// 	ApiProviders  *ApiProviders  `yaml:"api"`
// 	DataProviders *DataProviders `yaml:"data"`
// }

// type ApiProviders struct {
// 	File *FileProvider
// }

// type DataProviders struct {
// 	File *FileProvider
// }
