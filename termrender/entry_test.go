package termrender

import "testing"

func TestEntryTimes(t *testing.T) {
	t.Parallel()

	started := NewLogScopeStarted("build")
	message := NewLogEntryMessage(nil, InfoLevel, "working")
	finished := NewLogScopeFinished(FinishTypeSucceeded, "build")

	if started.GetTime().IsZero() {
		t.Fatal("scope start time is zero")
	}
	if message.GetTime().IsZero() {
		t.Fatal("message time is zero")
	}
	if finished.GetTime().IsZero() {
		t.Fatal("scope finish time is zero")
	}
	if message.IsRaw() {
		t.Fatal("formatted message marked raw")
	}
}
