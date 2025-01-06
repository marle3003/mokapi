package schema

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
)

type UnionType[T1 any, T2 any] struct {
	t uint8
	A T1
	B T2
}

func NewUnionTypeA[T1 any, T2 any](v T1) *UnionType[T1, T2] {
	return &UnionType[T1, T2]{t: 0, A: v}
}

func NewUnionTypeB[T1 any, T2 any](v T2) *UnionType[T1, T2] {
	return &UnionType[T1, T2]{t: 1, B: v}
}

func (ut *UnionType[T1, T2]) Value() any {
	if ut.t == 0 {
		return ut.A
	}
	return ut.B
}

func (ut *UnionType[T1, T2]) IsA() bool {
	return ut.t == 0
}

func (ut *UnionType[T1, T2]) UnmarshalYAML(value *yaml.Node) error {
	err := value.Decode(&ut.A)
	if err == nil {
		return nil
	}

	ut.t = 1
	return value.Decode(&ut.B)
}

func (ut *UnionType[T1, T2]) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &ut.A); err == nil {
		return nil
	}

	ut.t = 1
	return json.Unmarshal(b, &ut.B)
}

func (ut *UnionType[T1, T2]) String() string {
	if ut.IsA() {
		return fmt.Sprintf("%v", ut.A)
	}
	return fmt.Sprintf("%v", ut.B)
}
