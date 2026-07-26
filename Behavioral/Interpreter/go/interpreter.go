// Package interpreter demonstrates the Interpreter behavioral pattern.
//
// Support wants to segment users with rules like
//
//	plan = "pro" AND (seats > 10 OR trial = false)
//
// without shipping a release for every new rule. Interpreter defines a small
// grammar, represents each construct as a node type, and gives every node an
// Eval method. Parsing turns the text into a tree; evaluating the tree answers
// the question. Adding an operator means adding a node, not editing a switch.
package interpreter

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Context supplies the values a rule is evaluated against. Values may be
// string, int or bool.
type Context map[string]any

// Errors callers can test for.
var (
	ErrSyntax       = errors.New("interpreter: syntax error")
	ErrUnknownField = errors.New("interpreter: unknown field")
	ErrTypeMismatch = errors.New("interpreter: type mismatch")
)

// Expr is the abstract expression. Every node in the tree implements it.
type Expr interface {
	Eval(Context) (bool, error)
	String() string
}

// --- Terminal expression -----------------------------------------------------

// Comparison is the terminal node: one field tested against one literal.
type Comparison struct {
	Field string
	Op    string
	Value any
}

func (c *Comparison) Eval(ctx Context) (bool, error) {
	actual, ok := ctx[c.Field]
	if !ok {
		return false, fmt.Errorf("%w: %q", ErrUnknownField, c.Field)
	}

	return compare(actual, c.Op, c.Value)
}

func (c *Comparison) String() string {
	return fmt.Sprintf("%s %s %v", c.Field, c.Op, c.Value)
}

func compare(left any, op string, right any) (bool, error) {
	switch l := left.(type) {
	case int:
		r, ok := right.(int)
		if !ok {
			return false, fmt.Errorf("%w: cannot compare number with %T", ErrTypeMismatch, right)
		}

		return compareOrdered(l, r, op)

	case string:
		r, ok := right.(string)
		if !ok {
			return false, fmt.Errorf("%w: cannot compare string with %T", ErrTypeMismatch, right)
		}

		return compareOrdered(l, r, op)

	case bool:
		r, ok := right.(bool)
		if !ok {
			return false, fmt.Errorf("%w: cannot compare bool with %T", ErrTypeMismatch, right)
		}
		switch op {
		case "=":
			return l == r, nil
		case "!=":
			return l != r, nil
		default:
			return false, fmt.Errorf("%w: operator %q is not defined for bool", ErrTypeMismatch, op)
		}

	default:
		return false, fmt.Errorf("%w: unsupported value type %T", ErrTypeMismatch, left)
	}
}

func compareOrdered[T int | string](l, r T, op string) (bool, error) {
	switch op {
	case "=":
		return l == r, nil
	case "!=":
		return l != r, nil
	case ">":
		return l > r, nil
	case ">=":
		return l >= r, nil
	case "<":
		return l < r, nil
	case "<=":
		return l <= r, nil
	default:
		return false, fmt.Errorf("%w: unknown operator %q", ErrSyntax, op)
	}
}

// --- Non-terminal expressions ------------------------------------------------

// And evaluates to true only when both sides do. It short-circuits, so a rule
// can guard an expensive or error-prone term behind a cheap one.
type And struct{ Left, Right Expr }

func (e *And) Eval(ctx Context) (bool, error) {
	l, err := e.Left.Eval(ctx)
	if err != nil {
		return false, err
	}
	if !l {
		return false, nil
	}

	return e.Right.Eval(ctx)
}

func (e *And) String() string { return "(" + e.Left.String() + " AND " + e.Right.String() + ")" }

// Or evaluates to true when either side does, short-circuiting on the left.
type Or struct{ Left, Right Expr }

func (e *Or) Eval(ctx Context) (bool, error) {
	l, err := e.Left.Eval(ctx)
	if err != nil {
		return false, err
	}
	if l {
		return true, nil
	}

	return e.Right.Eval(ctx)
}

func (e *Or) String() string { return "(" + e.Left.String() + " OR " + e.Right.String() + ")" }

// Not inverts its operand.
type Not struct{ Inner Expr }

func (e *Not) Eval(ctx Context) (bool, error) {
	v, err := e.Inner.Eval(ctx)
	if err != nil {
		return false, err
	}

	return !v, nil
}

func (e *Not) String() string { return "NOT " + e.Inner.String() }

// --- Lexer -------------------------------------------------------------------

type tokenKind int

const (
	tokEOF tokenKind = iota
	tokIdent
	tokNumber
	tokString
	tokBool
	tokOp
	tokLParen
	tokRParen
	tokAnd
	tokOr
	tokNot
)

type token struct {
	kind tokenKind
	text string
}

