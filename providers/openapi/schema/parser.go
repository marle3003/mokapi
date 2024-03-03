package schema

type Parser struct {
	ConvertStringToNumber bool
	Xml                   bool
}

func (p *Parser) Parse(i interface{}, r *Ref) (interface{}, error) {
	p2 := parser{convertStringToNumber: p.ConvertStringToNumber, xml: p.Xml}
	return p2.parse(i, r)
}
