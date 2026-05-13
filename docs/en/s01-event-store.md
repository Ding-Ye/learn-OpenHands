# s01 — Event store

> **Upstream:** `openhands/app_server/event/event_service.py` and
> `openhands/app_server/event/filesystem_event_service.py` at SHA
> [`a89778f3`](https://github.com/OpenHands/OpenHands/tree/a89778f3d7036b8d81d57a1f93e31c6df8219eff/openhands/app_server/event).

## Problem

OpenHands' server is *agentic* — long conversations between a user and
an autonomous agent, where the agent issues actions, the sandbox
returns observations, the agent thinks again. Three things have to be
true about whatever sits at the bottom of that stack:

1. **Replayability.** When live status reconnects, when a postmortem
   needs to know what happened, when a callback wakes up late — they
   all want to re-read the conversation from t=0.
2. **No type erasure.** A user message and a sandbox observation are
   not interchangeable. The store has to remember which it was.
3. **No "current state" trap.** If the only source of truth is the
   in-memory agent, a process restart loses everything. So the store
   is the source of truth, and the agent is just a fold over it.

Upstream's answer is the **`EventService`** abstraction: an append-only,
per-conversation log of discriminated-union events, with one of four
concrete backends (filesystem / AWS / GCS / SQLite-backed). s01 ports
the filesystem variant to Go in ~400 lines.

## Solution

```go
type Store interface {
    Save(conv UUID, e Event) error
    Get(conv, id UUID) (Event, error)
    Search(conv UUID, f Filter) ([]Event, error)
    Count(conv UUID, f Filter) (int, error)
}
```

Plus three payload types — `Message`, `Action`, `Observation` — and
one `Event` struct that wraps them by `Kind`. The filesystem
implementation writes one JSON file per event under
`<root>/<conversation>/<event-id>.json`.

## How it works

```
            ┌──────────── one conversation ─────────────┐
            │                                            │
NewMessage ─▶ Event{Kind=message, Payload=…}              │
  Save      │                                            │
            ▼                                            │
       .../conv-abc/                                     │
         ├── ev-001.json   (message  20:08:56.214)       │
         ├── ev-002.json   (action   20:08:56.215)       │
         ├── ev-003.json   (observ.  20:08:56.215)       │
         ├── ev-004.json   (action   20:08:56.216)       │
         ├── ev-005.json   (observ.  20:08:56.216)       │
         └── ev-006.json   (message  20:08:56.216)       │
                                                          │
Search(Filter{Kind:KindAction, SortAsc:true})  ──────────┤
  ↓                                                       │
  walks dir, decodes, filters, sorts, truncates           │
  ↓                                                       │
  []Event  ←──  caller switches on .Kind and decodes      │
            └────────────────────────────────────────────┘
```

Key choices, with their upstream counterparts:

- **Filename is the id, payload carries the timestamp.** Mirrors
  upstream's `_load_event` (read timestamp from JSON, not from the
  filename). Makes the store tolerant of clock skew across writers.
- **Atomic-ish save via `.tmp` + rename.** s01's small upgrade over
  upstream's plain `path.write_text` — free durability under SIGKILL.
- **Filter is a value, not a builder.** The zero value matches every
  event. Same affordance as upstream's optional-kwargs search.
- **Sort happens at read time.** No index, no compaction. Works up to
  a few thousand events per conversation; appendix-b explains where
  upstream switches to SQL.

## What changed

This is the first chapter. The diff is against an empty workspace.
The single load-bearing decision is "what shape does the event take" —
once that's locked, every later chapter (callbacks, status, hooks)
gets to assume `Store` exists and consume from it.

## Try it

```sh
cd agents/s01-event-store
go test ./...         # 7 tests
go run ./cmd/demo     # 6-turn scripted conversation; prints replay
```

Inspect the files the demo wrote:

```sh
DIR=$(go run ./cmd/demo 2>&1 | awk '/store root:/{print $3}')
ls "$DIR"/*/
jq . "$DIR"/*/*.json | head -50
```

Each event is human-readable; that's deliberate. The upstream server's
operator tooling (`oh logs`, `oh export-conversation`) is mostly
"`cat` the directory tree".

## Upstream source reading

Open [`upstream-readings/s01-event-store.py`](../../upstream-readings/s01-event-store.py).
It excerpts `event_service.py` (the abstract base) and
`filesystem_event_service.py` (the only backend that runs without a
cloud account), with comments showing exactly which Go construct
corresponds to which Python method.

Then read the upstream files in full:

- [`openhands/app_server/event/event_service.py`](https://github.com/OpenHands/OpenHands/blob/a89778f3d7036b8d81d57a1f93e31c6df8219eff/openhands/app_server/event/event_service.py)
- [`openhands/app_server/event/filesystem_event_service.py`](https://github.com/OpenHands/OpenHands/blob/a89778f3d7036b8d81d57a1f93e31c6df8219eff/openhands/app_server/event/filesystem_event_service.py)
- [`openhands/app_server/event/event_service_base.py`](https://github.com/OpenHands/OpenHands/blob/a89778f3d7036b8d81d57a1f93e31c6df8219eff/openhands/app_server/event/event_service_base.py)

Once you're comfortable with the abstract base + filesystem backend,
glance at `aws_event_service.py` and `google_cloud_event_service.py` —
they swap the storage primitive but keep the four-method shape
intact. That's the discipline s02 inherits.
