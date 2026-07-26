package templatemethod

import (
	"errors"
	"testing"
)

func TestCSVImportUsesItsOverriddenHooks(t *testing.T) {
	w := &Warehouse{}
	raw := []byte("widget,1200\ngadget,900\nwidget,1200\nfreebie,0\n")

	report, err := Run(CSVImporter{}, raw, w)
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if report.Format != "csv" {
		t.Errorf("format = %q", report.Format)
	}
	if report.Parsed != 4 {
		t.Errorf("parsed = %d, want 4", report.Parsed)
	}
	// Dedupe hook is implemented, so the repeated widget is dropped.
	if report.Duplicates != 1 {
		t.Errorf("duplicates = %d, want 1", report.Duplicates)
	}
	// Validate hook is overridden, so the zero-price row is rejected.
	if report.Rejected != 1 {
		t.Errorf("rejected = %d, want 1", report.Rejected)
	}
	if report.Loaded != 2 || len(w.Rows) != 2 {
		t.Errorf("loaded = %d, warehouse = %d, want 2 and 2", report.Loaded, len(w.Rows))
	}
}

// JSONImporter implements neither optional hook, so it must fall back to the
// skeleton's defaults rather than failing or skipping the steps.
func TestJSONImportFallsBackToDefaults(t *testing.T) {
	w := &Warehouse{}
	raw := []byte(`[{"sku":"widget","price":1200},{"sku":"widget","price":1200},{"sku":"freebie","price":0}]`)

	report, err := Run(JSONImporter{}, raw, w)
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if report.Duplicates != 0 {
		t.Errorf("duplicates = %d, want 0 — JSON has no dedupe hook", report.Duplicates)
	}
	// Default validation permits a zero price.
	if report.Rejected != 0 {
		t.Errorf("rejected = %d, want 0", report.Rejected)
	}
	if report.Loaded != 3 {
		t.Errorf("loaded = %d, want 3", report.Loaded)
	}
}

// Both formats run through the identical skeleton.
func TestSkeletonIsSharedAcrossFormats(t *testing.T) {
	tests := []struct {
		name     string
		importer Importer
		raw      []byte
	}{
		{"csv", CSVImporter{}, []byte("widget,1200")},
		{"json", JSONImporter{}, []byte(`[{"sku":"widget","price":1200}]`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Warehouse{}

			report, err := Run(tt.importer, tt.raw, w)
			if err != nil {
				t.Fatalf("run: %v", err)
			}

			if report.Loaded != 1 {
				t.Errorf("loaded = %d, want 1", report.Loaded)
			}
			if len(w.Rows) != 1 || w.Rows[0].SKU != "widget" || w.Rows[0].Price != 1200 {
				t.Errorf("warehouse = %+v", w.Rows)
			}
		})
	}
}

func TestEmptyPayloadIsRejectedBeforeParsing(t *testing.T) {
	w := &Warehouse{}

	_, err := Run(CSVImporter{}, nil, w)

	if !errors.Is(err, ErrEmptyPayload) {
		t.Errorf("got %v, want ErrEmptyPayload", err)
	}
	if len(w.Rows) != 0 {
		t.Error("warehouse was written to despite the empty payload")
	}
}

func TestParseFailureAbortsTheRun(t *testing.T) {
	tests := []struct {
		name     string
		importer Importer
		raw      []byte
	}{
		{"csv wrong column count", CSVImporter{}, []byte("widget,1200,extra")},
		{"csv non-numeric price", CSVImporter{}, []byte("widget,abc")},
		{"json malformed", JSONImporter{}, []byte("{not json")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Warehouse{}

			_, err := Run(tt.importer, tt.raw, w)

			if err == nil {
				t.Fatal("expected a parse error")
			}
			if len(w.Rows) != 0 {
				t.Error("warehouse was written to despite the parse failure")
			}
		})
	}
}

// One bad row must not lose the whole feed.
func TestInvalidRowsAreSkippedNotFatal(t *testing.T) {
	w := &Warehouse{}
	raw := []byte(`[{"sku":"","price":100},{"sku":"good","price":100}]`)

	report, err := Run(JSONImporter{}, raw, w)
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if report.Rejected != 1 || report.Loaded != 1 {
		t.Errorf("rejected = %d, loaded = %d, want 1 and 1", report.Rejected, report.Loaded)
	}
	if len(report.Errors) != 1 {
		t.Errorf("errors = %v, want one entry", report.Errors)
	}
	if len(w.Rows) != 1 || w.Rows[0].SKU != "good" {
		t.Errorf("warehouse = %+v", w.Rows)
	}
}

func TestDedupePreservesFirstOccurrence(t *testing.T) {
	in := []Record{
		{SKU: "a", Price: 1},
		{SKU: "b", Price: 2},
		{SKU: "a", Price: 999},
	}

	out := CSVImporter{}.Dedupe(in)

	if len(out) != 2 {
		t.Fatalf("got %d records, want 2", len(out))
	}
	if out[0].Price != 1 {
		t.Errorf("kept the later duplicate: %+v", out[0])
	}
	// Dedupe must not corrupt the caller's slice.
	if in[1].SKU != "b" {
		t.Errorf("input slice was clobbered: %+v", in)
	}
}
