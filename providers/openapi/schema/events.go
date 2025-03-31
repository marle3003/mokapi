package schema

import "sync"

type changeManager struct {
	listeners []func(*Schema)
	mu        sync.Mutex
}

func (em *changeManager) Subscribe(callback func(ref *Schema)) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.listeners = append(em.listeners, callback)
}

// Notify triggers event listeners
func (em *changeManager) Notify(ref *Schema) {
	em.mu.Lock()
	defer em.mu.Unlock()
	for _, listener := range em.listeners {
		listener(ref)
	}
}
