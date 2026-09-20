package wal_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/AgenticDewd/distrikiwi/internal/engine"
	"github.com/AgenticDewd/distrikiwi/internal/wal"
)

func TestWAL_CrashAndRecovery(t *testing.T) {
	// 1. Setup a clean, temporary directory for our test WAL file.
	// t.TempDir() is a Go testing feature that automatically cleans up the file
	// from your Windows hard drive when the test finishes.
	tempDir := t.TempDir()
	walPath := filepath.Join(tempDir, "test_wal.log")

	// 2. Initialize our components
	dbEngine := engine.NewMemEngine()
	dbWAL, err := wal.NewFileWAL(walPath)
	if err != nil {
		t.Fatalf("failed to create WAL: %v", err)
	}

	// 3. Perform writes (simulating client traffic)
	// We must write to BOTH the WAL and the Memory Engine.
	writes := []struct {
		key string
		val []byte
	}{
		{"user:1", []byte("Alice")},
		{"user:2", []byte("Bob")},
		{"user:3", []byte("Charlie")},
	}

	for _, w := range writes {
		// Log the write first (WAL)
		if err := dbWAL.WriteOp(0, w.key, w.val); err != nil {
			t.Fatalf("failed to write to WAL: %v", err)
		}
		// Apply to memory engine
		if err := dbEngine.Put(w.key, w.val); err != nil {
			t.Fatalf("failed to write to memory engine: %v", err)
		}
	}

	// 4. Perform a deletion
	// Log the deletion (OpType = 1)
	if err := dbWAL.WriteOp(1, "user:2", nil); err != nil {
		t.Fatalf("failed to log delete to WAL: %v", err)
	}
	// Apply delete to memory engine
	if err := dbEngine.Delete("user:2"); err != nil {
		t.Fatalf("failed to delete from memory engine: %v", err)
	}

	// Close the WAL file to simulate the database server shutting down/crashing
	if err := dbWAL.Close(); err != nil {
		t.Fatalf("failed to close WAL: %v", err)
	}

	// ==========================================
	// SIMULATE CRASH: Destroy the memory engine
	// ==========================================
	dbEngine = nil

	// ==========================================
	// SIMULATE REBOOT: Recover from the WAL
	// ==========================================

	// Create a brand-new, completely empty memory engine
	recoveredEngine := engine.NewMemEngine()

	// Open the WAL in read mode
	recoveryWAL, err := wal.NewFileWAL(walPath)
	if err != nil {
		t.Fatalf("failed to open WAL for recovery: %v", err)
	}
	defer func() {
		if err := recoveryWAL.Close(); err != nil {
			t.Fatalf("failed to close recovery WAL: %v", err)
		}
	}()

	// Read all logged transactions sequentially
	logs, err := recoveryWAL.ReadWAL()
	if err != nil {
		t.Fatalf("failed to read WAL logs: %v", err)
	}

	// Replay each log entry into our empty recovered engine
	for _, log := range logs {
		switch log.OpType {
		case 0: // Put
			if err := recoveredEngine.Put(log.Key, log.Value); err != nil {
				t.Fatalf("recovery failed to replay PUT: %v", err)
			}
		case 1: // Delete
			if err := recoveredEngine.Delete(log.Key); err != nil {
				t.Fatalf("recovery failed to replay DELETE: %v", err)
			}
		default:
			t.Fatalf("unexpected op type during recovery: %d", log.OpType)
		}
	}

	// ==========================================
	// ASSERTIONS: Verify the recovered state
	// ==========================================

	// "user:1" should be "Alice"
	val1, err := recoveredEngine.Get("user:1")
	if err != nil {
		t.Errorf("expected user:1 to exist, got error: %v", err)
	}
	if !bytes.Equal(val1, []byte("Alice")) {
		t.Errorf("expected user:1 to be 'Alice', got '%s'", val1)
	}

	// "user:2" should have been deleted (should return ErrKeyNotFound)
	_, err = recoveredEngine.Get("user:2")
	if err != engine.ErrKeyNotFound {
		t.Errorf("expected user:2 to be deleted, got error: %v", err)
	}

	// "user:3" should be "Charlie"
	val3, err := recoveredEngine.Get("user:3")
	if err != nil {
		t.Errorf("expected user:3 to exist, got error: %v", err)
	}
	if !bytes.Equal(val3, []byte("Charlie")) {
		t.Errorf("expected user:3 to be 'Charlie', got '%s'", val3)
	}
}
