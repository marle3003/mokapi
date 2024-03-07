package schema

import "mokapi/sortedmap"

type Schemas struct {
	sortedmap.LinkedHashMap[string, *Ref]
}
