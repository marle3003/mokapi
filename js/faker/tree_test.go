package faker_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/enginetest"
	"mokapi/js"
	"mokapi/js/eventloop"
	"mokapi/js/faker"
	"mokapi/js/require"
	"mokapi/schema/json/generator"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/dop251/goja"
	r "github.com/stretchr/testify/require"
)

func TestTree(t *testing.T) {
	cleanup := func(host *enginetest.Host) {
		for index := len(host.CleanupFuncs) - 1; index >= 0; index-- {
			host.CleanupFuncs[index]()
		}
	}
	getNames := func(children []*generator.Node) []string {
		names := make([]string, len(children))
		for i, child := range children {
			if child != nil {
				names[i] = child.Name
			} else {
				names[i] = ""
			}
		}
		return names
	}

	testcases := []struct {
		name string
		test func(t *testing.T, vm *goja.Runtime, host *enginetest.Host)
	}{
		{
			name: "findByName nil",
			test: func(t *testing.T, vm *goja.Runtime, _ *enginetest.Host) {
				v, err := vm.RunString(`
					const m = require('faker');
					m.findByName('foo');
				`)
				r.NoError(t, err)
				r.Nil(t, v.Export())
			},
		},
		{
			name: "findByName root",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return generator.NewNode("root")
					}
					return nil
				}

				v, err := vm.RunString(`
					const m = require('faker');
					m.findByName('root');
				`)
				r.NoError(t, err)
				r.NotNil(t, v.Export())
				r.Len(t, host.CleanupFuncs, 1)
			},
		},
		{
			name: "get name",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return generator.NewNode("root")
					}
					return nil
				}

				v, err := vm.RunString(`
					const m = require('faker');
					const root = m.findByName('root');
					root.name
				`)
				r.NoError(t, err)
				r.Equal(t, "root", v.Export())
			},
		},
		{
			name: "set name",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				root := generator.NewNode("root")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return root
					}
					return nil
				}

				_, err := vm.RunString(`
					const m = require('faker');
					const root = m.findByName('root');
					root.name = 'foo';
				`)
				r.NoError(t, err)
				r.Equal(t, "foo", root.Name)
			},
		},
		{
			name: "restore name",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				root := generator.NewNode("root")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return root
					}
					return nil
				}

				_, err := vm.RunString(`
					const m = require('faker');
					const root = m.findByName('root');
					root.name = 'foo';
					root.restore();
				`)
				r.NoError(t, err)
				r.Equal(t, "root", root.Name)
			},
		},
		{
			name: "empty children",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				root := generator.NewNode("root")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return root
					}
					return nil
				}

				v, err := vm.RunString(`
					const m = require('faker');
					const root = m.findByName('root');
					root.children.length;
				`)
				r.NoError(t, err)
				r.Equal(t, int64(0), v.Export())
			},
		},
		{
			name: "push new node",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				root := generator.NewNode("root")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return root
					}
					return nil
				}

				_, err := vm.RunString(`
					const m = require('faker');
					const root = m.findByName('root');
					root.children.push({
						name: 'foo',
					});
				`)
				r.NoError(t, err)
				r.Len(t, root.Children, 1)
			},
		},
		{
			name: "restore push new node",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				root := generator.NewNode("root")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return root
					}
					return nil
				}

				_, err := vm.RunString(`
					const m = require('faker');
					const root = m.findByName('root');
					root.children.push({
						name: 'foo',
					});
				`)
				r.NoError(t, err)
				r.Len(t, root.Children, 1)
				cleanup(host)
				r.Len(t, root.Children, 0)
			},
		},
		{
			name: "restore push two new nodes",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				root := generator.NewNode("root")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return root
					}
					return nil
				}

				_, err := vm.RunString(`
					const m = require('faker');
					const root = m.findByName('root');
					root.children.push({ name: 'foo' });
					root.children.push({ name: 'bar' });
				`)
				r.NoError(t, err)
				r.Len(t, root.Children, 2)
				cleanup(host)
				r.Len(t, root.Children, 0)
			},
		},
		{
			name: "restore push, unshift and push",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				root := generator.NewNode("root")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return root
					}
					return nil
				}

				_, err := vm.RunString(`
					const m = require('faker');
					const root = m.findByName('root');
					root.children.push({ name: 'foo' });
					root.children.unshift({ name: 'bar' });
					root.children.push({ name: 'yuh' });
				`)
				r.NoError(t, err)
				r.Equal(t, []string{"bar", "foo", "yuh"}, getNames(root.Children))
				cleanup(host)
				r.Len(t, root.Children, 0)
			},
		},
		{
			name: "restore push and fake override",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				root := generator.NewNode("root")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return root
					}
					return nil
				}

				_, err := vm.RunString(`
					const m = require('faker');
					const root = m.findByName('root');
					root.children.push({ name: 'foo' });
					root.fake = () => 'bar';
					root.children.push({ name: 'yuh' });
				`)
				r.NoError(t, err)
				r.NotNil(t, root.Fake)
				r.True(t, root.Custom)
				cleanup(host)
				r.Len(t, root.Children, 0)
				r.Nil(t, root.Fake)
				r.False(t, root.Custom)
			},
		},
		{
			name: "set children",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				root := generator.NewNode("root")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return root
					}
					return nil
				}

				_, err := vm.RunString(`
					const m = require('faker');
					const root = m.findByName('root');
					root.children = [{
						name: 'foo',
					}];
				`)
				r.NoError(t, err)
				r.Equal(t, []string{"foo"}, getNames(root.Children))

				cleanup(host)
				r.Equal(t, []string{}, getNames(root.Children))
			},
		},
		{
			name: "access and set children",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				root := generator.NewNode("root")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return root
					}
					return nil
				}

				v, err := vm.RunString(`
					const m = require('faker');
					const root = m.findByName('root');
					const children = root.children
					root.children = [{
						name: 'foo',
					}];
					children[0].name
				`)
				r.NoError(t, err)
				r.Equal(t, "foo", v.Export())
				r.Equal(t, []string{"foo"}, getNames(root.Children))

				cleanup(host)
				r.Equal(t, []string{}, getNames(root.Children))
			},
		},
		{
			name: "set child using index operator",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				root := generator.NewNode("root")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return root
					}
					return nil
				}

				_, err := vm.RunString(`
					const m = require('faker');
					const root = m.findByName('root');
					root.children[0] = {
						name: 'foo',
					};
				`)
				r.NoError(t, err)
				r.Equal(t, []string{"foo"}, getNames(root.Children))

				cleanup(host)
				r.Equal(t, []string{}, getNames(root.Children))
			},
		},
		{
			name: "set child using index operator with children",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				root := generator.NewNode("root")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return root
					}
					return nil
				}

				_, err := vm.RunString(`
					const m = require('faker');
					const root = m.findByName('root');
					root.children[0] = {
						name: 'foo',
						children: [{ name: 'bar' }]
					};
				`)
				r.NoError(t, err)
				r.Equal(t, []string{"foo"}, getNames(root.Children))
				r.Equal(t, []string{"bar"}, getNames(root.Children[0].Children))

				cleanup(host)
				r.Equal(t, []string{}, getNames(root.Children))
			},
		},
		{
			name: "pop",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				root := generator.NewNode("root")
				root.Children = []*generator.Node{generator.NewNode("child1"), generator.NewNode("child2")}
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return root
					}
					return nil
				}

				_, err := vm.RunString(`
					const m = require('faker');
					const root = m.findByName('root');
					root.children.pop();
				`)
				r.NoError(t, err)
				r.Equal(t, []string{"child1"}, getNames(root.Children))

				cleanup(host)
				r.Equal(t, []string{"child1", "child2"}, getNames(root.Children))
			},
		},
		{
			name: "shift",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				root := generator.NewNode("root")
				root.Children = []*generator.Node{generator.NewNode("child1"), generator.NewNode("child2")}
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return root
					}
					return nil
				}

				_, err := vm.RunString(`
					const m = require('faker');
					const root = m.findByName('root');
					root.children.shift();
				`)
				r.NoError(t, err)
				r.Equal(t, "child2", root.Children[0].Name)

				cleanup(host)
				r.Equal(t, []string{"child1", "child2"}, getNames(root.Children))
			},
		},
		{
			name: "unshift",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				root := generator.NewNode("root")
				root.Children = append(root.Children, generator.NewNode("child3"))
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return root
					}
					return nil
				}

				_, err := vm.RunString(`
					const m = require('faker');
					const root = m.findByName('root');
					root.children.unshift({ name: 'child1' }, { name: 'child2' });
				`)
				r.NoError(t, err)
				r.Equal(t, []string{"child1", "child2", "child3"}, getNames(root.Children))

				cleanup(host)
				r.Equal(t, []string{"child3"}, getNames(root.Children))
			},
		},
		{
			name: "set child at pos 3",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				root := generator.NewNode("root")
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return root
					}
					return nil
				}

				_, err := vm.RunString(`
					const m = require('faker');
					const root = m.findByName('root');
					root.children[2] = {
						name: 'foo',
					};
				`)
				r.NoError(t, err)
				r.Equal(t, []string{"", "", "foo"}, getNames(root.Children))

				cleanup(host)
				r.Equal(t, []string{}, getNames(root.Children))
			},
		},
		{
			name: "splice",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				root := generator.NewNode("root")
				root.Children = append(root.Children, generator.NewNode("child1"), generator.NewNode("child2"), generator.NewNode("child3"))
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return root
					}
					return nil
				}

				_, err := vm.RunString(`
					const m = require('faker');
					const root = m.findByName('root');
					root.children.splice(1, 1)
				`)
				r.NoError(t, err)
				r.Len(t, root.Children, 2)

				cleanup(host)
				r.Equal(t, []string{"child1", "child2", "child3"}, getNames(root.Children))
			},
		},
		{
			name: "insert with splice",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				root := generator.NewNode("root")
				root.Children = append(root.Children, generator.NewNode("child1"), generator.NewNode("child3"), generator.NewNode("child4"))
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return root
					}
					return nil
				}

				_, err := vm.RunString(`
					const m = require('faker');
					const root = m.findByName('root');
					root.children.splice(1, 0, { name: 'child2' })
				`)
				r.NoError(t, err)
				r.Equal(t, []string{"child1", "child2", "child3", "child4"}, getNames(root.Children))

				cleanup(host)
				r.Equal(t, []string{"child1", "child3", "child4"}, getNames(root.Children))
			},
		},
		{
			name: "delete and insert with splice",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				root := generator.NewNode("root")
				root.Children = append(root.Children, generator.NewNode("child1"), generator.NewNode("child3"), generator.NewNode("child4"))
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return root
					}
					return nil
				}

				_, err := vm.RunString(`
					const m = require('faker');
					const root = m.findByName('root');
					root.children.splice(1, 1, { name: 'child2' })
				`)
				r.NoError(t, err)
				r.Equal(t, []string{"child1", "child2", "child4"}, getNames(root.Children))

				cleanup(host)
				r.Equal(t, []string{"child1", "child3", "child4"}, getNames(root.Children))
			},
		},
		{
			name: "native find",
			test: func(t *testing.T, vm *goja.Runtime, host *enginetest.Host) {
				root := generator.NewNode("root")
				root.Children = append(root.Children, generator.NewNode("child1"), generator.NewNode("child2"), generator.NewNode("child3"))
				host.FindFakerNodeFunc = func(name string) *generator.Node {
					if name == "root" {
						return root
					}
					return nil
				}

				v, err := vm.RunString(`
					const m = require('faker');
					const root = m.findByName('root');
					root.children.find( (x) => x.name === 'child2')
				`)
				r.NoError(t, err)

				n := v.Export().(*generator.Node)
				r.Equal(t, "child2", n.Name)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(11)

			reg, err := require.NewRegistry(nil)
			reg.RegisterNativeModule("faker", faker.Require)
			r.NoError(t, err)

			vm := goja.New()
			vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
			host := &enginetest.Host{}
			loop := eventloop.New(vm, host)
			defer loop.Stop()
			loop.StartLoop()
			js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
			reg.Enable(vm)

			tc.test(t, vm, host)
		})
	}
}

