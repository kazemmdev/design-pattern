package visitor

import (
	"strings"
	"testing"
)

func sample() *Document {
	return &Document{Nodes: []Node{
		&Heading{Level: 1, Text: "Release notes"},
		&Paragraph{
			Text:  "We shipped the new checkout today.",
			Links: []string{"https://example.com/changelog"},
		},
		&List{Ordered: false, Items: []string{"faster", "safer"}},
		&Code{Language: "go", Source: "func main() {}\nfmt.Println()"},
	}}
}

func TestHTMLRendering(t *testing.T) {
	r := &HTMLRenderer{}
	sample().Walk(r)

	got := r.String()
	for _, want := range []string{
		"<h1>Release notes</h1>",
		"<p>We shipped the new checkout today.</p>",
		"<ul><li>faster</li><li>safer</li></ul>",
		`<pre><code class="language-go">`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q\ngot: %s", want, got)
		}
	}
}

func TestTextRendering(t *testing.T) {
	r := &TextRenderer{}
	sample().Walk(r)

	got := r.String()

	if !strings.Contains(got, "RELEASE NOTES") {
		t.Errorf("heading not upper-cased: %s", got)
	}
	if !strings.Contains(got, "- faster") {
		t.Errorf("list not bulleted: %s", got)
	}
	// This visitor deliberately ignores code blocks.
	if strings.Contains(got, "func main") {
		t.Errorf("code block should be skipped in the digest: %s", got)
	}
}

// The same tree, walked by a third visitor, with no node type modified.
func TestStatsGathersIndexingDataInOnePass(t *testing.T) {
	s := &Stats{}
	sample().Walk(s)

	// "Release notes" = 2, paragraph = 6, list items = 2.
	if s.Words != 10 {
		t.Errorf("words = %d, want 10", s.Words)
	}
	if len(s.Links) != 1 || s.Links[0] != "https://example.com/changelog" {
		t.Errorf("links = %v", s.Links)
	}
	if s.CodeLines != 2 {
		t.Errorf("code lines = %d, want 2", s.CodeLines)
	}
}

func TestOrderedListRendersDifferently(t *testing.T) {
	doc := &Document{Nodes: []Node{
		&List{Ordered: true, Items: []string{"first", "second"}},
	}}

	html := &HTMLRenderer{}
	doc.Walk(html)
	if !strings.Contains(html.String(), "<ol>") {
		t.Errorf("html = %s, want an <ol>", html.String())
	}

	text := &TextRenderer{}
	doc.Walk(text)
	if !strings.Contains(text.String(), "1. first") {
		t.Errorf("text = %s, want numbered items", text.String())
	}
}

func TestHeadingLevelIsClamped(t *testing.T) {
	tests := []struct {
		name  string
		level int
		want  string
	}{
		{"below range", 0, "<h1>"},
		{"in range", 3, "<h3>"},
		{"above range", 99, "<h6>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HTMLRenderer{}
			(&Document{Nodes: []Node{&Heading{Level: tt.level, Text: "x"}}}).Walk(r)

			if !strings.Contains(r.String(), tt.want) {
				t.Errorf("got %s, want %s", r.String(), tt.want)
			}
		})
	}
}

func TestEmptyDocumentProducesEmptyOutput(t *testing.T) {
	doc := &Document{}

	html := &HTMLRenderer{}
	doc.Walk(html)
	if html.String() != "" {
		t.Errorf("html = %q, want empty", html.String())
	}

	s := &Stats{}
	doc.Walk(s)
	if s.Words != 0 || len(s.Links) != 0 {
		t.Errorf("stats = %+v, want zero", s)
	}
}

func TestEmptyCodeBlockCountsNoLines(t *testing.T) {
	s := &Stats{}
	(&Document{Nodes: []Node{&Code{Language: "go", Source: "   "}}}).Walk(s)

	if s.CodeLines != 0 {
		t.Errorf("code lines = %d, want 0", s.CodeLines)
	}
}

// Walk must visit nodes in document order.
func TestTraversalOrderIsPreserved(t *testing.T) {
	doc := &Document{Nodes: []Node{
		&Heading{Level: 2, Text: "second"},
		&Heading{Level: 1, Text: "first"},
	}}

	r := &TextRenderer{}
	doc.Walk(r)

	lines := strings.Split(r.String(), "\n")
	if len(lines) != 2 || lines[0] != "SECOND" || lines[1] != "FIRST" {
		t.Errorf("lines = %v, want document order", lines)
	}
}

// Every node type must be reachable through the Visitor interface — if a node
// were added without a matching method, this would not compile.
func TestVisitorCoversEveryNodeType(t *testing.T) {
	nodes := []Node{
		&Heading{Level: 1, Text: "h"},
		&Paragraph{Text: "p"},
		&List{Items: []string{"i"}},
		&Code{Language: "go", Source: "x"},
	}

	for _, n := range nodes {
		s := &Stats{}
		n.Accept(s)
	}
}
