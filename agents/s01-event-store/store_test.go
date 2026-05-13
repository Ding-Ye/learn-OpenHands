package eventstore

import (
	"testing"
	"time"
)

func newTestStore(t *testing.T) (*FilesystemStore, UUID) {
	t.Helper()
	s, err := NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewFilesystemStore: %v", err)
	}
	return s, NewUUID()
}

func TestRoundTripMessage(t *testing.T) {
	s, conv := newTestStore(t)

	e, err := NewMessage(conv, Message{Role: "user", Text: "hello"})
	if err != nil {
		t.Fatalf("NewMessage: %v", err)
	}
	if err := s.Save(conv, e); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := s.Get(conv, e.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Kind != KindMessage {
		t.Errorf("Kind: got %q want message", got.Kind)
	}
	m, err := got.DecodeMessage()
	if err != nil {
		t.Fatalf("DecodeMessage: %v", err)
	}
	if m.Role != "user" || m.Text != "hello" {
		t.Errorf("payload: got %+v", m)
	}
}

func TestSearchFiltersByKind(t *testing.T) {
	s, conv := newTestStore(t)

	for i := 0; i < 3; i++ {
		e, _ := NewMessage(conv, Message{Role: "user", Text: "m"})
		_ = s.Save(conv, e)
	}
	e, _ := NewAction(conv, Action{Name: "bash", Body: "ls"})
	_ = s.Save(conv, e)

	hits, err := s.Search(conv, Filter{Kind: KindAction})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("got %d action events, want 1", len(hits))
	}
	if hits[0].Kind != KindAction {
		t.Errorf("Kind: %q", hits[0].Kind)
	}
}

func TestSearchSortsAscending(t *testing.T) {
	s, conv := newTestStore(t)

	for i := 0; i < 5; i++ {
		e, _ := NewMessage(conv, Message{Role: "user", Text: "m"})
		_ = s.Save(conv, e)
		time.Sleep(2 * time.Millisecond) // separate timestamps
	}

	hits, err := s.Search(conv, Filter{SortAsc: true})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) != 5 {
		t.Fatalf("got %d, want 5", len(hits))
	}
	for i := 1; i < len(hits); i++ {
		if hits[i].Timestamp.Before(hits[i-1].Timestamp) {
			t.Fatalf("not ascending: %v then %v", hits[i-1].Timestamp, hits[i].Timestamp)
		}
	}
}

func TestSearchLimit(t *testing.T) {
	s, conv := newTestStore(t)

	for i := 0; i < 10; i++ {
		e, _ := NewMessage(conv, Message{Role: "user", Text: "m"})
		_ = s.Save(conv, e)
	}

	hits, err := s.Search(conv, Filter{Limit: 3})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(hits) != 3 {
		t.Errorf("got %d, want 3", len(hits))
	}
}

func TestCount(t *testing.T) {
	s, conv := newTestStore(t)

	for i := 0; i < 7; i++ {
		e, _ := NewMessage(conv, Message{Role: "user", Text: "m"})
		_ = s.Save(conv, e)
	}
	n, err := s.Count(conv, Filter{})
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if n != 7 {
		t.Errorf("Count: got %d, want 7", n)
	}
}

func TestSaveRejectsMismatchedConv(t *testing.T) {
	s, conv := newTestStore(t)

	e, _ := NewMessage(conv, Message{Role: "user", Text: "hi"})
	other := NewUUID()
	if err := s.Save(other, e); err == nil {
		t.Fatal("Save(other-conv, e): expected error, got nil")
	}
}

func TestUUIDRoundTrip(t *testing.T) {
	u := NewUUID()
	s := u.String()
	if len(s) != 36 {
		t.Fatalf("String: len=%d, want 36", len(s))
	}
	v, err := ParseUUID(s)
	if err != nil {
		t.Fatalf("ParseUUID: %v", err)
	}
	if v != u {
		t.Errorf("round-trip mismatch")
	}
}
