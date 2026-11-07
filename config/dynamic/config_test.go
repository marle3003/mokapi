package dynamic_test

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWrap(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "wrap simple",
			test: func(t *testing.T) {
				info := dynamictest.NewConfigInfo()
				config := &dynamic.Config{Info: dynamictest.NewConfigInfo()}
				inner := config.Info

				dynamic.Wrap(info, config)

				require.Equal(t, info.Key(), config.Info.Key())
				require.Equal(t, inner.Key(), config.Info.Inner().Key())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}

func TestListeners(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "listener add and invoke",
			test: func(t *testing.T) {
				l := dynamic.Listeners{}
				called := false
				l.Add("foo", func(event dynamic.ConfigEvent) {
					called = true
				})
				l.Invoke(dynamic.ConfigEvent{})
				require.True(t, called)
			},
		},
		{
			name: "removed listener should not be invoked",
			test: func(t *testing.T) {
				l := dynamic.Listeners{}
				called := false
				l.Add("foo", func(event dynamic.ConfigEvent) {
					called = true
				})
				l.Remove("foo")
				l.Invoke(dynamic.ConfigEvent{})
				require.False(t, called)
			},
		},
		{
			name: "remove on empty listeners does not panic",
			test: func(t *testing.T) {
				l := dynamic.Listeners{}
				require.NotPanics(t, func() { l.Remove("foo") })
			},
		},
		{
			name: "listener added while invoking is not called for that event",
			test: func(t *testing.T) {
				l := dynamic.Listeners{}
				late := false
				l.Add("first", func(event dynamic.ConfigEvent) {
					l.Add("late", func(event dynamic.ConfigEvent) { late = true })
				})
				l.Invoke(dynamic.ConfigEvent{})
				require.False(t, late)

				l.Invoke(dynamic.ConfigEvent{})
				require.True(t, late)
			},
		},
		{
			name: "list keeps insertion order",
			test: func(t *testing.T) {
				l := dynamic.Listeners{}
				var order []string
				for _, key := range []string{"a", "b", "c"} {
					l.Add(key, func(event dynamic.ConfigEvent) { order = append(order, key) })
				}
				for _, fn := range l.List() {
					fn(dynamic.ConfigEvent{})
				}
				require.Equal(t, []string{"a", "b", "c"}, order)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}

// Invoke and List must not walk the listener map without the lock while Add (e.g. from AddRef
// during parsing) modifies it.
func TestListeners_ConcurrentAccess(t *testing.T) {
	for _, name := range []string{"invoke", "list"} {
		t.Run(name+" while adding listeners", func(t *testing.T) {
			l := dynamic.Listeners{}
			l.Add("first", func(dynamic.ConfigEvent) {})

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 500; i++ {
					l.Add(strconv.Itoa(i), func(dynamic.ConfigEvent) {})
				}
			}()
			for i := 0; i < 500; i++ {
				if name == "invoke" {
					l.Invoke(dynamic.ConfigEvent{})
				} else {
					_ = l.List()
				}
			}
			wg.Wait()
		})
	}
}

func TestAddRef(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "reference is in list",
			test: func(t *testing.T) {
				parent := &dynamic.Config{Info: dynamictest.NewConfigInfo()}
				ref := &dynamic.Config{}

				dynamic.AddRef(parent, ref)

				require.Len(t, parent.Refs.List(true), 1)
				require.Equal(t, ref, parent.Refs.List(true)[0])
			},
		},
		{
			name: "ref updates parent time",
			test: func(t *testing.T) {
				parent := &dynamic.Config{Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("file://parent.yaml"))}
				child := &dynamic.Config{Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("file://ref.yaml"))}
				d, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
				child.Info.Time = d

				dynamic.AddRef(parent, child)
				child.Info.Checksum = []byte{1}
				child.Listeners.Invoke(dynamic.ConfigEvent{Config: child})

				require.Equal(t, d, parent.Info.Time)
			},
		},
		{
			name: "same ref is added only once",
			test: func(t *testing.T) {
				parent := &dynamic.Config{Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("file://parent.yaml"))}
				ref := &dynamic.Config{Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("file://ref.yaml"))}

				dynamic.AddRef(parent, ref)
				dynamic.AddRef(parent, ref)

				require.Len(t, parent.Refs.List(true), 1)
			},
		},
		{
			name: "add ref itself",
			test: func(t *testing.T) {
				parent := &dynamic.Config{Info: dynamictest.NewConfigInfo()}

				dynamic.AddRef(parent, parent)

				require.Len(t, parent.Refs.List(true), 0)
			},
		},
		{
			name: "add nested references and get all references",
			test: func(t *testing.T) {
				parent := &dynamic.Config{Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("file://parent.yaml"))}
				child := &dynamic.Config{Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("file://child.yaml"))}
				nested := &dynamic.Config{Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("file://nested.yaml"))}

				dynamic.AddRef(parent, child)
				dynamic.AddRef(child, nested)

				list := parent.Refs.List(true)
				require.Len(t, list, 2)
				require.Contains(t, list, child)
				require.Contains(t, list, nested)
			},
		},
		{
			name: "add nested references but get only first level",
			test: func(t *testing.T) {
				parent := &dynamic.Config{Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("file://parent.yaml"))}
				child := &dynamic.Config{Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("file://child.yaml"))}
				nested := &dynamic.Config{Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("file://nested.yaml"))}

				dynamic.AddRef(parent, child)
				dynamic.AddRef(child, nested)

				list := parent.Refs.List(false)
				require.Len(t, list, 1)
				require.Contains(t, list, child)
			},
		},
		{
			name: "ref is removed when deleted",
			test: func(t *testing.T) {
				parent := &dynamic.Config{Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("file://parent.yaml"))}
				child := &dynamic.Config{Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("file://child.yaml"))}

				dynamic.AddRef(parent, child)

				child.Listeners.Invoke(dynamic.ConfigEvent{Config: child, Event: dynamic.Delete})
				list := parent.Refs.List(false)
				require.Len(t, list, 0)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}

