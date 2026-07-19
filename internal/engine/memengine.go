package engine

import (
	"sync"
)

// define a struct which is our memory engine object
type MemEngine struct {
	mu   sync.RWMutex
	data map[string][]byte
}

// create function to return a new instance of memory engine
func NewMemEngine() *MemEngine {
	return &MemEngine{
		data: make(map[string][]byte),
	}
}

// define the GET function
func (m *MemEngine) Get(key string) ([]byte, error) {
	// if key is empty return error
	if key == "" {
		return nil, ErrKeyEmpty
	}
	m.mu.RLock()
	// clear this lock after this function is done executing
	defer m.mu.RUnlock()
	val, exists := m.data[key]
	if !exists {
		return nil, ErrKeyNotFound
	}
	// return the copy of the value not the pointer so that value stays the same
	valCopy := make([]byte, len(val))
	copy(valCopy, val)
	return valCopy, nil
}

func (m *MemEngine) Put(key string, value []byte) error {
	if key == "" {
		return ErrKeyEmpty
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	valCopy := make([]byte, len(value))
	copy(valCopy, value)
	m.data[key] = valCopy
	return nil
}

// delete the given key from the memory engine
func (m *MemEngine) Delete(key string) error {
	if key == "" {
		return ErrKeyEmpty
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	_, exists := m.data[key]
	if !exists {
		return ErrKeyNotFound
	}
	delete(m.data, key)
	return nil
}

func (m *MemEngine) Close() error {
	return nil
}
