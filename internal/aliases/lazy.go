package aliases

import (
	"context"
	"sync"
)

// LazyLookup defers loading ~/.clx/aliases.yaml until the first Lookup call.
type LazyLookup struct {
	maxAliases int

	once  sync.Once
	store *Store
	err   error
}

// NewLazyLookup returns a parser.AliasLookup that opens the alias store on first use.
func NewLazyLookup(maxAliases int) *LazyLookup {
	return &LazyLookup{maxAliases: maxAliases}
}

// Lookup implements parser.AliasLookup.
func (l *LazyLookup) Lookup(name string) (string, bool) {
	l.once.Do(func() {
		l.store, l.err = Open(context.Background(), l.maxAliases)
	})
	if l.store == nil {
		return "", false
	}
	return l.store.Lookup(name)
}

// Store returns the loaded store after first Lookup, or nil.
func (l *LazyLookup) Store() *Store {
	return l.store
}

// Err returns the open error from the first Lookup, if any.
func (l *LazyLookup) Err() error {
	return l.err
}
