package parser

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/schema/json/schema"
)

func checkValueIsInEnum(i interface{}, enum []interface{}, entrySchema *schema.Schema) error {
	found := false
	p := Parser{ConvertToSortedMap: true}
	for _, e := range enum {
		v, err := p.Parse(e, &schema.Ref{Value: entrySchema})
		if err != nil {
			log.Errorf("unable to parse enum value %v to %v: %v", toString(e), entrySchema, err)
			continue
		}
		if compare(i, v) {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("value '%v' does not match one in the enumeration %v", toString(i), toString(enum))
	}

	return nil
}
