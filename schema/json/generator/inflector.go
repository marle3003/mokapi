package generator

import (
	"container/list"
	"strings"
	"sync"

	"github.com/jinzhu/inflection"
)

type entry struct {
	key, value string
	prev, next *entry
}

type cachedInflector struct {
	mu         sync.Mutex
	capacity   int
	cache      map[string]*entry
	evictList  *list.List
	head, tail *entry
}

func newCachedInflector(capacity int) *cachedInflector {
	head := &entry{}
	tail := &entry{}
	head.next = tail
	tail.prev = head

	return &cachedInflector{
		capacity: capacity,
		cache:    make(map[string]*entry, capacity),
		head:     head,
		tail:     tail,
	}
}

func (m *cachedInflector) Singular(token string) string {
	m.mu.Lock()
	if e, found := m.cache[token]; found {
		m.moveToFront(e)
		v := e.value
		m.mu.Unlock()
		return v
	}
	m.mu.Unlock()

	// Expensive computation done without holding the lock.
	res := inflection.Singular(token)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Another goroutine may have inserted it while we were unlocked.
	if e, found := m.cache[token]; found {
		return e.value
	}

	// Detach the cached copies from any larger source string the
	// caller's token/res might be sliced from, so we don't pin
	// unrelated memory for the lifetime of the cache.
	key := strings.Clone(token)

	var value string
	if res == token {
		// Word is already singular: reuse the same backing array
		// instead of retaining a second copy of identical bytes.
		value = key
	} else {
		value = strings.Clone(res)
	}

	e := &entry{key: key, value: value}
	m.pushFront(e)
	m.cache[key] = e

	if len(m.cache) > m.capacity {
		m.evictOldest()
	}

	return value
}

func (m *cachedInflector) pushFront(e *entry) {
	e.next = m.head.next
	e.prev = m.head
	m.head.next.prev = e
	m.head.next = e
}

func (m *cachedInflector) remove(e *entry) {
	e.prev.next = e.next
	e.next.prev = e.prev
}

func (m *cachedInflector) moveToFront(e *entry) {
	m.remove(e)
	m.pushFront(e)
}

func (m *cachedInflector) evictOldest() {
	oldest := m.tail.prev
	if oldest == m.head {
		return
	}
	m.remove(oldest)
	delete(m.cache, oldest.key)
}
