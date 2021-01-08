package types

import (
	"strings"
)

type KeyValuePair struct {
	ObjectImpl
	Key   string
	Value Object
}

func NewKeyValuePair(key string, value Object) *KeyValuePair {
	return &KeyValuePair{Key: key, Value: value}
}

func (kv *KeyValuePair) String() string {
	var sb strings.Builder
	sb.WriteString("[")
	sb.WriteString(kv.Key)
	sb.WriteString(",")
	if kv.Value != nil {
		sb.WriteString(kv.Value.String())
	}
	sb.WriteString("]")

	return sb.String()
}
