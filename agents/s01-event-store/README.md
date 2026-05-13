# s01 — Event store

A tiny append-only event store for one conversation, modelled on
`openhands/app_server/event/filesystem_event_service.py`.

## Run

```sh
go test ./...        # 7 tests, all green
go run ./cmd/demo    # scripted 6-turn conversation; replay + filter
```

## What this teaches

- **Discriminated union events.** One `Event` shape with `kind` +
  `payload`, three concrete payload types (`Message`, `Action`,
  `Observation`). This is the *only* shape that flows through the rest
  of OpenHands' server.
- **Per-conversation directory.** `<root>/<conv-uuid>/<event-uuid>.json`.
  Trivial to inspect; trivial to copy off-box for postmortems.
- **`Store` interface, filesystem implementation.** Later chapters add
  AWS / GCS backends (see `upstream-readings/`); none touch callers.
- **Zero dependencies.** s01 ships its own UUIDv4 so a learner can audit
  the entire module in one `go doc` pass.

## Files

| File | What | Lines |
|------|------|------|
| `event.go` | `Event` struct + payload variants + decoders | ~115 |
| `store.go` | `Store` interface + `FilesystemStore` + `Filter` | ~165 |
| `uuid.go` | RFC 4122 v4 UUID with JSON round-trip | ~75 |
| `store_test.go` | round-trip / search / count / limit / mismatch | ~135 |
| `cmd/demo/main.go` | end-to-end 6-turn conversation replay | ~75 |

## Six-section spine

Full mental model, ASCII diagram, diff-from-previous, hands-on, and
upstream source reading live in
[`docs/en/s01-event-store.md`](../../docs/en/s01-event-store.md) /
[`docs/zh/s01-event-store.md`](../../docs/zh/s01-event-store.md).
