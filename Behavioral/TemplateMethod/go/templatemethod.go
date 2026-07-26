// Package templatemethod demonstrates the Template Method behavioral pattern.
//
// Every nightly import does the same five things in the same order: decode the
// payload, validate each record, drop duplicates, write to the warehouse, and
// report what happened. Only the decoding and the validation rules differ
// between a CSV feed and a JSON feed. Template Method fixes the skeleton in one
// place and lets each format supply the steps that genuinely vary.
//
// Go has no inheritance, so the "abstract method" role is played by an interface
// and the optional hooks by interface assertions — the idiomatic translation.
package templatemethod

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Record is one imported row.
type Record struct {
	SKU   string
	Price int // minor units
}

// Report summarises a run.
type Report struct {
	Format     string
	Parsed     int
	Duplicates int
	Rejected   int
	Loaded     int
	Errors     []string
}

// Warehouse is the destination the skeleton always writes to.
type Warehouse struct {
	Rows []Record
}

func (w *Warehouse) Insert(r Record) { w.Rows = append(w.Rows, r) }

// Importer is the set of steps a format MUST provide. These are the pattern's
// "abstract" operations.
type Importer interface {
	Format() string
	Parse(raw []byte) ([]Record, error)
}

// Validator is an OPTIONAL hook. A format that does not implement it gets the
// default validation instead — the equivalent of not overriding a hook method.
type Validator interface {
	Validate(Record) error
}

// Deduper is another optional hook, for formats whose feed is known to repeat
// rows.
type Deduper interface {
	Dedupe([]Record) []Record
}

// ErrEmptyPayload is returned when there is nothing to import.
var ErrEmptyPayload = errors.New("templatemethod: empty payload")

// Run is the template method itself. This ordering is the invariant part of the
// algorithm and no format is allowed to change it.
func Run(imp Importer, raw []byte, w *Warehouse) (Report, error) {
	report := Report{Format: imp.Format()}

	// Step 1 — guard, identical for every format.
	if len(raw) == 0 {
		return report, ErrEmptyPayload
	}

	// Step 2 — parse: always delegated.
	records, err := imp.Parse(raw)
	if err != nil {
		return report, fmt.Errorf("parse %s: %w", imp.Format(), err)
	}
	report.Parsed = len(records)

	// Step 3 — dedupe: optional hook, with a default of "do nothing".
	if d, ok := imp.(Deduper); ok {
		before := len(records)
		records = d.Dedupe(records)
		report.Duplicates = before - len(records)
	}

	// Step 4 — validate: optional hook, with a real default.
	validate := defaultValidate
	if v, ok := imp.(Validator); ok {
		validate = v.Validate
	}

	// Step 5 — load. A bad row is skipped, not fatal: one malformed line should
	// not lose the whole nightly feed.
	for _, r := range records {
		if err := validate(r); err != nil {
			report.Rejected++
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", r.SKU, err))

			continue
		}
		w.Insert(r)
		report.Loaded++
	}

	return report, nil
}

// defaultValidate is the behaviour used when a format does not override it.
func defaultValidate(r Record) error {
	if r.SKU == "" {
		return errors.New("missing sku")
	}
	if r.Price < 0 {
		return errors.New("negative price")
	}

	return nil
}

// --- Concrete implementations ------------------------------------------------

// CSVImporter reads "sku,price" lines. It overrides both optional hooks.
type CSVImporter struct{}

func (CSVImporter) Format() string { return "csv" }

func (CSVImporter) Parse(raw []byte) ([]Record, error) {
	var out []Record

	for i, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) != 2 {
			return nil, fmt.Errorf("line %d: want 2 columns, got %d", i+1, len(parts))
		}

		price, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}

		out = append(out, Record{SKU: strings.TrimSpace(parts[0]), Price: price})
	}

	return out, nil
}

// Dedupe keeps the first occurrence of each SKU.
func (CSVImporter) Dedupe(in []Record) []Record {
	seen := make(map[string]bool, len(in))
	out := in[:0:0]

	for _, r := range in {
		if seen[r.SKU] {
			continue
		}
		seen[r.SKU] = true
		out = append(out, r)
	}

	return out
}

// Validate is stricter than the default: this supplier must not send free items.
func (CSVImporter) Validate(r Record) error {
	if err := defaultValidate(r); err != nil {
		return err
	}
	if r.Price == 0 {
		return errors.New("zero price not allowed for this supplier")
	}

	return nil
}

// JSONImporter overrides nothing optional, so it gets the default validation and
// no deduplication.
type JSONImporter struct{}

func (JSONImporter) Format() string { return "json" }

func (JSONImporter) Parse(raw []byte) ([]Record, error) {
	var rows []struct {
		SKU   string `json:"sku"`
		Price int    `json:"price"`
	}

	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}

	out := make([]Record, 0, len(rows))
	for _, r := range rows {
		out = append(out, Record{SKU: r.SKU, Price: r.Price})
	}

	return out, nil
}
