// Package eventstore implements the event-sourced conversation core that
// sits at the bottom of OpenHands' app_server.
//
// Upstream reference: openhands/app_server/event/event_service.py and
// openhands/app_server/event/filesystem_event_service.py at SHA
// a89778f3d7036b8d81d57a1f93e31c6df8219eff.
//
// The shape we're teaching: every conversation between a user and an agent
// is reduced to an *append-only stream of typed events* on disk. Anything
// stateful in the rest of the server (live status, callbacks, replay)
// reads back from this stream. s01 keeps the type set small (Message /
// Action / Observation) and the storage trivial (one JSON file per event
// in a per-conversation directory).
package eventstore

import (
	"encoding/json"
	"fmt"
	"time"
)

// Kind is the discriminator stamped into every persisted event. Upstream
// calls this `EventKind`; the values match upstream's three primary
// categories (other kinds — error, system — land in s04 callbacks).
type Kind string

const (
	KindMessage     Kind = "message"
	KindAction      Kind = "action"
	KindObservation Kind = "observation"
)

// Event is the on-wire shape. Upstream uses pydantic's
// DiscriminatedUnionMixin to dispatch on `kind`; in Go we use a tagged
// struct and rebuild the variant from `Payload` on read.
type Event struct {
	ID             UUID            `json:"id"`
	ConversationID UUID            `json:"conversation_id"`
	Kind           Kind            `json:"kind"`
	Timestamp      time.Time       `json:"timestamp"`
	Payload        json.RawMessage `json:"payload"`
}

// Message is a chat-style turn (user or assistant text).
type Message struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

// Action is something the agent decided to do (a shell command, a tool call).
// In s01 we keep it as a free-form "command" body; s06 will type it.
type Action struct {
	Name string `json:"name"`
	Body string `json:"body"`
}

// Observation is the world's response to an Action.
type Observation struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

// NewMessage / NewAction / NewObservation are constructors that stamp
// id+timestamp and serialise the payload. They are the only ways the
// outside world should be making Events in s01.
func NewMessage(conv UUID, m Message) (Event, error) {
	return newEvent(conv, KindMessage, m)
}

func NewAction(conv UUID, a Action) (Event, error) {
	return newEvent(conv, KindAction, a)
}

func NewObservation(conv UUID, o Observation) (Event, error) {
	return newEvent(conv, KindObservation, o)
}

func newEvent(conv UUID, kind Kind, payload any) (Event, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return Event{}, fmt.Errorf("marshal payload: %w", err)
	}
	return Event{
		ID:             NewUUID(),
		ConversationID: conv,
		Kind:           kind,
		Timestamp:      time.Now().UTC(),
		Payload:        body,
	}, nil
}

// DecodeMessage / DecodeAction / DecodeObservation pull the typed payload
// back out. Callers usually switch on e.Kind first.
func (e Event) DecodeMessage() (Message, error) {
	var m Message
	err := json.Unmarshal(e.Payload, &m)
	return m, err
}

func (e Event) DecodeAction() (Action, error) {
	var a Action
	err := json.Unmarshal(e.Payload, &a)
	return a, err
}

func (e Event) DecodeObservation() (Observation, error) {
	var o Observation
	err := json.Unmarshal(e.Payload, &o)
	return o, err
}
