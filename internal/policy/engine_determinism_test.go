package policy

import (
	"strings"
	"testing"
	"time"

	"github.com/hackelia-micrantha/calathea-community/internal/domain"
)

func TestPolicySetValidationReportsMissingBaselinePoliciesDeterministically(t *testing.T) {
	set, err := domain.NewPolicySet("policy-set")
	if err != nil {
		t.Fatalf("NewPolicySet() error = %v", err)
	}
	onlyRequired, err := domain.NewRequiredEvaluationPolicyInstance(
		"only-required-evaluation",
		PolicyEvaluationRequired,
		20,
		domain.PolicyMissingExcludeSubject,
		domain.PolicyNotExceptionable,
		"test incomplete baseline",
	)
	if err != nil {
		t.Fatalf("NewRequiredEvaluationPolicyInstance() error = %v", err)
	}
	version, err := domain.NewPolicySetVersion(domain.PolicySetVersionInput{
		ID:                             "incomplete-policy-set",
		PolicySet:                      set,
		CreatedAt:                      policyTestTime(),
		Instances:                      []domain.PolicyInstance{onlyRequired},
		MinimumScoreMultiplierBasisPts: 10000,
		MaximumScoreMultiplierBasisPts: 10000,
	})
	if err != nil {
		t.Fatalf("NewPolicySetVersion() error = %v", err)
	}

	engine := NewV0Engine()
	var first string
	for i := 0; i < 20; i++ {
		err := engine.ValidatePolicySetVersion(version)
		if err == nil {
			t.Fatal("incomplete baseline policy set unexpectedly validated")
		}
		if i == 0 {
			first = err.Error()
			continue
		}
		if err.Error() != first {
			t.Fatalf("validation error changed across runs: first=%q later=%q", first, err.Error())
		}
	}
	if !strings.Contains(first, string(PolicyLifecycleEligibility)) {
		t.Fatalf("first deterministic missing-policy error = %q, want lifecycle policy", first)
	}
}

func TestFreshnessDecisionRetainsPlanningHorizon(t *testing.T) {
	engine := NewV0Engine()
	version := mustBaselinePolicySet(t)
	projectID := domain.ProjectID("project-1")
	evaluation := mustEvaluation(t, projectID, 8000, policyTestTime().Add(-24*time.Hour))
	result, err := engine.EvaluateProject(version, ProjectContext{
		OperationID:    "op-freshness-trace",
		ProjectID:      projectID,
		LifecycleState: domain.LifecycleApproved,
		Evaluation:     &evaluation,
		EvaluatedAt:    policyTestTime().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("EvaluateProject() error = %v", err)
	}
	for _, decision := range result.Decisions {
		if decision.PolicyID() != PolicyEvaluationFreshness {
			continue
		}
		for _, reference := range decision.InputReferences() {
			if reference == "planning_horizon:next 90 days" {
				return
			}
		}
		t.Fatalf("freshness decision input references = %#v, missing planning horizon", decision.InputReferences())
	}
	t.Fatal("freshness policy decision not found")
}

func TestScoreMultiplierMissingInputsMatchDeclaredContract(t *testing.T) {
	baseline := mustBaselinePolicySet(t)
	parameters, err := domain.NewScoreMultiplierParameters(10500)
	if err != nil {
		t.Fatalf("NewScoreMultiplierParameters() error = %v", err)
	}
	multiplier, err := domain.NewScoreMultiplierPolicyInstance(
		"score-multiplier",
		"orientation.test.score-multiplier",
		parameters,
		50,
		"test explicit score multiplier",
	)
	if err != nil {
		t.Fatalf("NewScoreMultiplierPolicyInstance() error = %v", err)
	}
	set, _ := domain.NewPolicySet(baseline.PolicySetID())
	version, err := domain.NewPolicySetVersion(domain.PolicySetVersionInput{
		ID:                             "policy-set-score-multiplier",
		PolicySet:                      set,
		CreatedAt:                      policyTestTime(),
		Instances:                      append(baseline.Instances(), multiplier),
		MinimumScoreMultiplierBasisPts: 10000,
		MaximumScoreMultiplierBasisPts: 11000,
	})
	if err != nil {
		t.Fatalf("NewPolicySetVersion() error = %v", err)
	}
	engine := NewV0Engine()
	if err := engine.ValidatePolicySetVersion(version); err != nil {
		t.Fatalf("ValidatePolicySetVersion() error = %v", err)
	}

	_, err = engine.EvaluateProject(version, ProjectContext{
		OperationID:    "op-score-missing",
		ProjectID:      "project-1",
		LifecycleState: domain.LifecycleApproved,
		EvaluatedAt:    policyTestTime().Add(time.Hour),
	})
	if err == nil {
		t.Fatal("score multiplier with missing accepted evaluation unexpectedly succeeded")
	}
	if !strings.Contains(err.Error(), "requires operation failure") {
		t.Fatalf("missing score input failed before producing a valid policy decision: %v", err)
	}
}

func TestExceptionRevocationDiagnosticUsesEarliestApplicableRevocation(t *testing.T) {
	engine := NewV0Engine()
	version := mustBaselinePolicySet(t)
	var confidence domain.PolicyInstance
	for _, instance := range version.Instances() {
		if instance.PolicyID() == PolicyEvaluationConfidence {
			confidence = instance
			break
		}
	}
	if confidence.ID() == "" {
		t.Fatal("confidence policy not found")
	}
	base := policyTestTime()
	exception, err := domain.NewPolicyException(domain.PolicyExceptionInput{
		ID:                 "exception-revocation-order",
		PolicySetVersionID: version.ID(),
		PolicyID:           confidence.PolicyID(),
		PolicyInstanceID:   confidence.ID(),
		EvaluatorVersion:   confidence.EvaluatorVersion(),
		ProjectID:          "project-1",
		Phase:              confidence.Phase(),
		PermittedDeviation: "accept after explicit review",
		Actor:              mustMaintainer(t),
		Rationale:          "reviewed evidence manually",
		EvidenceIDs:        []domain.EvidenceReferenceID{"evidence-1"},
		CreatedAt:          base,
		EffectiveAt:        base,
		ExpiresAt:          base.Add(24 * time.Hour),
		MaximumUses:        2,
	})
	if err != nil {
		t.Fatalf("NewPolicyException() error = %v", err)
	}
	later, err := domain.NewPolicyExceptionRevocation("revocation-later", exception.ID(), mustMaintainer(t), "later duplicate revocation", base.Add(20*time.Minute))
	if err != nil {
		t.Fatalf("NewPolicyExceptionRevocation(later) error = %v", err)
	}
	earlier, err := domain.NewPolicyExceptionRevocation("revocation-earlier", exception.ID(), mustMaintainer(t), "earlier revocation", base.Add(10*time.Minute))
	if err != nil {
		t.Fatalf("NewPolicyExceptionRevocation(earlier) error = %v", err)
	}

	err = engine.ValidateExceptionUse(
		version,
		exception,
		nil,
		[]domain.PolicyExceptionRevocation{later, earlier},
		"project-1",
		confidence.Phase(),
		base.Add(time.Hour),
	)
	if err == nil {
		t.Fatal("revoked policy exception unexpectedly validated")
	}
	want := earlier.RevokedAt().UTC().Format(time.RFC3339)
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("revocation error = %q, want earliest revocation %q", err.Error(), want)
	}
}
