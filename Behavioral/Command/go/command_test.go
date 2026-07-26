package command

import (
	"strings"
	"testing"
)

func TestReleaseAppliesEveryStep(t *testing.T) {
	flags := NewFeatureFlags()
	dep := NewDeployment()
	dep.SetReplicas("api", 2)

	r := (&Release{}).
		Add(&Scale{Deployment: dep, Service: "api", To: 6}).
		Add(&SetFlag{Flags: flags, Flag: "new-checkout", Value: true})

	if err := r.Run(); err != nil {
		t.Fatalf("run: %v", err)
	}

	if got := dep.Replicas("api"); got != 6 {
		t.Errorf("replicas = %d, want 6", got)
	}
	if !flags.Get("new-checkout") {
		t.Error("flag was not enabled")
	}
}

// The whole point of pairing Execute with Undo: a mid-release failure leaves the
// system exactly as it was.
func TestFailedReleaseRollsBackAppliedSteps(t *testing.T) {
	flags := NewFeatureFlags()
	dep := NewDeployment()
	dep.SetReplicas("api", 2)
	flags.Set("new-checkout", false)

	r := (&Release{}).
		Add(&Scale{Deployment: dep, Service: "api", To: 6}).
		Add(&SetFlag{Flags: flags, Flag: "new-checkout", Value: true}).
		Add(&FailingCommand{Reason: "migration timed out"})

	err := r.Run()

	if err == nil {
		t.Fatal("expected the release to fail")
	}
	if !strings.Contains(err.Error(), "migration timed out") {
		t.Errorf("error = %v, want the original cause", err)
	}

	if got := dep.Replicas("api"); got != 2 {
		t.Errorf("replicas = %d, want the original 2", got)
	}
	if flags.Get("new-checkout") {
		t.Error("flag was left enabled after rollback")
	}
}

func TestRollbackRunsInReverseOrder(t *testing.T) {
	dep := NewDeployment()

	r := (&Release{}).
		Add(&Scale{Deployment: dep, Service: "api", To: 1}).
		Add(&Scale{Deployment: dep, Service: "worker", To: 2}).
		Add(&FailingCommand{Reason: "boom"})

	_ = r.Run()

	var reverted []string
	for _, line := range r.Log() {
		if strings.HasPrefix(line, "reverted: ") {
			reverted = append(reverted, strings.TrimPrefix(line, "reverted: "))
		}
	}

	want := []string{"scale worker to 2", "scale api to 1"}
	if len(reverted) != len(want) {
		t.Fatalf("reverted %v, want %v", reverted, want)
	}
	for i := range want {
		if reverted[i] != want[i] {
			t.Errorf("reverted[%d] = %q, want %q", i, reverted[i], want[i])
		}
	}
}

func TestUndoRestoresThePreviousValueNotAFixedDefault(t *testing.T) {
	flags := NewFeatureFlags()
	flags.Set("beta", true) // already on before the release

	cmd := &SetFlag{Flags: flags, Flag: "beta", Value: false}

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if flags.Get("beta") {
		t.Fatal("flag should be off after execute")
	}

	if err := cmd.Undo(); err != nil {
		t.Fatalf("undo: %v", err)
	}
	if !flags.Get("beta") {
		t.Error("undo did not restore the pre-existing value")
	}
}

func TestUndoOnAnUnexecutedCommandIsANoop(t *testing.T) {
	dep := NewDeployment()
	dep.SetReplicas("api", 3)

	cmd := &Scale{Deployment: dep, Service: "api", To: 9}

	if err := cmd.Undo(); err != nil {
		t.Fatalf("undo: %v", err)
	}
	if got := dep.Replicas("api"); got != 3 {
		t.Errorf("replicas = %d, want an untouched 3", got)
	}
}

func TestCommandValidatesItsInput(t *testing.T) {
	dep := NewDeployment()
	dep.SetReplicas("api", 3)

	cmd := &Scale{Deployment: dep, Service: "api", To: -1}

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected a negative replica count to be rejected")
	}
	if got := dep.Replicas("api"); got != 3 {
		t.Errorf("replicas = %d, want the state left untouched on failure", got)
	}
}

func TestEmptyReleaseSucceeds(t *testing.T) {
	if err := (&Release{}).Run(); err != nil {
		t.Errorf("got %v, want nil", err)
	}
}

func TestLogRecordsAppliedSteps(t *testing.T) {
	dep := NewDeployment()

	r := (&Release{}).Add(&Scale{Deployment: dep, Service: "api", To: 4})
	if err := r.Run(); err != nil {
		t.Fatalf("run: %v", err)
	}

	log := r.Log()
	if len(log) != 1 || log[0] != "applied: scale api to 4" {
		t.Errorf("log = %v", log)
	}
}
