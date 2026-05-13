package eventstore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Store is the abstract shape every OpenHands event service must
// implement. Upstream defines four backends (filesystem, AWS, GCS, plus
// the abstract base) — s01 ships only the filesystem one. Later chapters
// (s04 callbacks, s10 live-status) talk to Store, not to its concrete
// type.
type Store interface {
	Save(conv UUID, e Event) error
	Get(conv, id UUID) (Event, error)
	Search(conv UUID, f Filter) ([]Event, error)
	Count(conv UUID, f Filter) (int, error)
}

// Filter mirrors the search parameters of upstream's EventService.search.
// All fields are optional; the zero value matches every event in the
// conversation. SortAsc=true gives oldest-first (the default upstream
// behaviour); false gives newest-first.
type Filter struct {
	Kind         Kind
	SinceInclude time.Time
	UntilExclude time.Time
	Limit        int
	SortAsc      bool
}

// FilesystemStore stores each event as one JSON file under
// <root>/<conversation>/<event-id>.json. Sort order is recovered from the
// `timestamp` field at read time — the filename is just the id, so a
// reader doesn't need to trust the writer's clock.
type FilesystemStore struct {
	Root string
}

// NewFilesystemStore creates the root directory if missing and returns a
// store rooted at it.
func NewFilesystemStore(root string) (*FilesystemStore, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir root: %w", err)
	}
	return &FilesystemStore{Root: root}, nil
}

func (s *FilesystemStore) convDir(conv UUID) string {
	return filepath.Join(s.Root, conv.String())
}

func (s *FilesystemStore) eventPath(conv, id UUID) string {
	return filepath.Join(s.convDir(conv), id.String()+".json")
}

// Save serialises the event and writes it to disk. Atomic-ish: we write
// to a `.tmp` sibling and rename. (Real durability — fsync, crash-safe —
// lands when s09 introduces a write-ahead log.)
func (s *FilesystemStore) Save(conv UUID, e Event) error {
	if e.ConversationID != conv {
		return fmt.Errorf("event.ConversationID=%s != conv=%s", e.ConversationID, conv)
	}
	if err := os.MkdirAll(s.convDir(conv), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}
	final := s.eventPath(conv, e.ID)
	tmp := final + ".tmp"
	if err := os.WriteFile(tmp, body, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, final)
}

// Get reads one event by id. Returns os.ErrNotExist if the conversation
// or event is unknown.
func (s *FilesystemStore) Get(conv, id UUID) (Event, error) {
	body, err := os.ReadFile(s.eventPath(conv, id))
	if err != nil {
		return Event{}, err
	}
	var e Event
	if err := json.Unmarshal(body, &e); err != nil {
		return Event{}, fmt.Errorf("decode %s: %w", id, err)
	}
	return e, nil
}

// Search walks the conversation directory, decodes every event, applies
// the filter, sorts, and truncates to Limit. For s01's "small N per
// conversation" assumption that's fine; production replaces this with a
// SQL or RocksDB-backed index (see appendix-b for the upstream's path).
func (s *FilesystemStore) Search(conv UUID, f Filter) ([]Event, error) {
	entries, err := os.ReadDir(s.convDir(conv))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var hits []Event
	for _, ent := range entries {
		if ent.IsDir() || filepath.Ext(ent.Name()) != ".json" {
			continue
		}
		body, err := os.ReadFile(filepath.Join(s.convDir(conv), ent.Name()))
		if err != nil {
			return nil, err
		}
		var e Event
		if err := json.Unmarshal(body, &e); err != nil {
			return nil, fmt.Errorf("decode %s: %w", ent.Name(), err)
		}
		if !f.matches(e) {
			continue
		}
		hits = append(hits, e)
	}

	sort.Slice(hits, func(i, j int) bool {
		if f.SortAsc {
			return hits[i].Timestamp.Before(hits[j].Timestamp)
		}
		return hits[i].Timestamp.After(hits[j].Timestamp)
	})

	if f.Limit > 0 && len(hits) > f.Limit {
		hits = hits[:f.Limit]
	}
	return hits, nil
}

// Count returns the size of Search(...) without allocating the slice.
func (s *FilesystemStore) Count(conv UUID, f Filter) (int, error) {
	out, err := s.Search(conv, f)
	if err != nil {
		return 0, err
	}
	return len(out), nil
}

func (f Filter) matches(e Event) bool {
	if f.Kind != "" && e.Kind != f.Kind {
		return false
	}
	if !f.SinceInclude.IsZero() && e.Timestamp.Before(f.SinceInclude) {
		return false
	}
	if !f.UntilExclude.IsZero() && !e.Timestamp.Before(f.UntilExclude) {
		return false
	}
	return true
}
