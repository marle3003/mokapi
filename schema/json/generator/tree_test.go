package generator_test

import (
	"mokapi/schema/json/generator"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindByName(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "root",
			test: func(t *testing.T) {
				n := generator.FindByName("root")
				require.NotNil(t, n)
				require.Equal(t, "root", n.Name)
			},
		},
		{
			name: "street",
			test: func(t *testing.T) {
				n := generator.FindByName("street")
				require.NotNil(t, n)
				require.Equal(t, "street", n.Name)
				v, err := n.Fake(generator.NewRequest(nil, nil, nil))
				require.NoError(t, err)
				require.Equal(t, "295 Estatestown", v)
			},
		},
		{
			name: "absolute /street",
			test: func(t *testing.T) {
				n := generator.FindByName("/street")
				require.NotNil(t, n)
				require.Equal(t, "street", n.Name)
				v, err := n.Fake(generator.NewRequest(nil, nil, nil))
				require.NoError(t, err)
				require.Equal(t, "295 Estatestown", v)
			},
		},
		{
			name: "absolute /address/line1",
			test: func(t *testing.T) {
				n := generator.FindByName("/address/line1")
				require.NotNil(t, n)
				require.Equal(t, "line1", n.Name)
				v, err := n.Fake(generator.NewRequest(nil, nil, nil))
				require.NoError(t, err)
				require.Equal(t, "Grayson Anderson", v)
			},
		},
		{
			name: "absolute /address/line10",
			test: func(t *testing.T) {
				n := generator.FindByName("/house/number2")
				require.Nil(t, n)
			},
		},
		{
			name: "relative contact/phone",
			test: func(t *testing.T) {
				n := generator.FindByName("contact/phone")
				require.NotNil(t, n)
				require.Equal(t, "phone", n.Name)
				v, err := n.Fake(generator.NewRequest(nil, nil, nil))
				require.NoError(t, err)
				require.Equal(t, "+1062951049628", v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(1234578)
			tc.test(t)
		})
	}
}

func addSameNamedBranches(t *testing.T) (billing, shipping *generator.Node) {
	t.Helper()
	root := generator.FindByName(generator.RootName)
	billing = generator.NewNode("billing")
	billing.Append(generator.NewNode("code"))
	shipping = generator.NewNode("shipping")
	shipping.Append(generator.NewNode("code"))
	root.Append(billing)
	root.Append(shipping)
	t.Cleanup(func() {
		require.NoError(t, root.Remove("shipping"))
		require.NoError(t, root.Remove("billing"))
	})
	return
}

func TestFindByPath(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "slash is root",
			test: func(t *testing.T) {
				require.Same(t, generator.FindByName(generator.RootName), generator.FindByPath("/"))
			},
		},
		{
			name: "same name in two branches",
			test: func(t *testing.T) {
				billing, shipping := addSameNamedBranches(t)

				require.Same(t, billing.Children[0], generator.FindByPath("/billing/code"))
				require.Same(t, shipping.Children[0], generator.FindByPath("/shipping/code"))
			},
		},
		{
			name: "only direct children are searched",
			test: func(t *testing.T) {
				addSameNamedBranches(t)

				require.Nil(t, generator.FindByPath("/code"))
				require.Nil(t, generator.FindByPath("/billing/unknown"))
			},
		},
		{
			name: "three segments are exact",
			test: func(t *testing.T) {
				billing, _ := addSameNamedBranches(t)
				invoice := generator.NewNode("invoice")
				invoice.Append(generator.NewNode("code"))
				billing.Append(invoice)

				require.Same(t, invoice.Children[0], generator.FindByPath("/billing/invoice/code"))
				require.Nil(t, generator.FindByPath("/shipping/invoice/code"))
				require.Nil(t, generator.FindByPath("/anything/invoice/code"))
			},
		},
		{
			name: "unknown path",
			test: func(t *testing.T) {
				require.Nil(t, generator.FindByPath("/unknown"))
			},
		},
		{
			name: "path without leading slash",
			test: func(t *testing.T) {
				addSameNamedBranches(t)

				require.Nil(t, generator.FindByPath("billing/code"))
				require.Nil(t, generator.FindByPath(""))
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}

func TestNodePath(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "root",
			test: func(t *testing.T) {
				require.Equal(t, "/", generator.NodePath(generator.FindByName(generator.RootName)))
			},
		},
		{
			name: "nested nodes with the same name",
			test: func(t *testing.T) {
				billing, shipping := addSameNamedBranches(t)

				require.Equal(t, "/billing", generator.NodePath(billing))
				require.Equal(t, "/billing/code", generator.NodePath(billing.Children[0]))
				require.Equal(t, "/shipping/code", generator.NodePath(shipping.Children[0]))
			},
		},
		{
			name: "node outside of the tree",
			test: func(t *testing.T) {
				require.Equal(t, "", generator.NodePath(generator.NewNode("detached")))
			},
		},
		{
			name: "round trip with FindByPath",
			test: func(t *testing.T) {
				addSameNamedBranches(t)
				n := generator.FindByPath("/shipping/code")

				require.Same(t, n, generator.FindByPath(generator.NodePath(n)))
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}
