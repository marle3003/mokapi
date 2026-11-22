package generator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func appendToRoot(t *testing.T, n *Node) {
	t.Helper()
	g.root.Append(n)
	t.Cleanup(func() {
		require.NoError(t, g.root.Remove(n.Name))
	})
}

func leaf(name string, attributes ...string) *Node {
	return &Node{
		Name:       name,
		Attributes: attributes,
		Fake:       func(r *Request) (any, error) { return name, nil },
	}
}

func TestFindBestMatch_ConsumedTokens(t *testing.T) {
	testcases := []struct {
		name   string
		setup  func(t *testing.T)
		path   []string
		node   string
		tokens []string
	}{
		{
			name:   "leading tokens without a match are not consumed",
			path:   tokenize([]string{"customerAddress"}),
			node:   "address",
			tokens: []string{"address"},
		},
		{
			name:   "compound email address after a leading token",
			path:   tokenize([]string{"customerEmailAddress"}),
			node:   "email",
			tokens: []string{"emailaddress"},
		},
		{
			name:   "email address as attribute",
			path:   []string{"emailAddress"},
			node:   "email",
			tokens: []string{"emailAddress"},
		},
		{
			name:   "email address with underscore",
			path:   tokenize([]string{"email_address"}),
			node:   "email",
			tokens: []string{"emailaddress"},
		},
		{
			name:   "email address below contact",
			path:   []string{"contact", "emailAddress"},
			node:   "email",
			tokens: []string{"contact", "emailAddress"},
		},
		{
			name: "singular form of the first token",
			setup: func(t *testing.T) {
				invoice := leaf("invoice", "invoice")
				invoice.Append(leaf("line", "line"))
				invoice.Fake = nil
				appendToRoot(t, invoice)
			},
			path:   []string{"invoices", "line"},
			node:   "line",
			tokens: []string{"invoice", "line"},
		},
		{
			name: "merged path parts",
			setup: func(t *testing.T) {
				appendToRoot(t, leaf("orderStatus", "orderStatus"))
			},
			path:   []string{"order", "status"},
			node:   "orderStatus",
			tokens: []string{"orderstatus"},
		},
		{
			name: "wildcard attribute consumes the real token",
			setup: func(t *testing.T) {
				wrapper := &Node{Name: "wrapper", Attributes: []string{"wrapper"}}
				wrapper.Append(leaf("any", "*"))
				appendToRoot(t, wrapper)
			},
			path:   []string{"wrapper", "whatever"},
			node:   "any",
			tokens: []string{"wrapper", "whatever"},
		},
		{
			name:   "nested nodes consume one token per level",
			path:   []string{"customer", "address", "co"},
			node:   "co",
			tokens: []string{"address", "co"},
		},
		{
			name:   "no match",
			path:   []string{"zzzUnknownProperty"},
			node:   "fakeBySchema",
			tokens: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup(t)
			}
			r := NewRequest(tc.path, nil, nil)
			r.g = g

			n, tokens := findBestMatch(g.root, r)

			require.Equal(t, tc.node, n.Name)
			require.Equal(t, tc.tokens, tokens)
		})
	}
}

func TestResolve_StoresMatchOnFaker(t *testing.T) {
	r := NewRequest([]string{"customerEmailAddress"}, nil, nil)
	r.g = g

	f, err := resolve(r, true)

	require.NoError(t, err)
	require.Equal(t, "/email", NodePath(f.node))
	require.Equal(t, []string{"emailaddress"}, f.tokens)
}

func TestResolve_NodePathOfEmailAddress(t *testing.T) {
	testcases := []struct {
		path []string
		node string
	}{
		{path: []string{"emailAddress"}, node: "/email"},
		{path: []string{"contact", "emailAddress"}, node: "/contact/email"},
		{path: []string{"shippingAddress"}, node: "/address"},
		{path: []string{"billingAddress"}, node: "/address"},
		{path: []string{"customerAddress"}, node: "/address"},
	}

	for _, tc := range testcases {
		t.Run(strings.Join(tc.path, "."), func(t *testing.T) {
			r := NewRequest(tc.path, nil, nil)
			r.g = g

			f, err := resolve(r, true)

			require.NoError(t, err)
			require.Equal(t, tc.node, NodePath(f.node))
		})
	}
}
