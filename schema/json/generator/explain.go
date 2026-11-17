package generator

import (
	"strings"
)

// Sources of a value that is not produced by a node of the faker tree. The node of such a
// value is NoNode.
const (
	SourceNode     = "node"
	SourceEnum     = "enum"
	SourceConst    = "const"
	SourceExamples = "examples"
	SourceContext  = "context"
	SourceNull     = "null"
)

const (
	// NoNode is the node of a value that was not produced by a node of the faker tree.
	NoNode = "-"
	// noPropertyPath is the property path of a value that is not part of an object.
	noPropertyPath = "(value)"
	// itemSegment is the segment of the output path for the items of an array.
	itemSegment = "[]"
	// additionalPropertySegment is the segment of the output path for additional properties,
	// their names are not known before the value is resolved.
	additionalPropertySegment = "*"
)

// Match tells how a value of the result was produced.
type Match struct {
	// PropertyPath is the dot-joined path of the value in the result, e.g. address.city or tags[].name
	PropertyPath string `json:"propertyPath"`
	// Node is the path of the node that produced the value, fakeBySchema for the fallback
	// or NoNode if the value does not come from a node.
	Node string `json:"node"`
	// Tokens are the path tokens the node was matched with.
	Tokens []string `json:"tokens"`
	Source string   `json:"source"`
	Value  any      `json:"value"`
}

type Result struct {
	Value   any     `json:"value"`
	Matches []Match `json:"matches"`
}

// Explain generates a value like New and reports which node produced which value.
func Explain(r *Request) (Result, error) {
	r.trace = &trace{prefix: len(r.Path)}
	r.at = append([]string{}, r.Path...)
	v, err := New(r)
	if err != nil {
		return Result{}, err
	}
	return Result{Value: v, Matches: r.trace.matchesIn(v)}, nil
}

// trace collects the leaf values generated for a request. The generator builds closures and
// fakes them later, a node may be dropped when the value is faked (fallback) and faking is
// retried, so a value is recorded when it is faked and the last record for a property path wins.
type trace struct {
	// prefix is the number of segments at the beginning of the output path that come from the
	// request path and are not part of the result.
	prefix  int
	records []record
	index   map[string]int
}

type record struct {
	at []string
	Match
}

// record is safe to call on a nil trace, which is the case for all requests except Explain.
func (t *trace) record(at []string, m Match) {
	if t == nil {
		return
	}
	if m.Tokens == nil {
		m.Tokens = []string{}
	}
	m.PropertyPath = joinPropertyPath(at)
	rec := record{at: at, Match: m}
	if i, ok := t.index[m.PropertyPath]; ok {
		t.records[i] = rec
		return
	}
	if t.index == nil {
		t.index = map[string]int{}
	}
	t.index[m.PropertyPath] = len(t.records)
	t.records = append(t.records, rec)
}

// matchesIn returns the records that are part of the final value. Optional properties are
// faked but not always emitted.
func (t *trace) matchesIn(value any) []Match {
	matches := []Match{}
	for _, rec := range t.records {
		if hasValueAt(value, rec.at[t.prefix:]) {
			matches = append(matches, rec.Match)
		}
	}
	return matches
}

func hasValueAt(v any, at []string) bool {
	if len(at) == 0 {
		return true
	}
	switch segment := at[0]; v := v.(type) {
	case []any:
		if segment != itemSegment {
			return false
		}
		for _, item := range v {
			if hasValueAt(item, at[1:]) {
				return true
			}
		}
	case map[string]any:
		if segment == additionalPropertySegment {
			for _, item := range v {
				if hasValueAt(item, at[1:]) {
					return true
				}
			}
			return false
		}
		if item, ok := v[segment]; ok {
			return hasValueAt(item, at[1:])
		}
	}
	return false
}

func joinPropertyPath(at []string) string {
	if len(at) == 0 {
		return noPropertyPath
	}
	var sb strings.Builder
	for i, segment := range at {
		if i > 0 && segment != itemSegment {
			sb.WriteByte('.')
		}
		sb.WriteString(segment)
	}
	return sb.String()
}

func (r *Request) recordNode(n *Node, tokens []string, v any) {
	r.trace.record(r.at, Match{Node: nodeLabel(n), Tokens: tokens, Source: SourceNode, Value: v})
}

func (r *Request) recordSource(source string, v any) {
	r.trace.record(r.at, Match{Node: NoNode, Source: source, Value: v})
}

// recordFallback records a value produced by fakeBySchema and passes the result through.
func (r *Request) recordFallback(v any, err error) (any, error) {
	if err == nil {
		r.recordNode(g.root.defaultNode, nil, v)
	}
	return v, err
}

func nodeLabel(n *Node) string {
	if n == g.root.defaultNode {
		return n.Name
	}
	if path := NodePath(n); path != "" {
		return path
	}
	return n.Name
}
