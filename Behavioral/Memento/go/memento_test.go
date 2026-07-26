package memento

import (
	"errors"
	"testing"
)

func TestUndoRevertsTheLastChange(t *testing.T) {
	doc := NewDocument("Draft")
	h := NewHistory(doc)

	h.Do("write intro", func(d *Document) { d.Append("Hello") })
	h.Do("write more", func(d *Document) { d.Append(" world") })

	if doc.Body() != "Hello world" {
		t.Fatalf("body = %q", doc.Body())
	}

	if err := h.Undo(); err != nil {
		t.Fatalf("undo: %v", err)
	}
	if doc.Body() != "Hello" {
		t.Errorf("body = %q, want %q", doc.Body(), "Hello")
	}

	if err := h.Undo(); err != nil {
		t.Fatalf("undo: %v", err)
	}
	if doc.Body() != "" {
		t.Errorf("body = %q, want empty", doc.Body())
	}
}

func TestRedoReappliesAnUndoneChange(t *testing.T) {
	doc := NewDocument("Draft")
	h := NewHistory(doc)

	h.Do("write", func(d *Document) { d.Append("Hello") })
	_ = h.Undo()

	if err := h.Redo(); err != nil {
		t.Fatalf("redo: %v", err)
	}
	if doc.Body() != "Hello" {
		t.Errorf("body = %q, want %q", doc.Body(), "Hello")
	}
}

func TestUndoRestoresEveryField(t *testing.T) {
	doc := NewDocument("Draft")
	h := NewHistory(doc)

	h.Do("edit everything", func(d *Document) {
		d.SetTitle("Published")
		d.Append("content")
		d.AddTag("release")
	})

	if err := h.Undo(); err != nil {
		t.Fatalf("undo: %v", err)
	}

	if doc.Title() != "Draft" {
		t.Errorf("title = %q, want Draft", doc.Title())
	}
	if doc.Body() != "" {
		t.Errorf("body = %q, want empty", doc.Body())
	}
	if len(doc.Tags()) != 0 {
		t.Errorf("tags = %v, want none", doc.Tags())
	}
	if doc.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", doc.Cursor())
	}
}

// The snapshot must be a deep copy: a slice shared with the live document would
// make undo a silent no-op.
func TestSnapshotDoesNotAliasTheDocument(t *testing.T) {
	doc := NewDocument("Draft")
	doc.AddTag("first")

	snapshot := doc.Save("before")
	doc.AddTag("second")

	doc.Restore(snapshot)

	tags := doc.Tags()
	if len(tags) != 1 || tags[0] != "first" {
		t.Errorf("tags = %v, want only [first]", tags)
	}
}

// Restoring must also copy, or a later edit would corrupt the stored memento.
func TestRestoreDoesNotAliasTheSnapshot(t *testing.T) {
	doc := NewDocument("Draft")
	doc.AddTag("first")
	snapshot := doc.Save("before")

	doc.Restore(snapshot)
	doc.AddTag("mutation")
	doc.Restore(snapshot)

	if got := doc.Tags(); len(got) != 1 {
		t.Errorf("tags = %v, want the snapshot to have survived", got)
	}
}

func TestTagsAccessorIsACopy(t *testing.T) {
	doc := NewDocument("Draft")
	doc.AddTag("first")

	tags := doc.Tags()
	tags[0] = "hacked"

	if doc.Tags()[0] != "first" {
		t.Error("Tags() exposed the document's internal slice")
	}
}

// A new edit after an undo discards the redo branch — standard editor behaviour.
func TestNewEditClearsTheRedoBranch(t *testing.T) {
	doc := NewDocument("Draft")
	h := NewHistory(doc)

	h.Do("a", func(d *Document) { d.Append("a") })
	_ = h.Undo()

	if h.RedoDepth() != 1 {
		t.Fatalf("redo depth = %d, want 1", h.RedoDepth())
	}

	h.Do("b", func(d *Document) { d.Append("b") })

	if h.RedoDepth() != 0 {
		t.Errorf("redo depth = %d, want 0 after a new edit", h.RedoDepth())
	}
	if err := h.Redo(); !errors.Is(err, ErrNothingToRedo) {
		t.Errorf("got %v, want ErrNothingToRedo", err)
	}
}

func TestExhaustedHistoryReportsAnError(t *testing.T) {
	doc := NewDocument("Draft")
	h := NewHistory(doc)

	if err := h.Undo(); !errors.Is(err, ErrNothingToUndo) {
		t.Errorf("got %v, want ErrNothingToUndo", err)
	}
	if err := h.Redo(); !errors.Is(err, ErrNothingToRedo) {
		t.Errorf("got %v, want ErrNothingToRedo", err)
	}
}

// Unbounded undo is a memory leak, so old snapshots are dropped.
func TestHistoryRespectsItsLimit(t *testing.T) {
	doc := NewDocument("Draft")
	h := NewHistory(doc)
	h.Limit = 3

	for _, label := range []string{"a", "b", "c", "d", "e"} {
		l := label
		h.Do(l, func(d *Document) { d.Append(l) })
	}

	if h.UndoDepth() != 3 {
		t.Errorf("undo depth = %d, want 3", h.UndoDepth())
	}

	want := []string{"c", "d", "e"}
	got := h.Labels()
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("labels = %v, want the newest %v", got, want)
		}
	}
}

// The caretaker can read the metadata but nothing about the content.
func TestCaretakerSeesLabelsOnly(t *testing.T) {
	doc := NewDocument("Secret")
	h := NewHistory(doc)

	h.Do("rename title", func(d *Document) { d.SetTitle("Public") })

	labels := h.Labels()
	if len(labels) != 1 || labels[0] != "rename title" {
		t.Errorf("labels = %v", labels)
	}

	m := doc.Save("snapshot")
	if m.Label() != "snapshot" {
		t.Errorf("label = %q", m.Label())
	}
	if m.TakenAt().IsZero() {
		t.Error("takenAt was not set")
	}
}

func TestRepeatedUndoRedoIsStable(t *testing.T) {
	doc := NewDocument("Draft")
	h := NewHistory(doc)

	h.Do("write", func(d *Document) { d.Append("Hello") })

	for i := 0; i < 5; i++ {
		if err := h.Undo(); err != nil {
			t.Fatalf("undo %d: %v", i, err)
		}
		if doc.Body() != "" {
			t.Fatalf("iteration %d: body = %q, want empty", i, doc.Body())
		}
		if err := h.Redo(); err != nil {
			t.Fatalf("redo %d: %v", i, err)
		}
		if doc.Body() != "Hello" {
			t.Fatalf("iteration %d: body = %q, want Hello", i, doc.Body())
		}
	}
}
