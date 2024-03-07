package schema

import (
	"fmt"
	"mokapi/json/ref"
)

type Ref struct {
	ref.Reference
	Value *Schema
}

func (r *Ref) IsAny() bool {
	return r == nil || r.Value == nil || len(r.Value.Type) == 0
}

func (r *Ref) IsString() bool {
	return r != nil && r.Value != nil && r.Value.IsString()
}

func (r *Ref) IsInteger() bool {
	return r != nil && r.Value != nil && r.Value.IsInteger()
}

func (r *Ref) IsNumber() bool {
	return r != nil && r.Value != nil && r.Value.IsNumber()
}

func (r *Ref) IsArray() bool {
	return r != nil && r.Value != nil && r.Value.IsArray()
}

func (r *Ref) IsObject() bool {
	return r != nil && r.Value != nil && r.Value.IsObject()
}

func (r *Ref) IsDictionary() bool {
	return r != nil && r.Value != nil && r.Value.IsDictionary()
}

func (r *Ref) Type() string {
	if r == nil || r.Value == nil || len(r.Value.Type) == 0 {
		return ""
	}
	if len(r.Value.Type) == 1 {
		return r.Value.Type[0]
	}
	return fmt.Sprintf("%v", r.Value.Type)
}
