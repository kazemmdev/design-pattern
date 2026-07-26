package observer

import (
	"strings"
	"sync"
	"testing"
)

func TestPublishReachesEverySubscriber(t *testing.T) {
	d := NewDispatcher()
	audit := &AuditLog{}
	inventory := &Inventory{Units: 5}
	mailer := &Mailer{}

	d.Subscribe(OrderPaid, audit)
	d.Subscribe(OrderPaid, inventory)
	d.Subscribe(OrderPaid, mailer)

	if err := d.Publish(Event{Name: OrderPaid, OrderID: "A-1", Amount: 2500}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	if got := audit.Entries(); len(got) != 1 || got[0] != "order.paid:A-1" {
		t.Errorf("audit = %v", got)
	}
	if inventory.Units != 4 {
		t.Errorf("stock = %d, want 4", inventory.Units)
	}
	if len(mailer.Sent) != 1 {
		t.Errorf("mailer sent %d, want 1", len(mailer.Sent))
	}
}

func TestListenersOnlyReceiveTheirOwnEvent(t *testing.T) {
	d := NewDispatcher()
	paid := &AuditLog{}
	shipped := &AuditLog{}

	d.Subscribe(OrderPaid, paid)
	d.Subscribe(OrderShipped, shipped)

	_ = d.Publish(Event{Name: OrderPaid, OrderID: "A-1"})

	if len(paid.Entries()) != 1 {
		t.Errorf("paid listener got %d events, want 1", len(paid.Entries()))
	}
	if len(shipped.Entries()) != 0 {
		t.Errorf("shipped listener got %d events, want 0", len(shipped.Entries()))
	}
}

func TestUnsubscribeStopsDelivery(t *testing.T) {
	d := NewDispatcher()
	audit := &AuditLog{}

	unsubscribe := d.Subscribe(OrderPaid, audit)
	_ = d.Publish(Event{Name: OrderPaid, OrderID: "A-1"})

	unsubscribe()
	_ = d.Publish(Event{Name: OrderPaid, OrderID: "A-2"})

	if got := audit.Entries(); len(got) != 1 {
		t.Errorf("got %v, want only the pre-unsubscribe event", got)
	}
	if n := d.ListenerCount(OrderPaid); n != 0 {
		t.Errorf("listener count = %d, want 0", n)
	}
}

func TestUnsubscribeIsIdempotent(t *testing.T) {
	d := NewDispatcher()
	unsubscribeA := d.Subscribe(OrderPaid, &AuditLog{})
	d.Subscribe(OrderPaid, &AuditLog{})

	unsubscribeA()
	unsubscribeA() // must not remove the second listener

	if n := d.ListenerCount(OrderPaid); n != 1 {
		t.Errorf("listener count = %d, want 1", n)
	}
}

// A broken listener must not silently swallow the others' work.
func TestOneFailingListenerDoesNotBlockTheRest(t *testing.T) {
	d := NewDispatcher()
	mailer := &Mailer{Fail: true}
	audit := &AuditLog{}

	d.Subscribe(OrderPaid, mailer)
	d.Subscribe(OrderPaid, audit)

	err := d.Publish(Event{Name: OrderPaid, OrderID: "A-1"})

	if err == nil {
		t.Fatal("expected the mailer failure to surface")
	}
	if !strings.Contains(err.Error(), "smtp unavailable") {
		t.Errorf("error = %v, want it to mention the cause", err)
	}
	if len(audit.Entries()) != 1 {
		t.Error("later listener was skipped because an earlier one failed")
	}
}

func TestPublishWithNoSubscribersIsNotAnError(t *testing.T) {
	d := NewDispatcher()

	if err := d.Publish(Event{Name: OrderRefunded, OrderID: "A-1"}); err != nil {
		t.Errorf("got %v, want nil", err)
	}
}

// Listeners are notified outside the lock, so this must not deadlock.
func TestListenerMaySubscribeDuringPublish(t *testing.T) {
	d := NewDispatcher()

	d.Subscribe(OrderPaid, ListenerFunc(func(e Event) error {
		d.Subscribe(OrderShipped, &AuditLog{})
		return nil
	}))

	done := make(chan error, 1)
	go func() { done <- d.Publish(Event{Name: OrderPaid, OrderID: "A-1"}) }()

	if err := <-done; err != nil {
		t.Fatalf("publish: %v", err)
	}
	if n := d.ListenerCount(OrderShipped); n != 1 {
		t.Errorf("listener count = %d, want 1", n)
	}
}

func TestConcurrentSubscribeAndPublish(t *testing.T) {
	d := NewDispatcher()
	audit := &AuditLog{}
	d.Subscribe(OrderPaid, audit)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); d.Subscribe(OrderShipped, &AuditLog{}) }()
		go func() { defer wg.Done(); _ = d.Publish(Event{Name: OrderPaid, OrderID: "A-1"}) }()
	}
	wg.Wait()

	if got := len(audit.Entries()); got != 50 {
		t.Errorf("audit recorded %d events, want 50", got)
	}
}

func TestInventoryReportsOutOfStock(t *testing.T) {
	inventory := &Inventory{Units: 0}

	if err := inventory.Notify(Event{Name: OrderPaid}); err == nil {
		t.Error("expected an out-of-stock error, got nil")
	}
}
