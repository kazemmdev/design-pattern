// Package iterator demonstrates the Iterator behavioral pattern.
//
// A REST API hands back results one page at a time behind an opaque cursor. The
// calling code should not have to know that: it wants "give me every user". The
// Iterator holds the paging state — current page, next cursor, whether the feed
// is exhausted — and exposes a flat sequence, fetching lazily as it goes.
package iterator

import (
	"errors"
	"iter"
)

// Page is one response from the upstream API.
type Page[T any] struct {
	Items []T
	// NextCursor is empty on the last page.
	NextCursor string
}

// Fetcher retrieves a single page. An empty cursor means "the first page".
type Fetcher[T any] func(cursor string) (Page[T], error)

// Cursor is the Iterator. It walks every item across every page, requesting the
// next page only when the current one runs out.
type Cursor[T any] struct {
	fetch Fetcher[T]

	buf     []T
	pos     int
	next    string
	started bool
	done    bool

	current T
	err     error
	pages   int
}

// New returns an iterator over everything fetch can produce.
func New[T any](fetch Fetcher[T]) *Cursor[T] {
	return &Cursor[T]{fetch: fetch}
}

// Next advances to the next item, returning false when the feed is exhausted or
// a fetch failed. Check Err after the loop to tell those two apart.
func (c *Cursor[T]) Next() bool {
	if c.err != nil || c.done {
		return false
	}

	// Refill from the next page, skipping any empty pages the API returns.
	for c.pos >= len(c.buf) {
		if c.started && c.next == "" {
			c.done = true

			return false
		}

		page, err := c.fetch(c.next)
		if err != nil {
			c.err = err

			return false
		}

		c.started = true
		c.pages++
		c.buf = page.Items
		c.pos = 0
		c.next = page.NextCursor

		// A page with no items and no cursor means we are finished.
		if len(page.Items) == 0 && page.NextCursor == "" {
			c.done = true

			return false
		}
	}

	c.current = c.buf[c.pos]
	c.pos++

	return true
}

// Value returns the item found by the most recent Next.
func (c *Cursor[T]) Value() T { return c.current }

// Err reports why iteration stopped, if it was not simple exhaustion.
func (c *Cursor[T]) Err() error { return c.err }

// Pages reports how many upstream requests were made — the evidence that
// fetching really is lazy.
func (c *Cursor[T]) Pages() int { return c.pages }

// All adapts the iterator to Go 1.23 range-over-func, so callers can write
// `for u := range cursor.All()`. Check Err afterwards.
func (c *Cursor[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for c.Next() {
			if !yield(c.Value()) {
				return
			}
		}
	}
}

// Collect drains the iterator into a slice. Convenient, but it gives up the
// memory advantage of iterating lazily.
func Collect[T any](c *Cursor[T]) ([]T, error) {
	var out []T
	for c.Next() {
		out = append(out, c.Value())
	}

	return out, c.Err()
}

// ErrUpstream is a sample failure from the paged API.
var ErrUpstream = errors.New("iterator: upstream request failed")
