package dynamic_test

import (
	"mokapi/config/dynamic"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAddRef_Listener(t *testing.T) {
	setup := func() (parent, ref *dynamic.Config, events *[]dynamic.ConfigEvent) {
		parent = &dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{Scheme: "foo", Host: "parent.yml"}}}
		ref = &dynamic.Config{Info: dynamic.ConfigInfo{Url: &url.URL{Scheme: "foo", Host: "ref.yml"}, Checksum: []byte{1}}}
		events = &[]dynamic.ConfigEvent{}
		parent.Listeners.Add("test", func(e dynamic.ConfigEvent) { *events = append(*events, e) })
		dynamic.AddRef(parent, ref)
		return
	}
	invoke := func(ref *dynamic.Config, evt dynamic.Event, checksum byte, at time.Time) {
		for _, l := range ref.Listeners.List() {
			l(dynamic.ConfigEvent{Event: evt, Config: ref, Checksum: []byte{checksum}, Time: at})
		}
	}

	t.Run("first Create with the checksum read at AddRef does not reload the parent", func(t *testing.T) {
		_, ref, events := setup()

		invoke(ref, dynamic.Create, 1, time.Now())

		require.Empty(t, *events)
	})

	t.Run("Create with a different checksum reloads the parent", func(t *testing.T) {
		_, ref, events := setup()

		invoke(ref, dynamic.Create, 2, time.Now())

		require.Len(t, *events, 1)
	})

	t.Run("Update with the checksum read at AddRef reloads the parent", func(t *testing.T) {
		_, ref, events := setup()

		invoke(ref, dynamic.Update, 1, time.Now())

		require.Len(t, *events, 1)
	})

	t.Run("parent event carries the time of the reference", func(t *testing.T) {
		parent, ref, events := setup()
		at := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)

		invoke(ref, dynamic.Update, 2, at)

		require.Len(t, *events, 1)
		require.Equal(t, dynamic.Update, (*events)[0].Event)
		require.Same(t, parent, (*events)[0].Config)
		require.Equal(t, at, (*events)[0].Time)
		require.True(t, parent.Info.Time.IsZero(), "listener must not write the parent's info")
	})

	t.Run("Delete of the reference removes it from the parent", func(t *testing.T) {
		parent, ref, events := setup()
		require.Len(t, parent.Refs.List(false), 1)

		invoke(ref, dynamic.Delete, 1, time.Time{})

		require.Len(t, *events, 1)
		require.Empty(t, parent.Refs.List(false))
	})
}
