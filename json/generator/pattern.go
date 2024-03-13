package generator

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"math"
	"math/rand"
	"regexp/syntax"
	"strings"
)

func StringPattern() *Tree {
	return &Tree{
		Name: "Pattern",
		Test: func(r *Request) bool {
			return r.Schema.IsString() && r.Schema.Pattern != ""
		},
		Fake: func(r *Request) (v interface{}, err error) {
			re, err := syntax.Parse(r.Schema.Pattern, syntax.Perl)
			if err != nil {
				return nil, fmt.Errorf("could not parse regex string: %v", r.Schema.Pattern)
			}

			// Panic catch
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("%v", r)
				}
			}()

			g := regexGenerator{ra: r.g.rand}
			g.regexGenerate(re, len(r.Schema.Pattern)*100)
			if r.Schema != nil && r.Schema.MinLength != nil {
				min := *r.Schema.MinLength
				max := min + 100
				if r.Schema.MaxLength != nil {
					max = *r.Schema.MaxLength
				}
				err = g.refill(min, max)
			}
			if err != nil {
				return nil, fmt.Errorf("cannot generate value for pattern %v and minimum length %v", r.Schema.Pattern, *r.Schema.MinLength)
			}
			return g.sb.String(), nil
		},
	}
}

type regexGenerator struct {
	sb      strings.Builder
	ra      *rand.Rand
	fillers []func(max int)
}

func (g *regexGenerator) refill(min, max int) error {
	if g.sb.Len() < min && len(g.fillers) == 0 {
		return fmt.Errorf("no fillers exists")
	}

	const limit = 30
	counter := 0
	for g.sb.Len() < min {
		nOld := g.sb.Len()
		for _, fill := range g.fillers {
			n := g.sb.Len()
			fill(max - n)
		}
		if nOld == g.sb.Len() {
			if counter == limit {
				return fmt.Errorf("fillers no longer provide data")
			}
			counter++
		}
	}
	return nil
}

