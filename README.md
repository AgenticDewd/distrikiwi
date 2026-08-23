# DistriKiwi

DistriKiwi is a Go key-value store with an in-memory engine, a file-backed write-ahead log (WAL), and a gRPC API. Mutations are written to the WAL before they are applied to memory, allowing the in-memory state to be rebuilt after a restart.

## Current Features

- Thread-safe in-memory key-value storage
- `Put`, `Get`, and `Delete` gRPC operations
- Append-only binary WAL with `Sync()` after every write
- WAL replay during server startup
- Readable WAL inspection CLI
- Engine, concurrency, and crash-recovery tests

Replication is currently a placeholder for future development.

## Requirements

- Go 1.26 or newer

## Quick Start

Run the test suite:

```bash
go test ./...
```

Start the database server:

```bash
go run ./cmd/distrikiwidatabase
```

The server listens on `localhost:50051` and stores its WAL in `distrikiwi.wal` in the current working directory. On startup it reads that file and replays every valid operation into a new in-memory engine.

In another terminal, run the example gRPC client:

```bash
go run ./cmd/client
```

The client writes and reads `user:99`, then requests a missing key to demonstrate the `Found` response field.

## WAL Reader

Use the WAL reader to inspect transactions in a human-readable format:

```bash
go run ./cmd/walreader -wal ./distrikiwi.wal
```

The `-wal` flag defaults to `distrikiwi.wal`:

```bash
go run ./cmd/walreader
```

Example output:

```text
WAL: ./distrikiwi.wal
Entries: 2

[0001] PUT
  Key:   "user:99"
  Value: "Developer Bob" (13 bytes)

[0002] DELETE
  Key:   "user:99"
  Value: <empty> (0 bytes)
```

Printable UTF-8 values are displayed as quoted text. Binary or non-printable values are displayed as hexadecimal bytes. The reader opens the WAL read-only and does not create a missing file.

## Architecture

### Memory Engine

`internal/engine/MemEngine` stores values in a map protected by `sync.RWMutex`.

- Empty keys return `ErrKeyEmpty`.
- Missing keys return `ErrKeyNotFound` from engine `Get` and `Delete` calls.
- Values are copied when written and read, preventing callers from mutating stored data through a shared slice.

### Write-Ahead Log

`internal/wal/FileWAL` appends each mutation to disk and calls `Sync()` before returning from `WriteOp`.

Each record has this layout:

| Field | Size | Encoding |
| --- | ---: | --- |
| Operation type | 1 byte | `0` = PUT, `1` = DELETE |
| Key length | 4 bytes | Big-endian `uint32` |
| Value length | 4 bytes | Big-endian `uint32` |
| Key | N bytes | Raw UTF-8/string bytes |
| Value | M bytes | Raw bytes |

`ReadWAL()` reads records in append order. `ReadWALFile(path)` provides a read-only path-based reader for tools such as `cmd/walreader`.

### Recovery

The database startup sequence is:

1. Open or create `distrikiwi.wal`.
2. Read the WAL records.
3. Replay PUT and DELETE operations into a fresh memory engine.
4. Start the gRPC server.

For client mutations, the request follows the same durability order:

1. Append the operation to the WAL and sync it.
2. Apply the operation to the memory engine.
3. Return success to the client.

## gRPC API

The service definition is in [proto/distrikiwi.proto](proto/distrikiwi.proto).

| RPC | Request | Response |
| --- | --- | --- |
| `Put` | `key`, `value` | `success`, `message` |
| `Get` | `key` | `value`, `found` |
| `Delete` | `key` | `success`, `message` |

Empty keys are rejected with gRPC status code `InvalidArgument`. A missing key requested through `Get` returns a successful response with `found: false` and an empty value.

## Project Layout

```text
cmd/
  client/                 Example gRPC client
  distrikiwidatabase/    Database server and WAL recovery
  walreader/              Human-readable WAL inspection tool
internal/
  engine/                 In-memory engine and tests
  replication/            Replication placeholder
  server/                 gRPC service implementation
  wal/                    WAL interface, file implementation, and tests
proto/                    Protocol Buffer definition and generated Go code
```

## Tests

Run all tests with:

```bash
go test ./...
```

The tests cover basic engine behavior, concurrent access, WAL encoding/decoding, and crash recovery by replaying a WAL into a new engine.

## Roadmap

See [ROADMAP.md](ROADMAP.md) for planned improvements, semantic caching and replication ideas, and completed milestones. The roadmap tracks project direction rather than promising a specific release schedule.

## Known Limitations

- WAL corruption and truncated records are not reported with a dedicated corruption error yet.
- The server currently uses a fixed address, `localhost:50051`, and a fixed WAL path, `distrikiwi.wal`.
- Replication is not implemented yet.
- The generated Protocol Buffer files are checked into `proto/` and require the Protocol Buffer toolchain to regenerate after changing the `.proto` file.

## License

See [LICENSE](LICENSE).
