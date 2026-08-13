package domain

import (
	"testing"
	"time"
)

func mustPolicyInstance(t *testing.T, id PolicyInstanceID, priority int) PolicyInstance {
	t.Helper()
	instance, err := NewRequiredEvaluationPolicyInstance(
		id,
		PolicyID("orientation.test."+string(id)),
		priority,
		PolicyMissingExcludeSubject,
		PolicyNotExceptionable,
		"test policy",
	)
	if err != nil {
		t.Fatalf("NewRequiredEvaluationPolicyInstance() error = %v", err)
	}
	return instance
}

func TestPolicySetVersionOrdersInstancesAndDefensivelyCopies(t *testing.T) {
	set, err := NewPolicySet("policy-set")
	if err != nil {
		t.Fatalf("NewPolicySet() error = %v", err)
	}
	late := mustPolicyInstance(t, "late", 20)
	early := mustPolicyInstance(t, "early", 10)
	version, err := NewPolicySetVersion(PolicySetVersionInput{
		ID:                             "policy-set-v1",
		PolicySet:                      set,
		CreatedAt:                      testTime(),
		Instances:                      []PolicyInstance{late, early},
		MinimumScoreMultiplierBasisPts: 10000,
		MaximumScoreMultiplierBasisPts: 10000,
	})
	if err != nil {
		t.Fatalf("NewPolicySetVersion() error = %v", err)
	}

	instances := version.Instances()
	if got, want := instances[0].ID(), PolicyInstanceID("early"); got != want {
		t.Fatalf("first policy instance = %q, want %q", got, want)
	}
	instances[0] = PolicyInstance{}
	if got := version.Instances()[0].ID(); got != "early" {
		t.Fatalf("mutating returned instances changed policy set: first id = %q", got)
	}
}

func TestPolicySetVersionRejectsDuplicateInstancesAndInvalidBounds(t *testing.T) {
	set, _ := NewPolicySet("policy-set")
	instance := mustPolicyInstance(t, "instance", 10)
	base := PolicySetVersionInput{
		ID:                             "policy-set-v1",
		PolicySet:                      set,
		CreatedAt:                      testTime(),
		Instances:                      []PolicyInstance{instance},
		MinimumScoreMultiplierBasisPts: 10000,
		MaximumScoreMultiplierBasisPts: 10000,
	}

	duplicate := base
	duplicate.Instances = []PolicyInstance{instance, instance}
	if _, err := NewPolicySetVersion(duplicate); err == nil {
		t.Fatal("policy set version with duplicate instance id succeeded")
	}

	invalidBounds := base
	invalidBounds.MinimumScoreMultiplierBasisPts = 11000
	invalidBounds.MaximumScoreMultiplierBasisPts = 10000
	if _, err := NewPolicySetVersion(invalidBounds); err == nil {
		t.Fatal("policy set version with inverted multiplier bounds succeeded")
	}
}

func TestPolicyDecisionCopiesTraceInputsAndEffects(t *testing.T) {
	effect, err := NewDiagnosticPolicyEffect("confidence_band", "weak_evidence")
	if err != nil {
		t.Fatalf("NewDiagnosticPolicyEffect() error = %v", err)
	}
	inputReferences := []string{"evaluation_version:evaluation-v1"}
	evidenceIDs := []EvidenceReferenceID{"evidence-1"}
	effects := []PolicyEffect{effect}
	decision, err := NewPolicyDecision(PolicyDecisionInput{
		ID:                   "decision-1",
		PolicySetVersionID:   "policy-set-v1",
		PolicyID:             "orientation.evaluation.confidence",
		PolicyInstanceID:     "confidence",
		EvaluatorType:        PolicyEvaluatorConfidenceGate,
		EvaluatorVersion:     PolicyEvaluatorSemanticVersionV1,
		ProjectID:            "project-1",
		OperationID:          "operation-1",
		Result:               PolicyDecisionRequireReview,
		EffectClass:          PolicyEffectReviewRequired,
		Phase:                PolicyPhaseCandidateAdjustment,
		Effects:              effects,
		ReasonCode:           "confidence_below_threshold",
		InputReferences:      inputReferences,
		EvidenceIDs:          evidenceIDs,
		MissingInputBehavior: PolicyMissingRequireReview,
		Priority:             10,
		CreatedAt:            testTime(),
	})
	if err != nil {
		t.Fatalf("NewPolicyDecision() error = %v", err)
	}

	inputReferences[0] = "mutated"
	evidenceIDs[0] = "mutated"
	effects[0] = PolicyEffect{}
	if got := decision.InputReferences()[0]; got != "evaluation_version:evaluation-v1" {
		t.Fatalf("decision input reference mutated: %q", got)
	}
	if got := decision.EvidenceIDs()[0]; got != "evidence-1" {
		t.Fatalf("decision evidence id mutated: %q", got)
	}
	if got := decision.Effects()[0].Code(); got != "confidence_band" {
		t.Fatalf("decision effect mutated: code = %q", got)
	}
}

func TestPolicyExceptionRequiresAuthorityScopeAndBoundedLifetime(t *testing.T) {
	actor := testMaintainer(t)
	base := PolicyExceptionInput{
		ID:                 "exception-1",
		PolicySetVersionID: "policy-set-v1",
		PolicyID:           "orientation.evaluation.confidence",
		PolicyInstanceID:   "confidence",
		EvaluatorVersion:   PolicyEvaluatorSemanticVersionV1,
		ProjectID:          "project-1",
		Phase:              PolicyPhaseCandidateAdjustment,
		PermittedDeviation: "allow acceptance after explicit evidence review",
		Actor:              actor,
		Rationale:          "maintainer reviewed evidence",
		EvidenceIDs:        []EvidenceReferenceID{"evidence-1"},
		CreatedAt:          testTime(),
		EffectiveAt:        testTime(),
		ExpiresAt:          testTime().Add(24 * time.Hour),
		MaximumUses:        1,
	}
	if _, err := NewPolicyException(base); err != nil {
		t.Fatalf("valid NewPolicyException() error = %v", err)
	}

	withoutAuthority := base
	withoutAuthority.Actor = Actor{}
	if _, err := NewPolicyException(withoutAuthority); err == nil {
		t.Fatal("policy exception without maintainer authority succeeded")
	}

	unboundedUses := base
	unboundedUses.MaximumUses = 0
	if _, err := NewPolicyException(unboundedUses); err == nil {
		t.Fatal("policy exception with zero maximum uses succeeded")
	}

	invalidLifetime := base
	invalidLifetime.ExpiresAt = invalidLifetime.EffectiveAt
	if _, err := NewPolicyException(invalidLifetime); err == nil {
		t.Fatal("policy exception with non-positive lifetime succeeded")
	}
}
