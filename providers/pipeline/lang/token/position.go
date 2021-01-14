package token

type Position struct {
	Line   int
	Column int
}

func (p *Position) NewLine() {
	p.Line++
	p.Column = 0
}
