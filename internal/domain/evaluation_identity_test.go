package domain

import "testing"

func TestEvaluationVersionRetainsProjectIdentity(t *testing.T) {
	input := validVersionInput(t)
	version, err := NewEvaluationVersion(input)
	if err != nil {
		t.Fatalf("NewEvaluationVersion() error = %v", err)
	}
	if got, want := version.ProjectID(), input.Evaluation.ProjectID(); got != want {
		t.Fatalf("ProjectID() = %q, want %q", got, want)
	}
}