// A custom node outlives its script when the loop stops before the node is removed.
// Faking with it must fail instead of waiting for a loop that never answers.
func TestTree_CustomNodeOfStoppedLoop(t *testing.T) {
	reg, err := require.NewRegistry(nil)
	r.NoError(t, err)
	reg.RegisterNativeModule("faker", faker.Require)

	vm := goja.New()
	host := &enginetest.Host{}
	root := generator.NewNode("root")
	host.FindFakerNodeFunc = func(name string) *generator.Node { return root }
	loop := eventloop.New(vm, host)
	loop.StartLoop()
	defer loop.Stop()
	js.EnableInternal(vm, host, loop, &dynamic.Config{Info: dynamictest.NewConfigInfo()})
	reg.Enable(vm)

	_, err = loop.RunSync(func(vm *goja.Runtime) (goja.Value, error) {
		return vm.RunString(`
			const m = require('faker');
			m.findByName('root').children.push({ name: 'foo', fake: () => 'bar' });
		`)
	})
	r.NoError(t, err)
	r.Len(t, root.Children, 1)

	loop.Stop()
	r.Eventually(t, func() bool {
		_, err := loop.RunSync(func(*goja.Runtime) (goja.Value, error) { return nil, nil })
		return err != nil
	}, time.Second, time.Millisecond)

	_, err = root.Children[0].Fake(&generator.Request{Context: &generator.Context{}})
	r.ErrorIs(t, err, eventloop.ErrStopped)
}
