// Package state demonstrates the State behavioral pattern.
//
// An order moves through a lifecycle, and what it is allowed to do depends
// entirely on where it currently is: you can cancel a pending order but not a
// delivered one, you can only ship what has been paid for. Written as flags and
// if-statements this becomes an unreadable knot. State gives each stage its own
// type, so the rules live next to the stage they belong to.
package state

import (
	"errors"
	"fmt"
)

// Status names each stage of the lifecycle.
type Status string

const (
	StatusPending   Status = "pending"
	StatusPaid      Status = "paid"
	StatusShipped   Status = "shipped"
	StatusDelivered Status = "delivered"
	StatusCancelled Status = "cancelled"
)

// ErrInvalidTransition is returned when an action is not legal in the current
// state. Callers can test for it with errors.Is.
var ErrInvalidTransition = errors.New("state: invalid transition")

// State is the interface each lifecycle stage implements.
type State interface {
	Status() Status
	Pay(*Order) error
	Ship(*Order, string) error
	Deliver(*Order) error
	Cancel(*Order, string) error
}

// Order is the Context. It holds the current state and forwards every action to
// it — note there is not a single switch statement in this type.
type Order struct {
	ID       string
	Tracking string
	Reason   string

	state   State
	history []Status
}

func NewOrder(id string) *Order {
	return &Order{
		ID:      id,
		state:   pendingState{base{StatusPending}},
		history: []Status{StatusPending},
	}
}

func (o *Order) Status() Status { return o.state.Status() }

// History reports every status the order has held, in order.
func (o *Order) History() []Status { return append([]Status(nil), o.history...) }

// transition is the only place the current state is replaced.
func (o *Order) transition(s State) {
	o.state = s
	o.history = append(o.history, s.Status())
}

func (o *Order) Pay() error                 { return o.state.Pay(o) }
func (o *Order) Ship(tracking string) error { return o.state.Ship(o, tracking) }
func (o *Order) Deliver() error             { return o.state.Deliver(o) }
func (o *Order) Cancel(reason string) error { return o.state.Cancel(o, reason) }

// --- Concrete states --------------------------------------------------------

// base refuses every action and reports the status it was built with. Each
// concrete state embeds it and overrides only the transitions it permits, so a
// newly added action defaults to "not allowed" rather than silently succeeding.
type base struct{ status Status }

func (b base) Status() Status { return b.status }

func (b base) Pay(*Order) error { return b.refuse("pay") }

func (b base) Ship(*Order, string) error { return b.refuse("ship") }

func (b base) Deliver(*Order) error { return b.refuse("deliver") }

func (b base) Cancel(*Order, string) error { return b.refuse("cancel") }

func (b base) refuse(action string) error {
	return fmt.Errorf("%w: cannot %s an order that is %s", ErrInvalidTransition, action, b.status)
}

// pendingState: awaiting payment. Can be paid or cancelled.
type pendingState struct{ base }

func (pendingState) Pay(o *Order) error {
	o.transition(paidState{base{StatusPaid}})

	return nil
}

func (pendingState) Cancel(o *Order, reason string) error {
	o.Reason = reason
	o.transition(cancelledState{base{StatusCancelled}})

	return nil
}

// paidState: money taken, not yet dispatched. Can be shipped, or cancelled as a
// refund.
type paidState struct{ base }

func (paidState) Ship(o *Order, tracking string) error {
	if tracking == "" {
		return errors.New("state: a tracking number is required to ship")
	}
	o.Tracking = tracking
	o.transition(shippedState{base{StatusShipped}})

	return nil
}

func (paidState) Cancel(o *Order, reason string) error {
	o.Reason = reason
	o.transition(cancelledState{base{StatusCancelled}})

	return nil
}

// shippedState: with the courier. Can only be delivered — too late to cancel.
type shippedState struct{ base }

func (shippedState) Deliver(o *Order) error {
	o.transition(deliveredState{base{StatusDelivered}})

	return nil
}

// deliveredState and cancelledState are terminal: they permit nothing, so they
// add no overrides at all and inherit base's refusals.
type deliveredState struct{ base }

type cancelledState struct{ base }
