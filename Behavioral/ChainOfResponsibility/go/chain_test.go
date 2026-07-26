package chain

import (
	"strings"
	"testing"
)

func tokens() map[string]string {
	return map[string]string{"good-token": "alice"}
}

func TestRequestReachesTheEndWhenEveryLinkPasses(t *testing.T) {
	head := Chain(
		&AuthHandler{ValidTokens: tokens()},
		&RateLimitHandler{Limit: 10},
		&BodyLimitHandler{MaxBytes: 100},
		&EchoHandler{},
	)

	got := head.Handle(&Request{Path: "/orders", ClientIP: "10.0.0.1", Token: "good-token"})

	if got.Status != 200 {
		t.Errorf("status = %d, want 200", got.Status)
	}
	if got.Body != "handled /orders" {
		t.Errorf("body = %q", got.Body)
	}
}

// Each link must be able to stop the chain on its own.
func TestAnyLinkCanRejectTheRequest(t *testing.T) {
	tests := []struct {
		name       string
		request    *Request
		wantStatus int
	}{
		{
			name:       "auth rejects a bad token",
			request:    &Request{Path: "/orders", ClientIP: "10.0.0.1", Token: "nope"},
			wantStatus: 401,
		},
		{
			name: "body limit rejects a large payload",
			request: &Request{
				Path: "/orders", ClientIP: "10.0.0.1", Token: "good-token",
				Body: strings.Repeat("x", 101),
			},
			wantStatus: 413,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			head := Chain(
				&AuthHandler{ValidTokens: tokens()},
				&RateLimitHandler{Limit: 10},
				&BodyLimitHandler{MaxBytes: 100},
				&EchoHandler{},
			)

			if got := head.Handle(tt.request); got.Status != tt.wantStatus {
				t.Errorf("status = %d, want %d", got.Status, tt.wantStatus)
			}
		})
	}
}

// A rejection must stop the chain — later links should never run.
func TestRejectionShortCircuitsTheRestOfTheChain(t *testing.T) {
	audit := &AuditHandler{}
	head := Chain(
		&AuthHandler{ValidTokens: tokens()},
		audit,
		&EchoHandler{},
	)

	got := head.Handle(&Request{Path: "/orders", ClientIP: "10.0.0.1", Token: "bad"})

	if got.Status != 401 {
		t.Fatalf("status = %d, want 401", got.Status)
	}
	if len(audit.Seen) != 0 {
		t.Errorf("audit ran after a rejection: %v", audit.Seen)
	}
}

func TestRateLimitTripsOnlyAfterTheLimit(t *testing.T) {
	head := Chain(
		&RateLimitHandler{Limit: 2},
		&EchoHandler{},
	)

	req := func() *Request { return &Request{Path: "/orders", ClientIP: "10.0.0.1"} }

	if got := head.Handle(req()); got.Status != 200 {
		t.Errorf("first request: status = %d, want 200", got.Status)
	}
	if got := head.Handle(req()); got.Status != 200 {
		t.Errorf("second request: status = %d, want 200", got.Status)
	}
	if got := head.Handle(req()); got.Status != 429 {
		t.Errorf("third request: status = %d, want 429", got.Status)
	}
}

func TestRateLimitIsPerClient(t *testing.T) {
	head := Chain(&RateLimitHandler{Limit: 1}, &EchoHandler{})

	_ = head.Handle(&Request{Path: "/a", ClientIP: "10.0.0.1"})
	got := head.Handle(&Request{Path: "/a", ClientIP: "10.0.0.2"})

	if got.Status != 200 {
		t.Errorf("status = %d, want 200 — a different client has its own budget", got.Status)
	}
}

// Reordering the links changes which rejection wins, without any link changing.
func TestOrderOfLinksDeterminesPrecedence(t *testing.T) {
	oversized := func() *Request {
		return &Request{Path: "/orders", ClientIP: "10.0.0.1", Token: "bad", Body: strings.Repeat("x", 200)}
	}

	authFirst := Chain(
		&AuthHandler{ValidTokens: tokens()},
		&BodyLimitHandler{MaxBytes: 10},
		&EchoHandler{},
	)
	if got := authFirst.Handle(oversized()); got.Status != 401 {
		t.Errorf("auth-first: status = %d, want 401", got.Status)
	}

	sizeFirst := Chain(
		&BodyLimitHandler{MaxBytes: 10},
		&AuthHandler{ValidTokens: tokens()},
		&EchoHandler{},
	)
	if got := sizeFirst.Handle(oversized()); got.Status != 413 {
		t.Errorf("size-first: status = %d, want 413", got.Status)
	}
}

// A chain that runs off the end without anyone answering must not panic.
func TestChainWithNoTerminalHandler(t *testing.T) {
	head := Chain(&AuditHandler{})

	got := head.Handle(&Request{Path: "/orders"})

	if got.Status != 404 {
		t.Errorf("status = %d, want 404", got.Status)
	}
}

func TestEmptyChainIsNil(t *testing.T) {
	if Chain() != nil {
		t.Error("expected nil for an empty chain")
	}
}

func TestHandlerMayEnrichTheRequest(t *testing.T) {
	audit := &AuditHandler{}
	head := Chain(
		&AuthHandler{ValidTokens: tokens()},
		audit,
		&EchoHandler{},
	)

	// Auth strips the trailing slash before forwarding.
	got := head.Handle(&Request{Path: "/orders/", ClientIP: "10.0.0.1", Token: "good-token"})

	if got.Body != "handled /orders" {
		t.Errorf("body = %q, want the normalised path", got.Body)
	}
	if len(audit.Seen) != 1 || audit.Seen[0] != "/orders" {
		t.Errorf("audit saw %v, want the normalised path", audit.Seen)
	}
}
