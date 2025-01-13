package parser

import (
	log "github.com/sirupsen/logrus"
	"mokapi/schema/json/schema"
)

func checkValueIsInEnum(i interface{}, enum []interface{}, entrySchema *schema.Schema) error {
	found := false
	p := Parser{ConvertToSortedMap: true}
	for _, e := range enum {
		v, err := p.parse(e, entrySchema)
		if err != nil {
			log.Errorf("unable to parse enum value %v to %v: %v", ToString(e), entrySchema, err)
			continue
		}
		if compare(i, v) {
			found = true
			break
		}
	}
	if !found {
		return Errorf("enum", "value '%v' does not match one in the enumeration %v", ToString(i), ToString(enum))
	}

	return nil
}