// regexGenerate based on https://github.com/brianvoe/gofakeit
func (g *regexGenerator) regexGenerate(re *syntax.Regexp, max int) {
	if max-g.sb.Len() <= 0 {
		panic("length limit reached when generating output")
	}

	op := re.Op
	switch op {
	case syntax.OpNoMatch: // matches no strings
		// Do Nothing
	case syntax.OpEmptyMatch: // matches empty string

	case syntax.OpLiteral: // matches Runes sequence
		for _, ru := range re.Rune {
			g.sb.WriteRune(ru)
		}
	case syntax.OpCharClass: // matches Runes interpreted as range pair list
		// number of possible chars
		sum := 0
		for i := 0; i < len(re.Rune); i += 2 {
			sum += int(re.Rune[i+1]-re.Rune[i]) + 1
			if re.Rune[i+1] == 0x10ffff { // rune range end
				sum = -1
				break
			}
		}

		// pick random char in range (inverse match group)
		if sum == -1 {
			var chars []uint8
			for j := 0; j < len(allStr); j++ {
				c := allStr[j]

				// Check c in range
				for i := 0; i < len(re.Rune); i += 2 {
					if rune(c) >= re.Rune[i] && rune(c) <= re.Rune[i+1] {
						chars = append(chars, c)
						break
					}
				}
			}
			if len(chars) > 0 {
				g.sb.Write([]byte{chars[g.ra.Intn(len(chars))]})
				return
			}
		}

		r := g.ra.Intn(sum)
		var ru rune
		sum = 0
		for i := 0; i < len(re.Rune); i += 2 {
			gap := int(re.Rune[i+1]-re.Rune[i]) + 1
			if sum+gap > r {
				ru = re.Rune[i] + rune(r-sum)
				break
			}
			sum += gap
		}

		g.sb.WriteRune(ru)
	case syntax.OpAnyCharNotNL, syntax.OpAnyChar: // matches any character(and except newline)
		g.sb.WriteString(string(allStr[g.ra.Int63()%int64(len(allStr))]))
	case syntax.OpBeginLine: // matches empty string at beginning of line
	case syntax.OpEndLine: // matches empty string at end of line
	case syntax.OpBeginText: // matches empty string at beginning of text
	case syntax.OpEndText: // matches empty string at end of text
	case syntax.OpWordBoundary: // matches word boundary `\b`
	case syntax.OpNoWordBoundary: // matches word non-boundary `\B`
	case syntax.OpCapture: // capturing subexpression with index Cap, optional name Name
		g.regexGenerate(re.Sub0[0], max)
	case syntax.OpStar: // matches Sub[0] zero or more times
		g.opStar(re, max)
		index := g.sb.Len()
		g.fillers = append(g.fillers, func(max int) {
			g.fill(index, func() {
				g.opStar(re, max)
			})
		})
	case syntax.OpPlus: // matches Sub[0] one or more times
		g.opPlus(re, 1, max)
		index := g.sb.Len()
		g.fillers = append(g.fillers, func(max int) {
			g.fill(index, func() {
				g.opPlus(re, 0, max)
			})
		})
	case syntax.OpQuest: // matches Sub[0] zero or one times
		used := g.opQuest(re, max)
		if !used {
			index := g.sb.Len()
			g.fillers = append(g.fillers, func(max int) {
				if used {
					return
				}
				g.fill(index, func() {
					used = g.opQuest(re, max)
				})
			})
		}
	case syntax.OpRepeat: // matches Sub[0] at least Min times, at most Max (Max == -1 is no limit)
		count := g.opRepeat(re, max)
		re.Max = re.Max - count
		re.Min = 0
		index := g.sb.Len()
		g.fillers = append(g.fillers, func(max int) {
			g.fill(index, func() {
				count = g.opRepeat(re, max)
				re.Max = re.Max - count
			})
		})
	case syntax.OpConcat: // matches concatenation of Subs
		for _, rs := range re.Sub {
			g.regexGenerate(rs, max)
		}
	case syntax.OpAlternate: // matches alternation of Subs
		g.regexGenerate(re.Sub[gofakeit.Number(0, len(re.Sub)-1)], max)
	}
}

func (g *regexGenerator) opStar(re *syntax.Regexp, limit int) {
	max := int(math.Min(float64(limit), float64(10)))
	for i := 0; i < gofakeit.Number(0, max); i++ {
		for _, rs := range re.Sub {
			g.regexGenerate(rs, limit)
		}
	}
}

func (g *regexGenerator) opPlus(re *syntax.Regexp, min, limit int) {
	max := int(math.Min(10, float64(limit)))
	for i := 0; i < gofakeit.Number(min, max); i++ {
		for _, rs := range re.Sub {
			g.regexGenerate(rs, limit)

		}
	}
}

func (g *regexGenerator) opRepeat(re *syntax.Regexp, limit int) int {
	max := int(math.Min(float64(re.Max), math.Min(float64(limit), float64(10))))
	count := 0
	if re.Max > re.Min {
		count = g.ra.Intn(max - re.Min + 1)
	}
	count = int(math.Max(float64(re.Min), float64(re.Min+count)))
	for i := 0; i < count; i++ {
		for _, rs := range re.Sub {
			g.regexGenerate(rs, limit)
		}
	}
	return count
}

func (g *regexGenerator) opQuest(re *syntax.Regexp, limit int) bool {
	n := gofakeit.Number(0, 1)
	if n == 1 {
		for _, rs := range re.Sub {
			g.regexGenerate(rs, limit)
		}
		return true
	}
	return false
}

func (g *regexGenerator) fill(index int, filler func()) {
	s := g.sb.String()
	g.sb.Reset()
	filler()
	fill := fmt.Sprintf("%v%v%v", s[:index-1], g.sb.String(), s[index-1:])
	g.sb.Reset()
	g.sb.WriteString(fill)
}
