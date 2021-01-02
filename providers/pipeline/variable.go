package pipeline

import "mokapi/providers/pipeline/types"

type Variable interface {
	Name() string
	Value() types.Object
}
