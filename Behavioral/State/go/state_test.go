package state

import (
	"errors"
	"strings"
	"testing"
)

func TestHappyPath(t *testing.T) {
	o := NewOrder("A-1")

	if o.Status() != StatusPending {
		t.Fatalf("new order is %s, want pending", o.Status())
	}
	if err := o.Pay(); err != nil {
		t.Fatalf("pay: %v", err)
	}
	if err := o.Ship("TRACK-99"); err != nil {
		t.Fatalf("ship: %v", err)
	}
	if err := o.Deliver(); err != nil {
		t.Fatalf("deliver: %v", err)
	}

	if o.Status() != StatusDelivered {
		t.Errorf("status = %s, want delivered", o.Status())
	}
	if o.Tracking != "TRACK-99" {
		t.Errorf("tracking = %q", o.Tracking)
	}

	want := []Status{StatusPending, StatusPaid, StatusShipped, StatusDelivered}
	got := o.History()
	if len(got) != len(want) {
		t.Fatalf("history = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("history[%d] = %s, want %s", i, got[i], want[i])
		}
	}
}

// Every illegal move must be refused, and refused with a usable error.
func TestIllegalTransitions(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *Order
		act   func(*Order) error
	}{
		{
			name:  "cannot ship before paying",
			setup: func() *Order { return NewOrder("A-1") },
			act:   func(o *Order) error { return o.Ship("T-1") },
		},
		{
			name:  "cannot deliver before shipping",
			setup: func() *Order { return NewOrder("A-1") },
			act:   func(o *Order) error { return o.Deliver() },
		},
		{
			name: "cannot pay twice",
			setup: func() *Order {
				o := NewOrder("A-1")
				_ = o.Pay()
				return o
			},
			act: func(o *Order) error { return o.Pay() },
		},
		{
			name: "cannot cancel once shipped",
			setup: func() *Order {
				o := NewOrder("A-1")
				_ = o.Pay()
				_ = o.Ship("T-1")
				return o
			},
			act: func(o *Order) error { return o.Cancel("changed my mind") },
		},
		{
			name: "delivered is terminal",
			setup: func() *Order {
				o := NewOrder("A-1")
				_ = o.Pay()
				_ = o.Ship("T-1")
				_ = o.Deliver()
				return o
			},
			act: func(o *Order) error { return o.Cancel("too late") },
		},
		{
			name: "cancelled is terminal",
			setup: func() *Order {
				o := NewOrder("A-1")
				_ = o.Cancel("out of stock")
				return o
			},
			act: func(o *Order) error { return o.Pay() },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.setup()
			before := o.Status()

			err := tt.act(o)

			if !errors.Is(err, ErrInvalidTransition) {
				t.Fatalf("got %v, want ErrInvalidTransition", err)
			}
			if o.Status() != before {
				t.Errorf("status changed to %s despite the refusal", o.Status())
			}
			// A terminal state must still name itself in the message.
			if !strings.Contains(err.Error(), string(before)) {
				t.Errorf("error %q does not mention the current status %q", err, before)
			}
		})
	}
}

func TestPendingOrderCanBeCancelled(t *testing.T) {
	o := NewOrder("A-1")

	if err := o.Cancel("customer changed their mind"); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if o.Status() != StatusCancelled {
		t.Errorf("status = %s, want cancelled", o.Status())
	}
	if o.Reason != "customer changed their mind" {
		t.Errorf("reason = %q", o.Reason)
	}
}

// A paid order can still be cancelled — that is a refund, not a rejection.
func TestPaidOrderCanBeRefunded(t *testing.T) {
	o := NewOrder("A-1")
	if err := o.Pay(); err != nil {
		t.Fatalf("pay: %v", err)
	}

	if err := o.Cancel("refund requested"); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if o.Status() != StatusCancelled {
		t.Errorf("status = %s, want cancelled", o.Status())
	}
}

// A state may reject an action on grounds other than the transition itself.
func TestShippingRequiresATrackingNumber(t *testing.T) {
	o := NewOrder("A-1")
	_ = o.Pay()

	err := o.Ship("")

	if err == nil {
		t.Fatal("expected shipping without tracking to fail")
	}
	if errors.Is(err, ErrInvalidTransition) {
		t.Error("this is a validation failure, not an invalid transition")
	}
	if o.Status() != StatusPaid {
		t.Errorf("status = %s, want the order left as paid", o.Status())
	}
}

func TestHistoryIsACopy(t *testing.T) {
	o := NewOrder("A-1")
	h := o.History()
	h[0] = StatusDelivered

	if o.History()[0] != StatusPending {
		t.Error("History() exposed the order's internal slice")
	}
}
