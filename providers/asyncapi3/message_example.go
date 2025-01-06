package asyncapi3

type MessageExample struct {
	Name    string                 `yaml:"name" json:"name"`
	Summary string                 `yaml:"summary" json:"summary"`
	Headers map[string]interface{} `yaml:"headers" json:"headers"`
	Payload interface{}            `yaml:"payload" json:"payload"`
}
