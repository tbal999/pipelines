package main

import "sync"

type MemoryStore struct {
	storage map[string]interface{}
	mu      *sync.Mutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		storage: make(map[string]interface{}),
		mu:      &sync.Mutex{},
	}
}

func (m *MemoryStore) Put(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.storage[key] = value
}

func (m *MemoryStore) Get(key string) interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.storage[key]
}
