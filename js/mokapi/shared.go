package mokapi

import (
	"mokapi/engine/common"
	"sync"
)

type SharedMemory struct {
	data  map[string]any
	clear func()
	m     sync.RWMutex
}

func NewSharedMemory(store common.Store) *SharedMemory {
	v := store.Get("shared-memory")
	var data map[string]any
	if v != nil {
		data = v.(map[string]any)
	} else {
		data = make(map[string]any)
		store.Set("shared-memory", data)
	}

	return &SharedMemory{data: data}
}

func (m *SharedMemory) Get(key string) any {
	m.m.RLock()
	defer m.m.RUnlock()

	return m.data[key]
}

func (m *SharedMemory) Has(key string) bool {
	m.m.RLock()
	defer m.m.RUnlock()

	_, b := m.data[key]
	return b
}

func (m *SharedMemory) Set(key string, value any) {
	m.m.Lock()
	defer m.m.Unlock()
	m.data[key] = value
}

func (m *SharedMemory) Delete(key string) {
	m.m.Lock()
	defer m.m.Unlock()
	delete(m.data, key)
}

func (m *SharedMemory) Clear() {
	m.m.Lock()
	defer m.m.Unlock()
	for k := range m.data {
		delete(m.data, k)
	}
}

func (m *SharedMemory) Update(key string, fn func(v any) any) any {
	m.m.Lock()
	defer m.m.Unlock()
	v := fn(m.data[key])
	m.data[key] = v
	return v
}

func (m *SharedMemory) Keys() []string {
	m.m.RLock()
	defer m.m.RUnlock()
	var keys []string
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

func (m *SharedMemory) Namespace(name string) *SharedMemory {
	m.m.Lock()
	v, ok := m.data[name]
	m.m.Unlock()
	if ok {
		ns, ok := v.(*SharedMemory)

		if !ok {
			return nil
		}
		return ns
	}

	m.m.Lock()
	defer m.m.Unlock()
	ns := &SharedMemory{data: make(map[string]any)}
	m.data[name] = ns
	return ns
}
