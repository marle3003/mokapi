package types

import (
	"github.com/pkg/errors"
	"mokapi/providers/pipeline/lang/token"
	"reflect"
	"strings"
)

type KeyValuePair struct {
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

func (kv *KeyValuePair) Set(o Object) error {
	kv.Value = o
	return nil
}

func (kv *KeyValuePair) GetType() reflect.Type {
	return reflect.TypeOf(kv.Value)
}

func (kv *KeyValuePair) Elem() interface{} {
	return map[string]interface{}{kv.Key: kv.Value.Elem()}
}

func (kv *KeyValuePair) GetField(name string) (Object, error) {
	return getField(kv, name)
}

func (kv *KeyValuePair) HasField(name string) bool {
	return hasField(kv, name)
}

func (kv *KeyValuePair) InvokeFunc(name string, args map[string]Object) (Object, error) {
	return invokeFunc(kv, name, args)
}

func (kv *KeyValuePair) SetField(field string, value Object) error {
	return setField(kv, field, value)
}

func (kv *KeyValuePair) InvokeOp(op token.Token, _ Object) (Object, error) {
	return nil, errors.Errorf("type keyvaluepair does not support operator %v", op)
}
