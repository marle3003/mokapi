package asyncApi

type Config struct {
	Info       Info
	Servers    map[string]Server
	Channels   map[string]*Channel
	Components Components
}

type Info struct {
	Name           string `yaml:"title" json:"title"`
	Description    string
	Version        string
	TermsOfService string
	Contact        *Contact
	License        *License
	MokapiFile     string `yaml:"x-mokapifile" json:"x-mokapifile"`
}

type Contact struct {
	Name  string
	Url   string
	Email string
}

type License struct {
	Name string
	Url  string
}

type Server struct {
	Url             string
	Protocol        string
	ProtocolVersion string
	Description     string
	Variables       map[string]*ServerVariable
}

type ServerVariable struct {
	Enum        []string
	Default     string
	Description string
	Examples    []string
}

type Channel struct {
	Description string
	Subscribe   *Operation
	Publish     *Operation
}

type Operation struct {
	Id          string `yaml:"operationId" json:"operationId"`
	Summary     string
	Description string
	Message     *Message
}
type Message struct {
	Name        string
	Title       string
	Summary     string
	Description string
	ContentType string
	Payload     *Schema
	Reference   string `yaml:"$ref" json:"$ref"`
}

type Schema struct {
	Type        string
	Reference   string `yaml:"$ref" json:"$ref"`
	Description string
	Properties  map[string]*Schema
	Items       *Schema
	Faker       string `yaml:"x-faker" json:"x-faker"`
}

type Components struct {
	Schemas  map[string]*Schema
	Messages map[string]*Message
}
