package engine

import (
	"sync"
	"testing"
)

func TestMemEngine_BasicOperations(t *testing.T) {
	me := NewMemEngine()
	err := me.Put("username", []byte("alice"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// 2. Test Get
	val, err := me.Get("username")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(val) != "alice" {
		t.Errorf("expected 'alice', got '%s'", val)
	}
}

func TestMemEngine_concurrentReadAndWrite(_ *testing.T) {
	me := NewMemEngine()
	var wg sync.WaitGroup
	numGoroutines := 100
	// concurrent writers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = me.Put("key", []byte("value"))
		}()
	}
	// Concurrent Readers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = me.Get("key")
		}()
	}
	wg.Wait()
}
