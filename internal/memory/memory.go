package memory

import "time"

const SchemaVersion = 1

// Session is the on-disk session document (see memory-management rule).
type Session struct {
	SchemaVersion int               `json:"schema_version"`
	SessionID     string            `json:"session_id"`
	StartedAt     time.Time         `json:"started_at"`
	Commands      []CommandEntry    `json:"commands"`
	Preferences   map[string]string `json:"preferences"`
}

// CommandEntry records one pipeline invocation in the session.
type CommandEntry struct {
	RawInput  string            `json:"raw_input"`
	Intent    string            `json:"intent"`
	Params    map[string]string `json:"params"`
	Shell     string            `json:"shell"`
	RecordedAt time.Time        `json:"recorded_at"`
}

// Empty returns a new empty session.
func Empty(sessionID string) Session {
	return Session{
		SchemaVersion: SchemaVersion,
		SessionID:     sessionID,
		StartedAt:     time.Now().UTC(),
		Commands:      make([]CommandEntry, 0, 8),
		Preferences:   make(map[string]string),
	}
}
