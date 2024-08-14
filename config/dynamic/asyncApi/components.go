package asyncApi

type Components3 struct {
	Servers         map[string]Server3Ref        `yaml:"servers" json:"servers"`
	Channels        map[string]Channel3Ref       `yaml:"channels" json:"channels"`
	Schemas         map[string]SchemaRef         `yaml:"schemas" json:"schemas"`
	Messages        map[string]Message3Ref       `yaml:"messages" json:"messages"`
	Operations      map[string]Operation3Ref     `yaml:"operations" json:"operations"`
	Parameters      map[string]Parameter3Ref     `yaml:"parameters" json:"parameters"`
	CorrelationIds  map[string]CorrelationIdRef  `yaml:"correlationIds" json:"correlationIds"`
	ExternalDocs    map[string]ExternalDocRef    `yaml:"externalDocs" json:"externalDocs"`
	OperationTraits map[string]OperationTraitRef `yaml:"operationTraits" json:"operationTraits"`
	MessageTraits   map[string]MessageTraitRef   `yaml:"messageTraits" json:"messageTraits"`
}
