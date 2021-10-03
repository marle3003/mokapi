package kafka

import "mokapi/config/dynamic/openapi"

type Operation struct {
	GroupId *openapi.Schema `yaml:"groupId" json:"groupId"`
}
