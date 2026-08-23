# DistriKiwi Roadmap

This document records planned improvements and longer-term ideas for DistriKiwi. It describes direction, not a committed release schedule.

## Near-Term Improvements

- [ ] Add configurable server address and WAL path through command-line flags or configuration.
- [ ] Improve corrupted and truncated WAL recovery with explicit errors and safer recovery behavior.
- [ ] Expand the command-line tooling beyond WAL inspection.
- [ ] Add integration tests for the gRPC API and restart/recovery workflows.

## Distributed Storage

- [ ] Add replication for higher availability and distributed consistency.
- [ ] Define replication roles, membership, operation ordering, and failure recovery.

## Performance

- [ ] Add semantic caching to reduce repeated reads and improve response latency.
- [ ] Measure cache hit rate, stale-read behavior, invalidation, and memory usage before choosing a caching strategy.

## Completed Milestones

- [x] Implement a thread-safe in-memory key-value engine.
- [x] Add an append-only binary WAL with `Sync()` after every write.
- [x] Replay WAL operations during server startup.
- [x] Expose `Put`, `Get`, and `Delete` through gRPC.
- [x] Add a human-readable WAL inspection CLI.
