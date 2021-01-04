package pipeline

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer/stateful"
	"github.com/pkg/errors"
	"io/ioutil"
)

var (
	lexer = stateful.MustSimple([]stateful.Rule{
		{"Comment", `(?:#|//)[^\n^\r]*?`, nil},
		{"Keyword", `pipeline|steps|true|false|stages|stage|when`, nil},
		{"Member", `[a-zA-Z_]+[a-zA-Z0-9_\[\]]*[.][a-zA-Z_]+[a-zA-Z0-9_.\[\]|'\*'|'\*\*']*`, nil},
		{"Ident", `[a-zA-Z_][a-zA-Z0-9_]*`, nil},
		{"Number", `[-+]?\d*\.?\d+([eE][-+]?\d+)?`, nil},
		{"String", `'[^'|\\']*'|"[^"|\\"]*"`, nil},
		{"Operators", `:=|==|!=|>=|<=|=>|&&|\|\||[{}=:><+-]`, nil},
		{"Punct", `[:,\(\)!]`, nil},
		{"whitespace", `\s+`, nil},
		{"EOL", `[\r\n]+`, nil},
	})
)

func getExpr(s string) (*expression, error) {
	parser := participle.MustBuild(
		&expression{},
		participle.Lexer(lexer),
		participle.UseLookahead(2),
		participle.Elide("Comment", "whitespace"),
	)

	exp := &expression{}

	err := parser.ParseString(s, s, exp)
	if err != nil {
		return nil, err
	}

	return exp, nil
}

func getPipeline(file string, name string) (*pipeline, error) {
	pipes, err := getPipelines(file)
	if err != nil {
		return nil, err
	}

	for _, p := range pipes {
		if len(p.Name) == 0 && len(name) == 0 {
			return p, nil
		} else if len(p.Name) > 0 {
			if name == p.Name[1:len(p.Name)-1] {
				return p, nil
			}
		}
	}
	return nil, errors.Errorf("pipeline '%v' in file '%v' not found", name, file)
}

func getPipelines(file string) ([]*pipeline, error) {
	parser := participle.MustBuild(
		&mokapiFile{},
		participle.Lexer(lexer),
		participle.UseLookahead(2),
		participle.Elide("Comment", "whitespace"),
	)
	mokapi := &mokapiFile{}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = parser.ParseString(file, string(data), mokapi)
	if err != nil {
		return nil, err
	}

	return mokapi.Pipelines, nil
}
