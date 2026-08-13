package domain

import "testing"

func TestPlacementRecommendationRejectsUnknownPlacement(t *testing.T) {
	_, err := NewPlacementRecommendation(ProjectID("project-1"), Placement("urgent"))
	if err == nil {
		t.Fatal("unknown placement was accepted")
	}
}

func TestAcceptedWithOverridesRequiresOverride(t *testing.T) {
	_, err := NewOrientationDisposition(
		OrientationDispositionID("disp-1"),
		OrientationRunID("run-1"),
		DispositionAcceptedWithOverrides,
		testMaintainer(t),
		OperationID("op-disposition"),
		"replace one placement",
		nil,
		nil,
		testTime(),
	)
	if err == nil {
		t.Fatal("accepted_with_overrides without overrides succeeded")
	}
}

func TestRejectedDispositionCannotCarryOverrides(t *testing.T) {
	override, err := NewPlacementOverride(ProjectID("project-1"), PlacementLater, "defer this project", nil)
	if err != nil {
		t.Fatalf("NewPlacementOverride() error = %v", err)
	}

	_, err = NewOrientationDisposition(
		OrientationDispositionID("disp-1"),
		OrientationRunID("run-1"),
		DispositionRejected,
		testMaintainer(t),
		OperationID("op-disposition"),
		"recommendation is not suitable",
		[]PlacementOverride{override},
		nil,
		testTime(),
	)
	if err == nil {
		t.Fatal("rejected disposition accepted placement overrides")
	}
}

func TestOrientationDispositionCopiesOverrides(t *testing.T) {
	exceptionID := PolicyExceptionID("exception-1")
	override, err := NewPlacementOverride(ProjectID("project-1"), PlacementLater, "capacity tradeoff", &exceptionID)
	if err != nil {
		t.Fatalf("NewPlacementOverride() error = %v", err)
	}

	disposition, err := NewOrientationDisposition(
		OrientationDispositionID("disp-1"),
		OrientationRunID("run-1"),
		DispositionAcceptedWithOverrides,
		testMaintainer(t),
		OperationID("op-disposition"),
		"accept with explicit override",
		[]PlacementOverride{override},
		nil,
		testTime(),
	)
	if err != nil {
		t.Fatalf("NewOrientationDisposition() error = %v", err)
	}

	got := disposition.Overrides()
	got[0] = PlacementOverride{}
	if disposition.Overrides()[0].ProjectID() != ProjectID("project-1") {
		t.Fatal("mutating returned override slice changed immutable disposition")
	}
}

func TestOrientationDispositionRejectsSelfSupersession(t *testing.T) {
	id := OrientationDispositionID("disp-1")
	_, err := NewOrientationDisposition(
		id,
		OrientationRunID("run-1"),
		DispositionAccepted,
		testMaintainer(t),
		OperationID("op-disposition"),
		"",
		nil,
		&id,
		testTime(),
	)
	if err == nil {
		t.Fatal("orientation disposition superseded itself")
	}
}

func TestRejectedDispositionCannotSupersedeAcceptedDisposition(t *testing.T) {
	prior := OrientationDispositionID("disp-prior")
	_, err := NewOrientationDisposition(
		OrientationDispositionID("disp-rejected"),
		OrientationRunID("run-2"),
		DispositionRejected,
		testMaintainer(t),
		OperationID("op-rejected"),
		"reject this proposal",
		nil,
		&prior,
		testTime(),
	)
	if err == nil {
		t.Fatal("rejected disposition superseded an accepted disposition")
	}
}

func TestOrientationDispositionCopiesSupersessionReference(t *testing.T) {
	prior := OrientationDispositionID("disp-prior")
	disposition, err := NewOrientationDisposition(
		OrientationDispositionID("disp-2"),
		OrientationRunID("run-2"),
		DispositionAccepted,
		testMaintainer(t),
		OperationID("op-disposition-2"),
		"",
		nil,
		&prior,
		testTime(),
	)
	if err != nil {
		t.Fatalf("NewOrientationDisposition() error = %v", err)
	}
	got := disposition.Supersedes()
	if got == nil || *got != prior {
		t.Fatalf("Supersedes() = %v, want %q", got, prior)
	}
	*got = OrientationDispositionID("changed")
	if current := disposition.Supersedes(); current == nil || *current != prior {
		t.Fatal("mutating returned supersession reference changed immutable disposition")
	}
}
