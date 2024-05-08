package dynamic_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"testing"
	"time"
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
				ref := &dynamic.Config{Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("file://ref.yaml"))}
				d, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
				ref.Info.Time = d

				dynamic.AddRef(parent, ref)
				ref.Listeners.Invoke(ref)

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

type validatedData struct {
	validate func() error
}

func (v *validatedData) Validate() error {
	return v.validate()
}
