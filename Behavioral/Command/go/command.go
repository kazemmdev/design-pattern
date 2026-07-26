// Package command demonstrates the Command behavioral pattern.
//
// A release is a sequence of operations: scale a service, flip a feature flag,
// run a migration. If step four fails, everything already applied has to be
// undone in reverse. Turning each operation into an object that knows both how
// to do and how to undo itself makes that rollback mechanical rather than a pile
// of bespoke cleanup code.
package command

import (
	"errors"
	"fmt"
)

// Command is the interface every operation implements.
type Command interface {
	Name() string
	Execute() error
	Undo() error
}

// --- Receivers --------------------------------------------------------------

// FeatureFlags is a receiver: the thing commands actually act on.
type FeatureFlags struct {
	flags map[string]bool
}

func NewFeatureFlags() *FeatureFlags {
	return &FeatureFlags{flags: make(map[string]bool)}
}

func (f *FeatureFlags) Get(name string) bool { return f.flags[name] }

func (f *FeatureFlags) Set(name string, on bool) { f.flags[name] = on }

// Deployment tracks replica counts per service.
type Deployment struct {
	replicas map[string]int
}

func NewDeployment() *Deployment {
	return &Deployment{replicas: make(map[string]int)}
}

func (d *Deployment) Replicas(service string) int { return d.replicas[service] }

func (d *Deployment) SetReplicas(service string, n int) { d.replicas[service] = n }

// --- Concrete commands ------------------------------------------------------

// SetFlag turns a feature flag on or off, remembering the previous value so the
// change can be reversed exactly.
type SetFlag struct {
	Flags *FeatureFlags
	Flag  string
	Value bool

	prev     bool
	executed bool
}

func (c *SetFlag) Name() string { return fmt.Sprintf("set-flag %s=%v", c.Flag, c.Value) }

func (c *SetFlag) Execute() error {
	c.prev = c.Flags.Get(c.Flag)
	c.Flags.Set(c.Flag, c.Value)
	c.executed = true

	return nil
}

func (c *SetFlag) Undo() error {
	if !c.executed {
		return nil
	}
	c.Flags.Set(c.Flag, c.prev)
	c.executed = false

	return nil
}

// Scale changes a service's replica count.
type Scale struct {
	Deployment *Deployment
	Service    string
	To         int

	prev     int
	executed bool
}

func (c *Scale) Name() string { return fmt.Sprintf("scale %s to %d", c.Service, c.To) }

func (c *Scale) Execute() error {
	if c.To < 0 {
		return fmt.Errorf("scale %s: replica count %d is negative", c.Service, c.To)
	}
	c.prev = c.Deployment.Replicas(c.Service)
	c.Deployment.SetReplicas(c.Service, c.To)
	c.executed = true

	return nil
}

func (c *Scale) Undo() error {
	if !c.executed {
		return nil
	}
	c.Deployment.SetReplicas(c.Service, c.prev)
	c.executed = false

	return nil
}

// FailingCommand is a deliberately broken step, used to exercise rollback.
type FailingCommand struct {
	Reason string
}

func (c *FailingCommand) Name() string { return "failing-step" }

func (c *FailingCommand) Execute() error { return errors.New(c.Reason) }

func (c *FailingCommand) Undo() error { return nil }

// --- Invoker ----------------------------------------------------------------

// Release is the Invoker. It runs commands in order and, if any step fails,
// rolls back the steps that already succeeded — in reverse order.
type Release struct {
	commands []Command
	executed []Command
	log      []string
}

func (r *Release) Add(c Command) *Release {
	r.commands = append(r.commands, c)

	return r
}

// Run applies every command. On failure it rolls back and returns the original
// error, so the caller sees the real cause rather than a cleanup error.
func (r *Release) Run() error {
	for _, c := range r.commands {
		if err := c.Execute(); err != nil {
			r.log = append(r.log, "failed: "+c.Name())

			if rollbackErr := r.Rollback(); rollbackErr != nil {
				return errors.Join(fmt.Errorf("%s: %w", c.Name(), err), rollbackErr)
			}

			return fmt.Errorf("%s: %w", c.Name(), err)
		}

		r.executed = append(r.executed, c)
		r.log = append(r.log, "applied: "+c.Name())
	}

	return nil
}

// Rollback undoes the applied commands, most recent first.
func (r *Release) Rollback() error {
	var errs []error

	for i := len(r.executed) - 1; i >= 0; i-- {
		c := r.executed[i]
		if err := c.Undo(); err != nil {
			errs = append(errs, fmt.Errorf("undo %s: %w", c.Name(), err))
			continue
		}
		r.log = append(r.log, "reverted: "+c.Name())
	}

	r.executed = nil

	return errors.Join(errs...)
}

// Log reports what the release did, in order. Handy for an audit trail.
func (r *Release) Log() []string { return append([]string(nil), r.log...) }
