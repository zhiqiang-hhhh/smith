// Package trace provides a lightweight global trace recorder for debugging
// Crush behavior at runtime. Users toggle tracing via the /trace command;
// while active, key agent lifecycle events are collected in memory. When
// tracing stops, the collected events are returned for analysis.
package trace

import (
	"encoding/json"
	"strings"
	"sync"
	"time"
)

// Event is a single trace record.
type Event struct {
	Time      string         `json:"time"`
	Category  string         `json:"category"`
	Event     string         `json:"event"`
	SessionID string         `json:"session_id,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
}

type Snapshot struct {
	SessionID  string
	StartedAt  int64
	StoppedAt  int64
	EventCount int
	DataJSONL  string
}

var (
	mu        sync.Mutex
	active    bool
	sessionID string
	startedAt int64
	events    []Event
)

// Start begins recording trace events in memory for the given session.
func Start(id string) {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now().UTC()
	active = true
	sessionID = id
	startedAt = now.Unix()
	events = []Event{{
		Time:      now.Format(time.RFC3339Nano),
		Category:  "trace",
		Event:     "started",
		SessionID: id,
	}}
}

// StopSnapshot ends recording and returns all collected events.
// The zero Snapshot is returned if tracing was not active.
func StopSnapshot() Snapshot {
	mu.Lock()
	defer mu.Unlock()

	if !active {
		return Snapshot{}
	}

	now := time.Now().UTC()
	events = append(events, Event{
		Time:      now.Format(time.RFC3339Nano),
		Category:  "trace",
		Event:     "stopped",
		SessionID: sessionID,
	})

	var sb strings.Builder
	enc := json.NewEncoder(&sb)
	enc.SetEscapeHTML(false)
	for _, e := range events {
		_ = enc.Encode(e)
	}

	result := Snapshot{
		SessionID:  sessionID,
		StartedAt:  startedAt,
		StoppedAt:  now.Unix(),
		EventCount: len(events),
		DataJSONL:  sb.String(),
	}
	active = false
	sessionID = ""
	startedAt = 0
	events = nil
	return result
}

// Stop ends recording and returns all collected events as a JSONL string.
// Returns empty string if tracing was not active.
func Stop() string {
	return StopSnapshot().DataJSONL
}

// IsActive reports whether tracing is currently active.
func IsActive() bool {
	mu.Lock()
	defer mu.Unlock()
	return active
}

// Emit records a trace event. No-op when tracing is inactive.
func Emit(category, event, sessionID string, data map[string]any) {
	mu.Lock()
	defer mu.Unlock()

	if !active {
		return
	}

	events = append(events, Event{
		Time:      time.Now().UTC().Format(time.RFC3339Nano),
		Category:  category,
		Event:     event,
		SessionID: sessionID,
		Data:      data,
	})
}
