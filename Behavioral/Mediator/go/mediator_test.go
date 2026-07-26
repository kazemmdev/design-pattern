package mediator

import (
	"strings"
	"testing"
)

func setup() (*Checkout, *Inventory, *Payments, *Shipping, *Notifier) {
	inv := NewInventory(3)
	pay := &Payments{}
	ship := &Shipping{}
	note := &Notifier{}

	return NewCheckout(inv, pay, ship, note), inv, pay, ship, note
}

func TestSuccessfulCheckoutCascadesThroughEveryComponent(t *testing.T) {
	c, inv, pay, ship, note := setup()

	if err := c.Place("A-1"); err != nil {
		t.Fatalf("place: %v", err)
	}

	if inv.Units != 2 {
		t.Errorf("units = %d, want 2", inv.Units)
	}
	if len(pay.Charged) != 1 {
		t.Errorf("charged = %v, want one charge", pay.Charged)
	}
	if len(ship.Dispatched) != 1 {
		t.Errorf("dispatched = %v, want one dispatch", ship.Dispatched)
	}
	if len(note.Sent) != 1 || !strings.Contains(note.Sent[0], "on its way") {
		t.Errorf("notifications = %v", note.Sent)
	}
}

// One Place call triggers the whole chain, but no component called another
// directly — every hop went through the mediator.
func TestEveryHopGoesThroughTheMediator(t *testing.T) {
	c, _, _, _, _ := setup()

	if err := c.Place("A-1"); err != nil {
		t.Fatalf("place: %v", err)
	}

	want := []string{
		"inventory -> stock.reserved",
		"payments -> payment.charged",
		"shipping -> shipment.dispatched",
	}
	if len(c.Log) != len(want) {
		t.Fatalf("log = %v, want %v", c.Log, want)
	}
	for i := range want {
		if c.Log[i] != want[i] {
			t.Errorf("log[%d] = %q, want %q", i, c.Log[i], want[i])
		}
	}
}

func TestOutOfStockStopsBeforeCharging(t *testing.T) {
	inv := NewInventory(0)
	pay := &Payments{}
	ship := &Shipping{}
	note := &Notifier{}
	c := NewCheckout(inv, pay, ship, note)

	err := c.Place("A-1")

	if err == nil {
		t.Fatal("expected an out-of-stock error")
	}
	if len(pay.Charged) != 0 {
		t.Error("customer was charged despite no stock")
	}
	if len(ship.Dispatched) != 0 {
		t.Error("parcel was dispatched despite no stock")
	}
}

// A declined card must give the reserved stock back.
func TestDeclinedPaymentReleasesStock(t *testing.T) {
	c, inv, pay, ship, note := setup()
	pay.Fail = true

	err := c.Place("A-1")

	if err == nil {
		t.Fatal("expected a payment failure")
	}
	if inv.Units != 3 {
		t.Errorf("units = %d, want the reservation released back to 3", inv.Units)
	}
	if inv.Reserved["A-1"] != 0 {
		t.Errorf("reserved = %d, want 0", inv.Reserved["A-1"])
	}
	if len(ship.Dispatched) != 0 {
		t.Error("parcel dispatched despite payment failure")
	}
	if len(note.Sent) != 1 || !strings.Contains(note.Sent[0], "declined") {
		t.Errorf("notifications = %v", note.Sent)
	}
}

// A shipping failure must both refund and release — the mediator owns that
// compensation logic, not the components.
func TestShippingFailureRefundsAndReleases(t *testing.T) {
	c, inv, pay, ship, note := setup()
	ship.Fail = true

	err := c.Place("A-1")

	if err == nil {
		t.Fatal("expected a shipping failure")
	}
	if len(pay.Refunded) != 1 {
		t.Errorf("refunded = %v, want one refund", pay.Refunded)
	}
	if inv.Units != 3 {
		t.Errorf("units = %d, want 3 after release", inv.Units)
	}
	if len(note.Sent) != 1 || !strings.Contains(note.Sent[0], "refunded") {
		t.Errorf("notifications = %v", note.Sent)
	}
}

func TestUnknownEventIsReported(t *testing.T) {
	c, _, _, _, _ := setup()

	err := c.Notify("somewhere", Event("nonsense"), "A-1")

	if err == nil {
		t.Fatal("expected an error for an unknown event")
	}
	if !strings.Contains(err.Error(), "nonsense") {
		t.Errorf("error = %v, want it to name the event", err)
	}
}

func TestReleaseNeverOverCredits(t *testing.T) {
	inv := NewInventory(1)

	// Releasing something never reserved must not invent stock.
	inv.Release("ghost")

	if inv.Units != 1 {
		t.Errorf("units = %d, want 1", inv.Units)
	}
}

func TestMultipleOrdersDrawDownStock(t *testing.T) {
	c, inv, _, _, _ := setup()

	for _, id := range []string{"A-1", "A-2", "A-3"} {
		if err := c.Place(id); err != nil {
			t.Fatalf("place %s: %v", id, err)
		}
	}

	if inv.Units != 0 {
		t.Errorf("units = %d, want 0", inv.Units)
	}
	if err := c.Place("A-4"); err == nil {
		t.Error("expected the fourth order to fail")
	}
}
