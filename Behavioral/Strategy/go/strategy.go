// Package strategy demonstrates the Strategy behavioral pattern.
//
// A Retrier needs to wait between failed attempts, but "how long to wait" is a
// policy that changes per caller: a health check wants a short constant delay, a
// payment gateway wants exponential backoff, a flaky third-party API wants
// jitter so a fleet of clients doesn't retry in lockstep. Strategy pulls that
// decision out into interchangeable objects.
package strategy

import (
	"context"
	"errors"
	"math"
	"time"
)

// Backoff is the Strategy interface. It answers one question: how long should we
// wait before attempt number n? Attempts are 1-based.
type Backoff interface {
	Name() string
	Delay(attempt int) time.Duration
}

// ConstantBackoff waits the same amount of time before every retry. Good for
// fast internal calls where the failure is expected to be transient.
type ConstantBackoff struct {
	Interval time.Duration
}

func (b ConstantBackoff) Name() string { return "constant" }

func (b ConstantBackoff) Delay(attempt int) time.Duration {
	if attempt < 1 {
		return 0
	}

	return b.Interval
}

// ExponentialBackoff doubles the wait on each attempt, capped at Max. This is
// the default policy for anything crossing a network boundary.
type ExponentialBackoff struct {
	Base time.Duration
	Max  time.Duration
}

func (b ExponentialBackoff) Name() string { return "exponential" }

func (b ExponentialBackoff) Delay(attempt int) time.Duration {
	if attempt < 1 {
		return 0
	}

	// Shifting by a large attempt count would overflow, so compute in float64
	// and clamp before converting back.
	scaled := float64(b.Base) * math.Pow(2, float64(attempt-1))
	if b.Max > 0 && scaled > float64(b.Max) {
		return b.Max
	}
	if scaled > math.MaxInt64 {
		return time.Duration(math.MaxInt64)
	}

	return time.Duration(scaled)
}

// JitteredBackoff wraps another strategy and spreads the delay out, so that many
// clients retrying at once do not all wake up at the same instant.
//
// Note this is itself a Strategy that *composes* a Strategy — the interface is
// what makes that possible.
type JitteredBackoff struct {
	Inner Backoff
	// Rand returns a value in [0,1). Injected so tests stay deterministic.
	Rand func() float64
}

func (b JitteredBackoff) Name() string { return "jittered-" + b.Inner.Name() }

func (b JitteredBackoff) Delay(attempt int) time.Duration {
	base := b.Inner.Delay(attempt)
	if base <= 0 {
		return 0
	}

	r := 0.5
	if b.Rand != nil {
		r = b.Rand()
	}

	// Full jitter: pick uniformly from [0, base].
	return time.Duration(float64(base) * r)
}

// ErrExhausted is returned when every attempt failed.
var ErrExhausted = errors.New("strategy: all retry attempts exhausted")

// Retrier is the Context in Strategy terms. It owns the retry loop but delegates
// the timing policy to whichever Backoff it was given.
type Retrier struct {
	Backoff     Backoff
	MaxAttempts int

	// Sleep is injected so tests do not actually wait. Defaults to time.Sleep.
	Sleep func(time.Duration)

	delays []time.Duration
}

// Do runs op until it succeeds or MaxAttempts is reached.
func (r *Retrier) Do(ctx context.Context, op func() error) error {
	if r.MaxAttempts < 1 {
		r.MaxAttempts = 1
	}
	r.delays = nil

	var lastErr error
	for attempt := 1; attempt <= r.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = op()
		if lastErr == nil {
			return nil
		}

		// No point sleeping after the final attempt.
		if attempt == r.MaxAttempts {
			break
		}

		d := r.Backoff.Delay(attempt)
		r.delays = append(r.delays, d)

		sleep := r.Sleep
		if sleep == nil {
			sleep = time.Sleep
		}
		sleep(d)
	}

	return errors.Join(ErrExhausted, lastErr)
}

// Delays reports the waits used by the most recent Do call.
func (r *Retrier) Delays() []time.Duration { return r.delays }
