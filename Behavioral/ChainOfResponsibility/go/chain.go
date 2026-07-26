// Package chain demonstrates the Chain of Responsibility behavioral pattern.
//
// Before a request reaches the code that actually answers it, a series of checks
// must pass: is the caller authenticated, are they within their rate limit, is
// the payload small enough. Each check can either reject the request outright or
// pass it along. Wiring them as a chain means any of them can be added, removed
// or reordered without touching the others or the handler at the end.
package chain

import (
	"fmt"
	"strings"
)

// Request is the thing travelling down the chain.
type Request struct {
	Path     string
	ClientIP string
	Token    string
	Body     string
}

// Response is what comes back out.
type Response struct {
	Status int
	Body   string
}

// Handler is the link interface.
type Handler interface {
	// SetNext links the following handler and returns it, so calls can be
	// chained fluently.
	SetNext(Handler) Handler
	Handle(*Request) Response
}

// base holds the "next" pointer and the pass-along logic every link shares.
type base struct {
	next Handler
}

func (b *base) SetNext(h Handler) Handler {
	b.next = h

	return h
}

// forward passes the request on, or returns 404 if this was the last link.
func (b *base) forward(r *Request) Response {
	if b.next == nil {
		return Response{Status: 404, Body: "no handler"}
	}

	return b.next.Handle(r)
}

// --- Concrete handlers -------------------------------------------------------

// AuthHandler rejects requests without a recognised token.
type AuthHandler struct {
	base
	ValidTokens map[string]string // token -> user
}

func (h *AuthHandler) Handle(r *Request) Response {
	user, ok := h.ValidTokens[r.Token]
	if !ok {
		// Reject: the chain stops here.
		return Response{Status: 401, Body: "unauthorized"}
	}

	// Handlers may enrich the request before passing it on.
	r.Path = strings.TrimSuffix(r.Path, "/")
	_ = user

	return h.forward(r)
}

// RateLimitHandler allows at most Limit requests per client IP.
type RateLimitHandler struct {
	base
	Limit  int
	counts map[string]int
}

func (h *RateLimitHandler) Handle(r *Request) Response {
	if h.counts == nil {
		h.counts = make(map[string]int)
	}

	h.counts[r.ClientIP]++
	if h.counts[r.ClientIP] > h.Limit {
		return Response{Status: 429, Body: "rate limit exceeded"}
	}

	return h.forward(r)
}

// BodyLimitHandler rejects oversized payloads.
type BodyLimitHandler struct {
	base
	MaxBytes int
}

func (h *BodyLimitHandler) Handle(r *Request) Response {
	if len(r.Body) > h.MaxBytes {
		return Response{
			Status: 413,
			Body:   fmt.Sprintf("payload too large: %d > %d", len(r.Body), h.MaxBytes),
		}
	}

	return h.forward(r)
}

// AuditHandler records what passed through it, then always forwards. A handler
// that never rejects is perfectly legitimate.
type AuditHandler struct {
	base
	Seen []string
}

func (h *AuditHandler) Handle(r *Request) Response {
	h.Seen = append(h.Seen, r.Path)

	return h.forward(r)
}

// EchoHandler is the terminal link: it answers instead of forwarding.
type EchoHandler struct {
	base
}

func (h *EchoHandler) Handle(r *Request) Response {
	return Response{Status: 200, Body: "handled " + r.Path}
}

// Chain builds a chain from handlers given in order and returns its head.
func Chain(handlers ...Handler) Handler {
	if len(handlers) == 0 {
		return nil
	}

	for i := 0; i < len(handlers)-1; i++ {
		handlers[i].SetNext(handlers[i+1])
	}

	return handlers[0]
}
