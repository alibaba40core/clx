package aliases

import "errors"

const (
	SchemaVersion = 1
	maxNameLen    = 64
	maxValueLen   = 4096
)

// Entry is one user-defined alias.
type Entry struct {
	Name  string
	Value string
}

// File is the on-disk aliases document.
type File struct {
	SchemaVersion int
	Entries       []Entry
}

var (
	ErrNotFound      = errors.New("alias not found")
	ErrExists        = errors.New("alias already exists")
	ErrInvalidName   = errors.New("invalid alias name")
	ErrInvalidValue  = errors.New("invalid alias value")
	ErrLimitExceeded = errors.New("alias limit exceeded")
)
