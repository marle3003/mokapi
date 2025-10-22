package mokapi

import (
	"mokapi/engine/common"
)

type SharedMemory struct {
	store common.Store
}

func NewSharedMemory(store common.Store) *SharedMemory {
	return &SharedMemory{store: store}
}

func (m *SharedMemory) Get(key string) any {
	return m.store.Get(key)
}

func (m *SharedMemory) Has(key string) bool {
	return m.store.Has(key)
}

func (m *SharedMemory) Set(key string, value any) {
	m.store.Set(key, value)
}

func (m *SharedMemory) Delete(key string) {
	m.store.Delete(key)
}

func (m *SharedMemory) Clear() {
	m.store.Clear()
}

func (m *SharedMemory) Update(key string, fn func(v any) any) any {
	return m.store.Update(key, fn)
}

func (m *SharedMemory) Keys() []string {
	return m.store.Keys()
}

func (m *SharedMemory) Namespace(name string) *SharedMemory {
	s := m.store.Namespace(name)
	return &SharedMemory{store: s}
}
