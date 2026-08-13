package domain

import "testing"

func TestProjectVersionRejectsSelfSupersession(t *testing.T) {
	id := ProjectVersionID("project-v1")
	_, err := NewProjectVersion(id, ProjectID("project-1"), "Calathea", testTime(), &id)
	if err == nil {
		t.Fatal("project version superseded itself")
	}
}

func TestPolicySelectionRequiresMaintainer(t *testing.T) {
	_, err := NewPolicySelectionRecord(
		PolicySelectionID("selection-1"),
		PortfolioID("portfolio-1"),
		PolicySetVersionID("policy-v1"),
		Actor{},
		OperationID("op-policy"),
		testTime(),
		"use baseline policy",
		nil,
	)
	if err == nil {
		t.Fatal("policy selection without maintainer authority succeeded")
	}
}

func TestOperationTraceCopiesEvidenceReferences(t *testing.T) {
	entry, err := NewTraceEntry("eligible", "project is lifecycle eligible", []EvidenceReferenceID{"evidence-1"})
	if err != nil {
		t.Fatalf("NewTraceEntry() error = %v", err)
	}
	trace, err := NewOperationTrace(OperationID("op-1"), []TraceEntry{entry})
	if err != nil {
		t.Fatalf("NewOperationTrace() error = %v", err)
	}

	entries := trace.Entries()
	evidence := entries[0].EvidenceIDs()
	evidence[0] = EvidenceReferenceID("changed")

	if got, want := trace.Entries()[0].EvidenceIDs()[0], EvidenceReferenceID("evidence-1"); got != want {
		t.Fatalf("trace evidence = %q, want %q", got, want)
	}
}
