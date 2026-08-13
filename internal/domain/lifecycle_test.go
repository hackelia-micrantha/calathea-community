package domain

import (
	"testing"
	"time"
)

func testMaintainer(t *testing.T) Actor {
	t.Helper()
	actor, err := NewMaintainerActor(ActorID("maintainer"))
	if err != nil {
		t.Fatalf("NewMaintainerActor() error = %v", err)
	}
	return actor
}

func testTime() time.Time {
	return time.Date(2026, time.August, 11, 12, 0, 0, 0, time.UTC)
}

func testProjectVersion(t *testing.T, id ProjectVersionID, projectID ProjectID) ProjectVersion {
	t.Helper()
	version, err := NewProjectVersion(id, projectID, "Test project", testTime(), nil)
	if err != nil {
		t.Fatalf("NewProjectVersion() error = %v", err)
	}
	return version
}

func TestRegistrationLifecycleDecisionCreatesCandidate(t *testing.T) {
	decision, err := NewRegistrationLifecycleDecision(
		LifecycleDecisionID("life-1"),
		ProjectID("project-1"),
		testMaintainer(t),
		OperationID("op-1"),
		testTime(),
	)
	if err != nil {
		t.Fatalf("NewRegistrationLifecycleDecision() error = %v", err)
	}
	if got, want := decision.RequestedState(), LifecycleCandidate; got != want {
		t.Fatalf("RequestedState() = %q, want %q", got, want)
	}
	if got, want := decision.Kind(), LifecycleDecisionRegistration; got != want {
		t.Fatalf("Kind() = %q, want %q", got, want)
	}
	if decision.HasPriorState() {
		t.Fatal("registration unexpectedly has a prior lifecycle state")
	}
	if decision.ProjectVersionID() != nil {
		t.Fatal("candidate registration unexpectedly bound to a project version")
	}
}

func TestApprovedBootstrapRequiresExplicitContextAndVersion(t *testing.T) {
	actor := testMaintainer(t)
	projectID := ProjectID("project-1")
	validVersion := testProjectVersion(t, ProjectVersionID("project-v1"), projectID)

	if _, err := NewApprovedBootstrapLifecycleDecision(
		LifecycleDecisionID("life-bootstrap"),
		projectID,
		ProjectVersion{},
		actor,
		OperationID("op-bootstrap"),
		testTime(),
		"existing approved work",
		"existing project predates Calathea intake",
	); err == nil {
		t.Fatal("bootstrap with empty project version succeeded")
	}

	if _, err := NewApprovedBootstrapLifecycleDecision(
		LifecycleDecisionID("life-bootstrap"),
		projectID,
		validVersion,
		actor,
		OperationID("op-bootstrap"),
		testTime(),
		"",
		"existing project predates Calathea intake",
	); err == nil {
		t.Fatal("bootstrap with empty rationale succeeded")
	}

	if _, err := NewApprovedBootstrapLifecycleDecision(
		LifecycleDecisionID("life-bootstrap"),
		projectID,
		validVersion,
		actor,
		OperationID("op-bootstrap"),
		testTime(),
		"existing approved work",
		"",
	); err == nil {
		t.Fatal("bootstrap with empty skipped-intake context succeeded")
	}
}

func TestApprovedBootstrapRejectsVersionFromAnotherProject(t *testing.T) {
	projectID := ProjectID("project-1")
	otherVersion := testProjectVersion(t, ProjectVersionID("project-v2"), ProjectID("project-2"))

	_, err := NewApprovedBootstrapLifecycleDecision(
		LifecycleDecisionID("life-bootstrap"),
		projectID,
		otherVersion,
		testMaintainer(t),
		OperationID("op-bootstrap"),
		testTime(),
		"existing approved portfolio work",
		"candidate/proposed intake occurred before Calathea",
	)
	if err == nil {
		t.Fatal("bootstrap accepted a project version owned by another project")
	}
}

func TestApprovedBootstrapRecordsApprovalWithoutSilentPromotion(t *testing.T) {
	projectID := ProjectID("project-1")
	projectVersion := testProjectVersion(t, ProjectVersionID("project-v1"), projectID)
	decision, err := NewApprovedBootstrapLifecycleDecision(
		LifecycleDecisionID("life-bootstrap"),
		projectID,
		projectVersion,
		testMaintainer(t),
		OperationID("op-bootstrap"),
		testTime(),
		"existing approved portfolio work",
		"candidate/proposed intake occurred before Calathea",
	)
	if err != nil {
		t.Fatalf("NewApprovedBootstrapLifecycleDecision() error = %v", err)
	}
	if got, want := decision.RequestedState(), LifecycleApproved; got != want {
		t.Fatalf("RequestedState() = %q, want %q", got, want)
	}
	if got, want := decision.Kind(), LifecycleDecisionBootstrap; got != want {
		t.Fatalf("Kind() = %q, want %q", got, want)
	}
	if decision.SkippedIntakeContext() == "" {
		t.Fatal("bootstrap decision lost skipped intake context")
	}
	versionID := decision.ProjectVersionID()
	if versionID == nil || *versionID != ProjectVersionID("project-v1") {
		t.Fatalf("ProjectVersionID() = %v, want project-v1", versionID)
	}
	*versionID = ProjectVersionID("changed")
	if got := decision.ProjectVersionID(); got == nil || *got != ProjectVersionID("project-v1") {
		t.Fatal("mutating returned project version id changed immutable lifecycle decision")
	}
}

func TestLifecycleDecisionRequiresMaintainerAuthority(t *testing.T) {
	_, err := NewRegistrationLifecycleDecision(
		LifecycleDecisionID("life-1"),
		ProjectID("project-1"),
		Actor{},
		OperationID("op-1"),
		testTime(),
	)
	if err == nil {
		t.Fatal("registration without maintainer authority succeeded")
	}
}

func TestDirectLifecycleTransitionMatrix(t *testing.T) {
	cases := []struct {
		name string
		from LifecycleState
		to   LifecycleState
		want bool
	}{
		{name: "candidate to proposed", from: LifecycleCandidate, to: LifecycleProposed, want: true},
		{name: "approved to active", from: LifecycleApproved, to: LifecycleActive, want: true},
		{name: "active to archived forbidden", from: LifecycleActive, to: LifecycleArchived, want: false},
		{name: "paused to archived forbidden", from: LifecyclePaused, to: LifecycleArchived, want: false},
		{name: "completed reopen", from: LifecycleCompleted, to: LifecycleActive, want: true},
		{name: "archived restore requires context", from: LifecycleArchived, to: LifecycleApproved, want: false},
		{name: "same state is not transition", from: LifecycleApproved, to: LifecycleApproved, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsDirectLifecycleTransitionAllowed(tc.from, tc.to); got != tc.want {
				t.Fatalf("IsDirectLifecycleTransitionAllowed(%q, %q) = %v, want %v", tc.from, tc.to, got, tc.want)
			}
		})
	}
}

func TestKillPlacementIsNotLifecycleState(t *testing.T) {
	if LifecycleState(PlacementKill).Valid() {
		t.Fatal("kill placement unexpectedly validates as lifecycle state")
	}
}
