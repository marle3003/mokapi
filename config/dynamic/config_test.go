package dynamic_test

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
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
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
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
