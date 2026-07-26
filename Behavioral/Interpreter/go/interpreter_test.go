package interpreter

import (
	"errors"
	"testing"
)

func customer() Context {
	return Context{
		"plan":  "pro",
		"seats": 12,
		"trial": false,
		"país":  "PT",
	}
}

func TestMatch(t *testing.T) {
	tests := []struct {
		name string
		rule string
		want bool
	}{
		{"string equality", `plan = "pro"`, true},
		{"string inequality", `plan != "free"`, true},
		{"bare word value", `plan = pro`, true},
		{"failed equality", `plan = "free"`, false},

		{"numeric greater than", `seats > 10`, true},
		{"numeric less than", `seats < 10`, false},
		{"numeric boundary is exclusive", `seats > 12`, false},
		{"numeric boundary is inclusive", `seats >= 12`, true},
		{"negative numbers", `seats > -5`, true},

		{"bool equality", `trial = false`, true},
		{"bool inequality", `trial != true`, true},

		{"and both true", `plan = "pro" AND seats > 10`, true},
		{"and one false", `plan = "pro" AND seats > 100`, false},
		{"or one true", `plan = "free" OR seats > 10`, true},
		{"or both false", `plan = "free" OR seats > 100`, false},
		{"not", `NOT plan = "free"`, true},
		{"not inverts a true term", `NOT plan = "pro"`, false},

		{"and binds tighter than or", `plan = "free" AND seats > 100 OR seats = 12`, true},
		{"parentheses override precedence", `plan = "free" AND (seats > 100 OR seats = 12)`, false},
		{"nested parentheses", `(plan = "pro" AND (seats > 10 OR trial = true))`, true},
		{"not applies to a group", `NOT (plan = "free" OR seats < 5)`, true},

		{"lowercase keywords", `plan = "pro" and seats > 10`, true},
		{"non-ascii field name", `país = "PT"`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Match(tt.rule, customer())
			if err != nil {
				t.Fatalf("Match(%q): %v", tt.rule, err)
			}
			if got != tt.want {
				t.Errorf("Match(%q) = %v, want %v", tt.rule, got, tt.want)
			}
		})
	}
}

func TestSyntaxErrors(t *testing.T) {
	tests := []struct {
		name string
		rule string
	}{
		{"empty rule", ``},
		{"missing operator", `plan "pro"`},
		{"missing value", `plan =`},
		{"unclosed parenthesis", `(plan = "pro"`},
		{"unterminated string", `plan = "pro`},
		{"trailing junk", `plan = "pro" garbage`},
		{"dangling and", `plan = "pro" AND`},
		{"lone bang", `plan ! "pro"`},
		{"unexpected character", `plan = "pro" & seats > 1`},
		{"operator with no field", `= "pro"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Match(tt.rule, customer())

			if !errors.Is(err, ErrSyntax) {
				t.Errorf("Match(%q) err = %v, want ErrSyntax", tt.rule, err)
			}
		})
	}
}

func TestUnknownFieldIsReported(t *testing.T) {
	_, err := Match(`nonexistent = "x"`, customer())

	if !errors.Is(err, ErrUnknownField) {
		t.Errorf("got %v, want ErrUnknownField", err)
	}
}

func TestTypeMismatchIsReported(t *testing.T) {
	tests := []struct {
		name string
		rule string
	}{
		{"number against string field", `plan > 5`},
		{"string against number field", `seats = "many"`},
		{"string against bool field", `trial = "no"`},
		{"ordering a bool", `trial > false`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Match(tt.rule, customer())

			if !errors.Is(err, ErrTypeMismatch) {
				t.Errorf("Match(%q) err = %v, want ErrTypeMismatch", tt.rule, err)
			}
		})
	}
}

// AND and OR short-circuit, so a term that would error is never reached.
func TestShortCircuitEvaluation(t *testing.T) {
	// The right side references an unknown field, but AND stops on the false
	// left side before it is evaluated.
	got, err := Match(`plan = "free" AND missing = 1`, customer())
	if err != nil {
		t.Fatalf("AND should have short-circuited: %v", err)
	}
	if got {
		t.Error("expected false")
	}

	got, err = Match(`plan = "pro" OR missing = 1`, customer())
	if err != nil {
		t.Fatalf("OR should have short-circuited: %v", err)
	}
	if !got {
		t.Error("expected true")
	}
}

// A parsed rule is reusable: parse once, evaluate against many contexts.
func TestParsedRuleIsReusable(t *testing.T) {
	rule, err := Parse(`plan = "pro" AND seats >= 10`)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	cases := []struct {
		ctx  Context
		want bool
	}{
		{Context{"plan": "pro", "seats": 12}, true},
		{Context{"plan": "pro", "seats": 3}, false},
		{Context{"plan": "free", "seats": 99}, false},
	}

	for i, c := range cases {
		got, err := rule.Eval(c.ctx)
		if err != nil {
			t.Fatalf("case %d: %v", i, err)
		}
		if got != c.want {
			t.Errorf("case %d = %v, want %v", i, got, c.want)
		}
	}
}

// String() renders the tree with its precedence made explicit — useful for
// showing an admin how their rule was actually understood.
func TestStringShowsTheParsedStructure(t *testing.T) {
	tests := []struct {
		rule string
		want string
	}{
		{`a = 1 AND b = 2 OR c = 3`, `((a = 1 AND b = 2) OR c = 3)`},
		{`a = 1 AND (b = 2 OR c = 3)`, `(a = 1 AND (b = 2 OR c = 3))`},
		{`NOT a = 1`, `NOT a = 1`},
	}

	for _, tt := range tests {
		t.Run(tt.rule, func(t *testing.T) {
			e, err := Parse(tt.rule)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if got := e.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
