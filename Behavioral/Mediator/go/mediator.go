// Package mediator demonstrates the Mediator behavioral pattern.
//
// Checkout touches inventory, payments, shipping and notifications. If each
// service calls the others directly you get a mesh: payments needs to know about
// shipping, shipping needs to know about notifications, and a change to one
// ripples through all of them. A Mediator puts one coordinator in the middle —
// every service talks only to it, so they stay unaware of each other.
package mediator

import (
	"errors"
	"fmt"
)

// Event is what a component reports to the mediator.
type Event string

const (
	EventReserved   Event = "stock.reserved"
	EventCharged    Event = "payment.charged"
	EventDispatched Event = "shipment.dispatched"
)

// Mediator is what components see. They can raise events, and nothing else.
type Mediator interface {
	Notify(sender string, e Event, orderID string) error
}

// --- Colleagues --------------------------------------------------------------
//
// Each component holds a Mediator, never another component.

// Inventory reserves and releases stock.
type Inventory struct {
	mediator Mediator
	Units    int
	Reserved map[string]int
}

func NewInventory(units int) *Inventory {
	return &Inventory{Units: units, Reserved: make(map[string]int)}
}

func (i *Inventory) SetMediator(m Mediator) { i.mediator = m }

func (i *Inventory) Reserve(orderID string) error {
	if i.Units < 1 {
		return errors.New("mediator: out of stock")
	}
	i.Units--
	i.Reserved[orderID]++

	return i.mediator.Notify("inventory", EventReserved, orderID)
}

// Release puts stock back. Called by the mediator when a later step fails.
func (i *Inventory) Release(orderID string) {
	if i.Reserved[orderID] > 0 {
		i.Reserved[orderID]--
		i.Units++
	}
}

// Payments charges the customer.
type Payments struct {
	mediator Mediator
	Charged  []string
	Refunded []string
	Fail     bool
}

func (p *Payments) SetMediator(m Mediator) { p.mediator = m }

func (p *Payments) Charge(orderID string) error {
	if p.Fail {
		return fmt.Errorf("mediator: card declined for %s", orderID)
	}
	p.Charged = append(p.Charged, orderID)

	return p.mediator.Notify("payments", EventCharged, orderID)
}

func (p *Payments) Refund(orderID string) {
	p.Refunded = append(p.Refunded, orderID)
}

// Shipping dispatches the parcel.
type Shipping struct {
	mediator   Mediator
	Dispatched []string
	Fail       bool
}

func (s *Shipping) SetMediator(m Mediator) { s.mediator = m }

func (s *Shipping) Dispatch(orderID string) error {
	if s.Fail {
		return fmt.Errorf("mediator: no courier available for %s", orderID)
	}
	s.Dispatched = append(s.Dispatched, orderID)

	return s.mediator.Notify("shipping", EventDispatched, orderID)
}

// Notifier emails the customer. It only ever receives instructions.
type Notifier struct {
	Sent []string
}

func (n *Notifier) Send(orderID, message string) {
	n.Sent = append(n.Sent, orderID+": "+message)
}

// --- Concrete mediator -------------------------------------------------------

// Checkout is the Mediator. It is the only place that knows the order of the
// steps and what to unwind when one of them fails.
type Checkout struct {
	Inventory *Inventory
	Payments  *Payments
	Shipping  *Shipping
	Notifier  *Notifier

	Log []string

	// failed guards the compensation logic. The failure of a late step travels
	// back up through the notification cascade, so every level would otherwise
	// run its own rollback and send its own email. Only the innermost failure
	// — the one that actually knows what went wrong — should act.
	failed bool
}

func NewCheckout(inv *Inventory, pay *Payments, ship *Shipping, note *Notifier) *Checkout {
	c := &Checkout{Inventory: inv, Payments: pay, Shipping: ship, Notifier: note}

	// Wire the colleagues to the mediator, not to each other.
	inv.SetMediator(c)
	pay.SetMediator(c)
	ship.SetMediator(c)

	return c
}

// Notify is the single point where cross-component reactions are decided.
func (c *Checkout) Notify(sender string, e Event, orderID string) error {
	c.Log = append(c.Log, fmt.Sprintf("%s -> %s", sender, e))

	switch e {
	case EventReserved:
		// Stock is held; take the money next.
		if err := c.Payments.Charge(orderID); err != nil {
			// Compensate: give the held stock back.
			c.compensate(orderID, "payment declined", func() {
				c.Inventory.Release(orderID)
			})

			return err
		}

		return nil

	case EventCharged:
		// Paid; try to dispatch.
		if err := c.Shipping.Dispatch(orderID); err != nil {
			// Compensate: refund and release the stock we were holding.
			c.compensate(orderID, "payment refunded", func() {
				c.Payments.Refund(orderID)
				c.Inventory.Release(orderID)
			})

			return err
		}

		return nil

	case EventDispatched:
		c.Notifier.Send(orderID, "your order is on its way")

		return nil
	}

	return fmt.Errorf("mediator: unknown event %q from %s", e, sender)
}

// compensate rolls the order back and tells the customer — but only for the
// first failure seen, so an unwinding cascade does not repeat the work.
func (c *Checkout) compensate(orderID, reason string, undo func()) {
	if c.failed {
		return
	}
	c.failed = true

	undo()
	c.Notifier.Send(orderID, "checkout failed, "+reason)
}

// Place starts the flow. The cascade after this is driven entirely by Notify.
func (c *Checkout) Place(orderID string) error {
	c.failed = false

	if err := c.Inventory.Reserve(orderID); err != nil {
		// Nothing was reserved, so there is nothing to undo.
		c.compensate(orderID, "out of stock", func() {})

		return err
	}

	return nil
}
