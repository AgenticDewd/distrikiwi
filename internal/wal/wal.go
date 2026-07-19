package wal

// define entry
type LogEntry struct {
	OpType uint8 // 0 - Put , 1 - Delete
	Key string
	Value [] byte
}

type WAL interface {
	WriteOp (opType uint8, key string, value []byte) error
	ReadWAL() ([]LogEntry, error)
	Close() error
}