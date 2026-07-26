// Package visitor demonstrates the Visitor behavioral pattern.
//
// A CMS stores an article as a tree of nodes: headings, paragraphs, lists, code
// blocks. That tree has to be rendered to HTML for the web, to plain text for
// the email digest, and walked again to count words and collect links for the
// search index.
//
// Putting a ToHTML, ToText and CountWords method on every node type means every
// new output format edits every node. Visitor turns it around: each new
// operation is one new type, and the nodes never change.
package visitor

import (
	"fmt"
	"strconv"
	"strings"
)

// Visitor declares one method per concrete node type.
type Visitor interface {
	VisitHeading(*Heading)
	VisitParagraph(*Paragraph)
	VisitList(*List)
	VisitCode(*Code)
}

// Node is the element interface. Accept is the double-dispatch hook: the node
// knows its own type, so it calls the matching visitor method.
type Node interface {
	Accept(Visitor)
}

// --- Concrete nodes ----------------------------------------------------------

type Heading struct {
	Level int
	Text  string
}

func (n *Heading) Accept(v Visitor) { v.VisitHeading(n) }

type Paragraph struct {
	Text  string
	Links []string
}

func (n *Paragraph) Accept(v Visitor) { v.VisitParagraph(n) }

type List struct {
	Ordered bool
	Items   []string
}

func (n *List) Accept(v Visitor) { v.VisitList(n) }

type Code struct {
	Language string
	Source   string
}

func (n *Code) Accept(v Visitor) { v.VisitCode(n) }

// Document is the object structure being traversed.
type Document struct {
	Nodes []Node
}

// Walk applies a visitor to every node in order.
func (d *Document) Walk(v Visitor) {
	for _, n := range d.Nodes {
		n.Accept(v)
	}
}

// --- Concrete visitors -------------------------------------------------------

// HTMLRenderer produces markup for the website.
type HTMLRenderer struct {
	buf strings.Builder
}

func (r *HTMLRenderer) VisitHeading(n *Heading) {
	level := n.Level
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}
	tag := "h" + strconv.Itoa(level)
	fmt.Fprintf(&r.buf, "<%s>%s</%s>", tag, n.Text, tag)
}

func (r *HTMLRenderer) VisitParagraph(n *Paragraph) {
	fmt.Fprintf(&r.buf, "<p>%s</p>", n.Text)
}

func (r *HTMLRenderer) VisitList(n *List) {
	tag := "ul"
	if n.Ordered {
		tag = "ol"
	}

	fmt.Fprintf(&r.buf, "<%s>", tag)
	for _, item := range n.Items {
		fmt.Fprintf(&r.buf, "<li>%s</li>", item)
	}
	fmt.Fprintf(&r.buf, "</%s>", tag)
}

func (r *HTMLRenderer) VisitCode(n *Code) {
	fmt.Fprintf(&r.buf, "<pre><code class=\"language-%s\">%s</code></pre>", n.Language, n.Source)
}

func (r *HTMLRenderer) String() string { return r.buf.String() }

// TextRenderer produces the plain-text version for the email digest.
type TextRenderer struct {
	lines []string
}

func (r *TextRenderer) VisitHeading(n *Heading) {
	r.lines = append(r.lines, strings.ToUpper(n.Text))
}

func (r *TextRenderer) VisitParagraph(n *Paragraph) {
	r.lines = append(r.lines, n.Text)
}

func (r *TextRenderer) VisitList(n *List) {
	for i, item := range n.Items {
		if n.Ordered {
			r.lines = append(r.lines, fmt.Sprintf("%d. %s", i+1, item))
		} else {
			r.lines = append(r.lines, "- "+item)
		}
	}
}

// Code is deliberately skipped in the digest — a visitor is free to ignore a
// node type.
func (r *TextRenderer) VisitCode(*Code) {}

func (r *TextRenderer) String() string { return strings.Join(r.lines, "\n") }

// Stats gathers indexing data in a single pass: word count plus every outbound
// link. Adding this required no change to any node type.
type Stats struct {
	Words     int
	Links     []string
	CodeLines int
}

func (s *Stats) VisitHeading(n *Heading) { s.Words += len(strings.Fields(n.Text)) }

func (s *Stats) VisitParagraph(n *Paragraph) {
	s.Words += len(strings.Fields(n.Text))
	s.Links = append(s.Links, n.Links...)
}

func (s *Stats) VisitList(n *List) {
	for _, item := range n.Items {
		s.Words += len(strings.Fields(item))
	}
}

// Source code is counted in lines, not words.
func (s *Stats) VisitCode(n *Code) {
	if strings.TrimSpace(n.Source) == "" {
		return
	}
	s.CodeLines += len(strings.Split(strings.TrimRight(n.Source, "\n"), "\n"))
}
