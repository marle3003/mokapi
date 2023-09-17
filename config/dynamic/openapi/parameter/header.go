package parameter

import (
	"fmt"
	"mokapi/config/dynamic/openapi/schema"
	"net/http"
)

func parseHeader(p *Parameter, r *http.Request) (rp RequestParameterValue, err error) {
	rp.Raw = r.Header.Get(p.Name)
	if len(rp.Raw) == 0 && p.Required {
		return rp, fmt.Errorf("header parameter '%v' is required", p.Name)
	}

	rp.Value, err = schema.ParseString(rp.Raw, p.Schema)
	return
}