func TestValidate(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "validate is called",
			test: func(t *testing.T) {
				called := false
				v := &validatedData{
					validate: func() error {
						called = true
						return nil
					},
				}

				err := dynamic.Validate(&dynamic.Config{Data: v})

				require.True(t, called, "validate is called")
				require.NoError(t, err)
			},
		},
		{
			name: "validate returns error",
			test: func(t *testing.T) {
				v := &validatedData{
					validate: func() error {
						return fmt.Errorf("TEST ERROR")
					},
				}

				err := dynamic.Validate(&dynamic.Config{Data: v})

				require.Error(t, err)
				require.EqualError(t, err, "TEST ERROR")
			},
		},
		{
			name: "no error when validator is not implemented",
			test: func(t *testing.T) {
				v := struct {
				}{}

				err := dynamic.Validate(&dynamic.Config{Data: v})

				require.NoError(t, err)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}

func TestConfigScope(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "open and close top scope",
			test: func(t *testing.T) {
				c := &dynamic.Config{}
				c.OpenScope("foo")
				require.Equal(t, "foo", c.Scope.Name())
				c.CloseScope()
				// top scope is never closed
				require.Equal(t, "foo", c.Scope.Name())
			},
		},
		{
			name: "open and close sub scope",
			test: func(t *testing.T) {
				c := &dynamic.Config{}
				c.OpenScope("foo")
				c.OpenScope("bar")
				require.Equal(t, "bar", c.Scope.Name())
				c.CloseScope()
				require.Equal(t, "foo", c.Scope.Name())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}

func TestConfigEventText(t *testing.T) {
	require.Equal(t, "Create", dynamic.Create.String())
	require.Equal(t, "Update", dynamic.Update.String())
	require.Equal(t, "Delete", dynamic.Delete.String())
	require.Equal(t, "Chmod", dynamic.Chmod.String())
}

type validatedData struct {
	validate func() error
}

func (v *validatedData) Validate() error {
	return v.validate()
}
