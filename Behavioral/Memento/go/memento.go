// Package memento demonstrates the Memento behavioral pattern.
//
// An editor needs undo. The naive approach is to let the history object reach
// into the document and copy its fields — but then every new field is a bug
// waiting to happen, and the document's internals are public forever.
//
// Memento inverts it: the document produces an opaque snapshot of itself and is
// the only thing that can read one back. The history stores snapshots without
// ever knowing what is inside them.
package memento

import (
	"errors"
	"time"
)

// Memento is the opaque snapshot. Its fields are unexported, so only the
// Document in this package can read them — the caretaker can only hold it and
// hand it back.
type Memento struct {
	title  string
	body   string
	tags   []string
	cursor int

	label   string
	takenAt time.Time
}

// Label and TakenAt are the only things a caretaker is allowed to see. Enough to
// build an "undo history" menu, not enough to reconstruct the document.
func (m Memento) Label() string { return m.label }

func (m Memento) TakenAt() time.Time { return m.takenAt }

// Document is the Originator: the object whose state gets snapshotted.
type Document struct {
	title  string
	body   string
	tags   []string
	cursor int
}

func NewDocument(title string) *Document {
	return &Document{title: title}
}

func (d *Document) Title() string  { return d.title }
func (d *Document) Body() string   { return d.body }
func (d *Document) Cursor() int    { return d.cursor }
func (d *Document) Tags() []string { return append([]string(nil), d.tags...) }

func (d *Document) SetTitle(t string) { d.title = t }

func (d *Document) Append(text string) {
	d.body += text
	d.cursor = len(d.body)
}

func (d *Document) AddTag(tag string) { d.tags = append(d.tags, tag) }

// Save captures the current state.
//
// The tags slice is copied, not referenced: without this, a later AddTag would
// mutate the snapshot too and undo would silently do nothing.
func (d *Document) Save(label string) Memento {
	return Memento{
		title:   d.title,
		body:    d.body,
		tags:    append([]string(nil), d.tags...),
		cursor:  d.cursor,
		label:   label,
		takenAt: time.Now(),
	}
}

// Restore rolls the document back to a snapshot.
func (d *Document) Restore(m Memento) {
	d.title = m.title
	d.body = m.body
	d.tags = append([]string(nil), m.tags...)
	d.cursor = m.cursor
}

// --- Caretaker ---------------------------------------------------------------

// ErrNothingToUndo and ErrNothingToRedo report an exhausted history.
var (
	ErrNothingToUndo = errors.New("memento: nothing to undo")
	ErrNothingToRedo = errors.New("memento: nothing to redo")
)

// History is the Caretaker. It stores mementos and decides when to put them
// back, but it cannot read them.
type History struct {
	doc  *Document
	undo []Memento
	redo []Memento

	// Limit caps how many snapshots are kept. Unbounded undo is a memory leak
	// in any long-lived editor.
	Limit int
}

func NewHistory(d *Document) *History {
	return &History{doc: d, Limit: 50}
}

// Do snapshots the document, then applies the change. Taking the snapshot first
// is what makes the change reversible.
func (h *History) Do(label string, mutate func(*Document)) {
	h.undo = append(h.undo, h.doc.Save(label))

	if h.Limit > 0 && len(h.undo) > h.Limit {
		h.undo = h.undo[len(h.undo)-h.Limit:]
	}

	// A fresh edit invalidates any redo branch.
	h.redo = nil

	mutate(h.doc)
}

// Undo reverts the most recent change.
func (h *History) Undo() error {
	if len(h.undo) == 0 {
		return ErrNothingToUndo
	}

	// Snapshot the present so Redo can come back to it.
	h.redo = append(h.redo, h.doc.Save("redo"))

	last := h.undo[len(h.undo)-1]
	h.undo = h.undo[:len(h.undo)-1]
	h.doc.Restore(last)

	return nil
}

// Redo re-applies the change that Undo reverted.
func (h *History) Redo() error {
	if len(h.redo) == 0 {
		return ErrNothingToRedo
	}

	h.undo = append(h.undo, h.doc.Save("undo"))

	last := h.redo[len(h.redo)-1]
	h.redo = h.redo[:len(h.redo)-1]
	h.doc.Restore(last)

	return nil
}

// Labels lists the pending undo steps, oldest first — the data an "Undo:
// rename title" menu entry is built from.
func (h *History) Labels() []string {
	out := make([]string, 0, len(h.undo))
	for _, m := range h.undo {
		out = append(out, m.Label())
	}

	return out
}

func (h *History) UndoDepth() int { return len(h.undo) }

func (h *History) RedoDepth() int { return len(h.redo) }
