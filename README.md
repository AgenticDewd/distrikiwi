# DistriKiwi

DistriKiwi is a Go-based distributed key-value store project focused on an in-memory engine with durability via a write-ahead log (WAL).

The current implementation includes:
- A thread-safe in-memory key-value engine
- A file-backed WAL for crash recovery
- Tests for concurrency and WAL replay

## Why This Project

In-memory stores are fast but volatile. DistriKiwi explores a practical middle ground:
- Keep reads and writes in memory for speed
- Persist every mutation in a WAL for recovery after crashes
- Add replication next for high availability and distributed consistency

## Current Status

Implemented:
- `Put`, `Get`, and `Delete` operations in memory
- Concurrent access safety using `sync.RWMutex`
- Binary WAL format with append-only writes
- WAL replay to rebuild memory state after restart

## Project Structure

```text
internal/
	engine/
		engine.go         # Engine interface and common errors
		memengine.go      # In-memory implementation
		mem_engine_test.go
	wal/
		wal.go            # WAL interface and log entry definition
		file_wal.go       # File-backed WAL implementation
		file_wal_test.go
	replication/        # Placeholder for replication module
```

## Core Components

### 1) Memory Engine (`internal/engine`)

`MemEngine` stores keys in a `map[string][]byte` protected by an RW lock.

Behavior:
- Empty key returns `ErrKeyEmpty`
- Missing key on `Get`/`Delete` returns `ErrKeyNotFound`
- Values are copied on write and read to avoid external mutation

### 2) Write-Ahead Log (`internal/wal`)

`FileWAL` appends operations to disk and calls `Sync()` after each write.

Each log record is encoded as:
- `1 byte` operation type (`0=Put`, `1=Delete`)
- `4 bytes` key length (big-endian)
- `4 bytes` value length (big-endian)
- `N bytes` key
- `M bytes` value

On restart, `ReadWAL()` scans records in order and returns entries for replay.

## Recovery Model

Typical flow:
1. On write/delete request, append operation to WAL
2. Apply operation to in-memory engine
3. On crash/restart, create a fresh memory engine
4. Replay WAL entries in order to rebuild state

This flow is validated in `internal/wal/file_wal_test.go`.

## Getting Started

### Prerequisites

- Go `1.26+`

### Clone and Test

```bash
git clone https://github.com/harhitosw/distrikiwi.git
cd distrikiwi
go test ./...
```

## Usage Example

The project is currently a library-style codebase (no CLI/server yet). The following pattern mirrors how tests use the engine and WAL together:

```go
package main

import (
		"fmt"

		"github.com/harhitosw/distrikiwi/internal/engine"
		"github.com/harhitosw/distrikiwi/internal/wal"
)

func main() {
		eng := engine.NewMemEngine()
		w, err := wal.NewFileWAL("distrikiwi.wal")
		if err != nil {
				panic(err)
		}
		defer w.Close()

		// Put key
		if err := w.Write(0, "user:1", []byte("Alice")); err != nil {
				panic(err)
		}
		if err := eng.Put("user:1", []byte("Alice")); err != nil {
				panic(err)
		}

		// Read key
		val, err := eng.Get("user:1")
		if err != nil {
				panic(err)
		}
		fmt.Println(string(val)) // Alice
}
```

## Testing

Current tests cover:
- Basic engine operations
- Concurrent readers/writers on the memory engine
- Crash/recovery behavior using WAL replay

Run all tests:

```bash
go test ./...
```

## Known Gaps / Notes

- `internal/wal/wal.go` defines `WriteOp(...)`, while `FileWAL` currently exposes `Write(...)`.
- Truncated/corrupted WAL entry handling can be improved.

## Roadmap

1. Align WAL interface and implementation (`WriteOp` vs `Write`)
2. Add a server API (gRPC mostly for high performance)
3. Command line interface for easy usage.
4. Semantic Caching for Agents, reducing token costs.

## Contributing

Issues and pull requests are welcome. If you want to contribute, start with:
- WAL interface cleanup
- Semantic caching ideas
- API layer proposal
- Command Line interface

## License

Add a `LICENSE` file to define usage terms.
