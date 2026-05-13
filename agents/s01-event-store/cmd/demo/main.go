// Demo: drive a fake conversation through the s01 event store, then read
// it back to recover the trajectory.
//
// Run from this directory:
//
//	go run ./cmd/demo
//
// Output is a few lines showing the persisted directory, the events that
// were saved, and the round-tripped read.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	es "github.com/Ding-Ye/learn-OpenHands/s01-event-store"
)

func main() {
	tmp, err := os.MkdirTemp("", "learn-openhands-s01-")
	if err != nil {
		die(err)
	}
	fmt.Println("store root:", tmp)

	store, err := es.NewFilesystemStore(tmp)
	if err != nil {
		die(err)
	}
	conv := es.NewUUID()
	fmt.Println("conversation:", conv)

	turns := []func() (es.Event, error){
		func() (es.Event, error) {
			return es.NewMessage(conv, es.Message{Role: "user", Text: "fix the off-by-one in count.go"})
		},
		func() (es.Event, error) {
			return es.NewAction(conv, es.Action{Name: "bash", Body: "cat count.go"})
		},
		func() (es.Event, error) {
			return es.NewObservation(conv, es.Observation{Stdout: "func Count(xs []int) int { return len(xs) - 1 }", ExitCode: 0})
		},
		func() (es.Event, error) {
			return es.NewAction(conv, es.Action{Name: "edit", Body: "s/len(xs) - 1/len(xs)/"})
		},
		func() (es.Event, error) {
			return es.NewObservation(conv, es.Observation{Stdout: "patched", ExitCode: 0})
		},
		func() (es.Event, error) {
			return es.NewMessage(conv, es.Message{Role: "assistant", Text: "fixed; tests pass."})
		},
	}

	for _, mk := range turns {
		e, err := mk()
		if err != nil {
			die(err)
		}
		if err := store.Save(conv, e); err != nil {
			die(err)
		}
		fmt.Printf("  saved %s %-12s\n", e.Timestamp.Format("15:04:05.000"), e.Kind)
	}

	files, _ := filepath.Glob(filepath.Join(tmp, conv.String(), "*.json"))
	fmt.Printf("on disk: %d files in %s/\n", len(files), filepath.Join(tmp, conv.String()))

	hits, err := store.Search(conv, es.Filter{SortAsc: true})
	if err != nil {
		die(err)
	}
	fmt.Println("replay:")
	for _, e := range hits {
		body, _ := json.Marshal(json.RawMessage(e.Payload))
		fmt.Printf("  %-12s %s\n", e.Kind, string(body))
	}

	onlyActions, _ := store.Search(conv, es.Filter{Kind: es.KindAction, SortAsc: true})
	fmt.Printf("actions only: %d\n", len(onlyActions))
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "demo:", err)
	os.Exit(1)
}
