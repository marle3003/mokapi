package types

import (
	"github.com/pkg/errors"
	"reflect"
	"strings"
)

type KeyValuePair struct {
	key   string
	value Object
}

func NewKeyValuePair(key string, value Object) *KeyValuePair {
	return &KeyValuePair{key: key, value: value}
}

func (kv *KeyValuePair) Equals(obj Object) bool {
	if other, ok := obj.(*KeyValuePair); ok {
		if kv.key != other.key {
			return false
		}
		if kv.value != nil {
			return kv.value.Equals(other.value)
		}
		return other.value == nil
	}
	return false
}

func (kv *KeyValuePair) String() string {
	var sb strings.Builder
	sb.WriteString("[")
	sb.WriteString(kv.key)
	sb.WriteString(",")
	if kv.value != nil {
		sb.WriteString(kv.value.String())
	}
	sb.WriteString("]")

	return sb.String()
}

func (kv *KeyValuePair) GetType() reflect.Type {
	return reflect.TypeOf(kv)
}

func (kv *KeyValuePair) Get(name string) (Object, error) {
	switch strings.ToLower(name) {
	case "key":
		return NewString(kv.key), nil
	case "value":
		return kv.value, nil
	}
	return nil, errors.Errorf("type %v does not contain member %v", kv.GetType(), name)
}

func (kv *KeyValuePair) Set(name string, value Object) error {
	switch strings.ToLower(name) {
	case "key":
		kv.key = value.String()
		return nil
	case "value":
		kv.value = value
		return nil
	}
	return errors.Errorf("type %v does not contain member %v", kv.GetType(), name)
}

func (kv *KeyValuePair) Invoke(path *Path, _ []Object) (Object, error) {
	if path.Head() == "" {
		return kv, nil
	}
	return nil, errors.Errorf("member '%v' in path '%v' is not defined on type keyvaluepair", path.Head(), path)
}