func lex(input string) ([]token, error) {
	var out []token
	runes := []rune(input)

	for i := 0; i < len(runes); {
		ch := runes[i]

		switch {
		case unicode.IsSpace(ch):
			i++

		case ch == '(':
			out = append(out, token{tokLParen, "("})
			i++

		case ch == ')':
			out = append(out, token{tokRParen, ")"})
			i++

		case ch == '"':
			j := i + 1
			for j < len(runes) && runes[j] != '"' {
				j++
			}
			if j >= len(runes) {
				return nil, fmt.Errorf("%w: unterminated string", ErrSyntax)
			}
			out = append(out, token{tokString, string(runes[i+1 : j])})
			i = j + 1

		case strings.ContainsRune("<>=!", ch):
			// Two-character operators first, so ">=" is not read as ">" then "=".
			if i+1 < len(runes) && runes[i+1] == '=' {
				out = append(out, token{tokOp, string(runes[i : i+2])})
				i += 2

				continue
			}
			if ch == '!' {
				return nil, fmt.Errorf("%w: expected != but found !", ErrSyntax)
			}
			out = append(out, token{tokOp, string(ch)})
			i++

		case unicode.IsDigit(ch) || (ch == '-' && i+1 < len(runes) && unicode.IsDigit(runes[i+1])):
			j := i + 1
			for j < len(runes) && unicode.IsDigit(runes[j]) {
				j++
			}
			out = append(out, token{tokNumber, string(runes[i:j])})
			i = j

		case unicode.IsLetter(ch) || ch == '_':
			j := i
			for j < len(runes) && (unicode.IsLetter(runes[j]) || unicode.IsDigit(runes[j]) || runes[j] == '_' || runes[j] == '.') {
				j++
			}
			word := string(runes[i:j])
			i = j

			switch strings.ToUpper(word) {
			case "AND":
				out = append(out, token{tokAnd, word})
			case "OR":
				out = append(out, token{tokOr, word})
			case "NOT":
				out = append(out, token{tokNot, word})
			case "TRUE", "FALSE":
				out = append(out, token{tokBool, strings.ToLower(word)})
			default:
				out = append(out, token{tokIdent, word})
			}

		default:
			return nil, fmt.Errorf("%w: unexpected character %q", ErrSyntax, string(ch))
		}
	}

	return append(out, token{tokEOF, ""}), nil
}

// --- Parser ------------------------------------------------------------------

type parser struct {
	tokens []token
	pos    int
}

func (p *parser) peek() token { return p.tokens[p.pos] }

func (p *parser) next() token {
	t := p.tokens[p.pos]
	if p.pos < len(p.tokens)-1 {
		p.pos++
	}

	return t
}

// Parse turns rule text into an expression tree.
//
// The grammar, lowest precedence first:
//
//	expr       := term ( "OR" term )*
//	term       := factor ( "AND" factor )*
//	factor     := "NOT" factor | "(" expr ")" | comparison
//	comparison := IDENT op literal
func Parse(input string) (Expr, error) {
	tokens, err := lex(input)
	if err != nil {
		return nil, err
	}

	p := &parser{tokens: tokens}

	e, err := p.parseExpr()
	if err != nil {
		return nil, err
	}

	if p.peek().kind != tokEOF {
		return nil, fmt.Errorf("%w: unexpected trailing input %q", ErrSyntax, p.peek().text)
	}

	return e, nil
}

func (p *parser) parseExpr() (Expr, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	for p.peek().kind == tokOr {
		p.next()
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		left = &Or{Left: left, Right: right}
	}

	return left, nil
}

func (p *parser) parseTerm() (Expr, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}

	for p.peek().kind == tokAnd {
		p.next()
		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		left = &And{Left: left, Right: right}
	}

	return left, nil
}

func (p *parser) parseFactor() (Expr, error) {
	switch p.peek().kind {
	case tokNot:
		p.next()
		inner, err := p.parseFactor()
		if err != nil {
			return nil, err
		}

		return &Not{Inner: inner}, nil

	case tokLParen:
		p.next()
		inner, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.peek().kind != tokRParen {
			return nil, fmt.Errorf("%w: missing closing parenthesis", ErrSyntax)
		}
		p.next()

		return inner, nil

	case tokIdent:
		return p.parseComparison()

	case tokEOF:
		return nil, fmt.Errorf("%w: unexpected end of input", ErrSyntax)

	default:
		return nil, fmt.Errorf("%w: unexpected token %q", ErrSyntax, p.peek().text)
	}
}

func (p *parser) parseComparison() (Expr, error) {
	field := p.next().text

	if p.peek().kind != tokOp {
		return nil, fmt.Errorf("%w: expected an operator after %q", ErrSyntax, field)
	}
	op := p.next().text

	lit := p.next()
	switch lit.kind {
	case tokNumber:
		n, err := strconv.Atoi(lit.text)
		if err != nil {
			return nil, fmt.Errorf("%w: bad number %q", ErrSyntax, lit.text)
		}

		return &Comparison{Field: field, Op: op, Value: n}, nil

	case tokString:
		return &Comparison{Field: field, Op: op, Value: lit.text}, nil

	case tokBool:
		return &Comparison{Field: field, Op: op, Value: lit.text == "true"}, nil

	case tokIdent:
		// A bare word is treated as a string, so plan = pro works too.
		return &Comparison{Field: field, Op: op, Value: lit.text}, nil

	default:
		return nil, fmt.Errorf("%w: expected a value after %q %s", ErrSyntax, field, op)
	}
}

// Match is the convenience entry point: parse a rule and evaluate it in one go.
func Match(rule string, ctx Context) (bool, error) {
	e, err := Parse(rule)
	if err != nil {
		return false, err
	}

	return e.Eval(ctx)
}
