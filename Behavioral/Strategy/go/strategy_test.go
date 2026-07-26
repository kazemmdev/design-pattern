package strategy

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBackoffDelays(t *testing.T) {
	tests := []struct {
		name    string
		backoff Backoff
		attempt int
		want    time.Duration
	}{
		{"constant is flat", ConstantBackoff{Interval: 100 * time.Millisecond}, 1, 100 * time.Millisecond},
		{"constant ignores attempt", ConstantBackoff{Interval: 100 * time.Millisecond}, 7, 100 * time.Millisecond},
		{"constant guards attempt zero", ConstantBackoff{Interval: 100 * time.Millisecond}, 0, 0},

		{"exponential first attempt is base", ExponentialBackoff{Base: time.Second, Max: time.Minute}, 1, time.Second},
		{"exponential doubles", ExponentialBackoff{Base: time.Second, Max: time.Minute}, 3, 4 * time.Second},
		{"exponential clamps at max", ExponentialBackoff{Base: time.Second, Max: 10 * time.Second}, 10, 10 * time.Second},
		{"exponential survives huge attempts", ExponentialBackoff{Base: time.Second, Max: time.Minute}, 1000, time.Minute},

		{"jitter halves with fixed rand", JitteredBackoff{
			Inner: ConstantBackoff{Interval: time.Second},
			Rand:  func() float64 { return 0.5 },
		}, 1, 500 * time.Millisecond},
		{"jitter can collapse to zero", JitteredBackoff{
			Inner: ConstantBackoff{Interval: time.Second},
			Rand:  func() float64 { return 0 },
		}, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.backoff.Delay(tt.attempt); got != tt.want {
				t.Errorf("Delay(%d) = %v, want %v", tt.attempt, got, tt.want)
			}
		})
	}
}

func TestBackoffNames(t *testing.T) {
	got := JitteredBackoff{Inner: ExponentialBackoff{Base: time.Second}}.Name()
	if want := "jittered-exponential"; got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

// The point of the pattern: the Retrier is written once and the timing policy is
// swapped underneath it.
func TestRetrierUsesTheInjectedStrategy(t *testing.T) {
	tests := []struct {
		name       string
		backoff    Backoff
		wantDelays []time.Duration
	}{
		{
			name:       "constant",
			backoff:    ConstantBackoff{Interval: time.Second},
			wantDelays: []time.Duration{time.Second, time.Second},
		},
		{
			name:       "exponential",
			backoff:    ExponentialBackoff{Base: time.Second, Max: time.Minute},
			wantDelays: []time.Duration{time.Second, 2 * time.Second},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var slept []time.Duration
			r := &Retrier{
				Backoff:     tt.backoff,
				MaxAttempts: 3,
				Sleep:       func(d time.Duration) { slept = append(slept, d) },
			}

			err := r.Do(context.Background(), func() error { return errors.New("boom") })
			if !errors.Is(err, ErrExhausted) {
				t.Fatalf("expected ErrExhausted, got %v", err)
			}

			if len(slept) != len(tt.wantDelays) {
				t.Fatalf("slept %v, want %v", slept, tt.wantDelays)
			}
			for i := range slept {
				if slept[i] != tt.wantDelays[i] {
					t.Errorf("delay %d = %v, want %v", i, slept[i], tt.wantDelays[i])
				}
			}
		})
	}
}

func TestRetrierStopsOnSuccess(t *testing.T) {
	calls := 0
	r := &Retrier{
		Backoff:     ConstantBackoff{Interval: time.Second},
		MaxAttempts: 5,
		Sleep:       func(time.Duration) {},
	}

	err := r.Do(context.Background(), func() error {
		calls++
		if calls < 3 {
			return errors.New("not yet")
		}

		return nil
	})

	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if calls != 3 {
		t.Errorf("op called %d times, want 3", calls)
	}
	// Two failures means two waits, not three.
	if got := len(r.Delays()); got != 2 {
		t.Errorf("recorded %d delays, want 2", got)
	}
}

func TestRetrierDoesNotSleepAfterFinalAttempt(t *testing.T) {
	var slept []time.Duration
	r := &Retrier{
		Backoff:     ConstantBackoff{Interval: time.Second},
		MaxAttempts: 1,
		Sleep:       func(d time.Duration) { slept = append(slept, d) },
	}

	_ = r.Do(context.Background(), func() error { return errors.New("boom") })

	if len(slept) != 0 {
		t.Errorf("slept %v after the only attempt, want no sleeps", slept)
	}
}

func TestRetrierHonoursContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	r := &Retrier{
		Backoff:     ConstantBackoff{Interval: time.Second},
		MaxAttempts: 3,
		Sleep:       func(time.Duration) {},
	}

	called := false
	err := r.Do(ctx, func() error {
		called = true
		return errors.New("boom")
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if called {
		t.Error("op ran despite a cancelled context")
	}
}

func TestRetrierWrapsTheUnderlyingError(t *testing.T) {
	sentinel := errors.New("gateway timeout")
	r := &Retrier{
		Backoff:     ConstantBackoff{},
		MaxAttempts: 2,
		Sleep:       func(time.Duration) {},
	}

	err := r.Do(context.Background(), func() error { return sentinel })

	if !errors.Is(err, sentinel) {
		t.Errorf("caller cannot recover the cause: %v", err)
	}
}
