package kafka

import (
	"mokapi/config/dynamic/openapi/schema"
)

type Operation struct {
	GroupId *schema.Schema `yaml:"groupId" json:"groupId"`
}
