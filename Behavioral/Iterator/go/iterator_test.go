package iterator

import (
	"errors"
	"testing"
)

// pagedAPI builds a Fetcher that serves the given pages in order and counts
// calls, so tests can assert on laziness.
func pagedAPI(pages [][]string) (Fetcher[string], *int) {
	calls := 0

	return func(cursor string) (Page[string], error) {
		calls++

		idx := 0
		if cursor != "" {
			// Cursors here are just "p1", "p2", ...
			idx = int(cursor[1] - '0')
		}
		if idx >= len(pages) {
			return Page[string]{}, nil
		}

		next := ""
		if idx+1 < len(pages) {
			next = "p" + string(rune('0'+idx+1))
		}

		return Page[string]{Items: pages[idx], NextCursor: next}, nil
	}, &calls
}

func TestIteratesEveryItemAcrossPages(t *testing.T) {
	fetch, _ := pagedAPI([][]string{
		{"alice", "bob"},
		{"carol"},
		{"dave", "erin"},
	})

	got, err := Collect(New(fetch))
	if err != nil {
		t.Fatalf("collect: %v", err)
	}

	want := []string{"alice", "bob", "carol", "dave", "erin"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("item %d = %q, want %q", i, got[i], want[i])
		}
	}
}

// The caller sees a flat sequence, but pages are only fetched as needed.
func TestFetchingIsLazy(t *testing.T) {
	fetch, calls := pagedAPI([][]string{
		{"alice", "bob"},
		{"carol"},
		{"dave"},
	})

	c := New(fetch)

	if *calls != 0 {
		t.Errorf("constructing the iterator already fetched %d pages", *calls)
	}

	c.Next()
	if *calls != 1 {
		t.Errorf("after one item: %d fetches, want 1", *calls)
	}

	// Second item is still on page one.
	c.Next()
	if *calls != 1 {
		t.Errorf("after two items: %d fetches, want still 1", *calls)
	}

	// Third item forces page two.
	c.Next()
	if *calls != 2 {
		t.Errorf("after three items: %d fetches, want 2", *calls)
	}
}

func TestEmptyFeed(t *testing.T) {
	fetch := func(cursor string) (Page[string], error) {
		return Page[string]{}, nil
	}

	got, err := Collect(New(fetch))

	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %v, want nothing", got)
	}
}

// An API that returns an empty page but a non-empty cursor must not end
// iteration prematurely.
func TestSkipsEmptyIntermediatePages(t *testing.T) {
	fetch, _ := pagedAPI([][]string{
		{"alice"},
		{},
		{"bob"},
	})

	got, err := Collect(New(fetch))
	if err != nil {
		t.Fatalf("collect: %v", err)
	}

	want := []string{"alice", "bob"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFetchErrorStopsIterationAndIsReported(t *testing.T) {
	calls := 0
	fetch := func(cursor string) (Page[string], error) {
		calls++
		if calls == 1 {
			return Page[string]{Items: []string{"alice"}, NextCursor: "p1"}, nil
		}

		return Page[string]{}, ErrUpstream
	}

	c := New(fetch)
	got, err := Collect(c)

	if !errors.Is(err, ErrUpstream) {
		t.Fatalf("err = %v, want ErrUpstream", err)
	}
	// The items retrieved before the failure are still returned.
	if len(got) != 1 || got[0] != "alice" {
		t.Errorf("got %v, want the pre-failure items", got)
	}
	// Further calls must not retry forever.
	if c.Next() {
		t.Error("Next kept returning true after an error")
	}
}

func TestRangeOverFunc(t *testing.T) {
	fetch, _ := pagedAPI([][]string{{"alice", "bob"}, {"carol"}})

	var got []string
	c := New(fetch)
	for v := range c.All() {
		got = append(got, v)
	}

	if c.Err() != nil {
		t.Fatalf("err = %v", c.Err())
	}
	if len(got) != 3 {
		t.Errorf("got %v, want 3 items", got)
	}
}

// Breaking out of the range must not keep fetching.
func TestRangeOverFuncCanBreakEarly(t *testing.T) {
	fetch, calls := pagedAPI([][]string{{"alice"}, {"bob"}, {"carol"}})

	c := New(fetch)
	for range c.All() {
		break
	}

	if *calls != 1 {
		t.Errorf("%d pages fetched after an early break, want 1", *calls)
	}
}

func TestPagesCountsUpstreamRequests(t *testing.T) {
	fetch, _ := pagedAPI([][]string{{"alice"}, {"bob"}, {"carol"}})

	c := New(fetch)
	if _, err := Collect(c); err != nil {
		t.Fatalf("collect: %v", err)
	}

	// Three data pages, plus the final request that reports no next cursor is
	// avoided because the last page carries an empty cursor.
	if c.Pages() != 3 {
		t.Errorf("pages = %d, want 3", c.Pages())
	}
}

// Iterator is generic: the same machinery works for any element type.
func TestWorksWithAnyElementType(t *testing.T) {
	type user struct {
		ID   int
		Name string
	}

	fetch := func(cursor string) (Page[user], error) {
		if cursor == "" {
			return Page[user]{Items: []user{{1, "alice"}}, NextCursor: "p1"}, nil
		}

		return Page[user]{Items: []user{{2, "bob"}}}, nil
	}

	got, err := Collect(New(fetch))
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	if len(got) != 2 || got[1].Name != "bob" {
		t.Errorf("got %+v", got)
	}
}
