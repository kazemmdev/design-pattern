// Package observer demonstrates the Observer behavioral pattern.
//
// When an order is paid, several unrelated things must happen: email the
// customer, decrement stock, write an audit record, maybe notify a webhook. The
// checkout code should not know that list — every new side effect would mean
// editing it again. Observer inverts that: interested parties subscribe to the
// dispatcher, and checkout just publishes an event.
package observer

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

// Event names used by the order pipeline.
const (
	OrderPaid     = "order.paid"
	OrderShipped  = "order.shipped"
	OrderRefunded = "order.refunded"
)

// Event is the message passed to every listener.
type Event struct {
	Name    string
	OrderID string
	Amount  int // in minor units, e.g. cents
}

// Listener is the Observer interface.
type Listener interface {
	Notify(Event) error
}

// ListenerFunc lets a plain function act as a Listener.
type ListenerFunc func(Event) error

func (f ListenerFunc) Notify(e Event) error { return f(e) }

// Dispatcher is the Subject. It is safe for concurrent use, because in a real
// service publishers and subscribers live on different goroutines.
type Dispatcher struct {
	mu        sync.RWMutex
	nextID    int
	listeners map[string]map[int]Listener
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{listeners: make(map[string]map[int]Listener)}
}

// Subscribe registers l for eventName and returns the function that removes it
// again. Handing back the unsubscribe closure means the caller cannot leak a
// subscription by losing the reference to its listener.
func (d *Dispatcher) Subscribe(eventName string, l Listener) (unsubscribe func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.listeners[eventName] == nil {
		d.listeners[eventName] = make(map[int]Listener)
	}

	d.nextID++
	id := d.nextID
	d.listeners[eventName][id] = l

	var once sync.Once

	return func() {
		once.Do(func() {
			d.mu.Lock()
			defer d.mu.Unlock()
			delete(d.listeners[eventName], id)
		})
	}
}

// Publish notifies every listener registered for the event.
//
// Two details that matter in production: listeners are invoked *outside* the
// lock, so a listener may subscribe or unsubscribe without deadlocking; and one
// failing listener does not prevent the rest from running.
func (d *Dispatcher) Publish(e Event) error {
	d.mu.RLock()
	ids := make([]int, 0, len(d.listeners[e.Name]))
	snapshot := make(map[int]Listener, len(d.listeners[e.Name]))
	for id, l := range d.listeners[e.Name] {
		ids = append(ids, id)
		snapshot[id] = l
	}
	d.mu.RUnlock()

	// Map iteration order is random in Go; sorting by subscription id keeps
	// delivery order stable and therefore testable.
	sort.Ints(ids)

	var errs []error
	for _, id := range ids {
		if err := snapshot[id].Notify(e); err != nil {
			errs = append(errs, fmt.Errorf("listener %d: %w", id, err))
		}
	}

	return errors.Join(errs...)
}

// ListenerCount reports how many listeners are registered for an event.
func (d *Dispatcher) ListenerCount(eventName string) int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return len(d.listeners[eventName])
}

// --- Concrete observers -----------------------------------------------------

// AuditLog records every event it is shown. Append-only, so it never fails.
type AuditLog struct {
	mu      sync.Mutex
	entries []string
}

func (a *AuditLog) Notify(e Event) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.entries = append(a.entries, fmt.Sprintf("%s:%s", e.Name, e.OrderID))

	return nil
}

func (a *AuditLog) Entries() []string {
	a.mu.Lock()
	defer a.mu.Unlock()

	return append([]string(nil), a.entries...)
}

// Inventory decrements stock when an order is paid.
type Inventory struct {
	Units int
}

func (i *Inventory) Notify(e Event) error {
	if i.Units <= 0 {
		return errors.New("out of stock")
	}
	i.Units--

	return nil
}

// Mailer stands in for a transactional email provider.
type Mailer struct {
	Sent []string
	Fail bool
}

func (m *Mailer) Notify(e Event) error {
	if m.Fail {
		return errors.New("smtp unavailable")
	}
	m.Sent = append(m.Sent, e.OrderID)

	return nil
}
