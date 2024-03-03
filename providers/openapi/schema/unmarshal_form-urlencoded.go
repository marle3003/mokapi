package schema

import "net/url"

func readFormUrlEncoded(data []byte, r *Ref) (interface{}, error) {
	values, err := url.ParseQuery(string(data))
	m := map[string]interface{}{}
	p := parser{convertStringToNumber: true}
	for k, v := range values {
		s := r.Value.Properties.Get(k)
		switch s.Value.Type {
		case "array":
			m[k], err = p.parse(v, s)
		default:
			m[k], err = p.parse(v[0], s)
		}
		if err != nil {
			return nil, err
		}
	}
	return m, err
}
