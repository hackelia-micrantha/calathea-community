package policy

import (
	"testing"
	"time"

	"github.com/hackelia-micrantha/calathea-community/internal/domain"
)

func TestHardDenialRecordsSuppressedSoftCompositionSteps(t *testing.T) {
	baseline := mustBaselinePolicySet(t)
	instances := baseline.Instances()
	firstParameters, _ := domain.NewScoreMultiplierParameters(12000)
	secondParameters, _ := domain.NewScoreMultiplierParameters(8000)
	first, _ := domain.NewScoreMultiplierPolicyInstance("suppression-multiplier-1", "orientation.test.suppression.one", firstParameters, 50, "test bounded multiplier")
	second, _ := domain.NewScoreMultiplierPolicyInstance("suppression-multiplier-2", "orientation.test.suppression.two", secondParameters, 60, "test bounded multiplier")
	instances = append(instances, first, second)
	set, _ := domain.NewPolicySet(baseline.PolicySetID())
	version, err := domain.NewPolicySetVersion(domain.PolicySetVersionInput{
		ID:                             "policy-set-suppression",
		PolicySet:                      set,
		CreatedAt:                      policyTestTime(),
		Instances:                      instances,
		MinimumScoreMultiplierBasisPts: 9000,
		MaximumScoreMultiplierBasisPts: 11000,
	})
	if err != nil {
		t.Fatalf("NewPolicySetVersion() error = %v", err)
	}
	engine := NewV0Engine()
	if err := engine.ValidatePolicySetVersion(version); err != nil {
		t.Fatalf("ValidatePolicySetVersion() error = %v", err)
	}
	projectID := domain.ProjectID("project-1")
	evaluation := mustEvaluation(t, projectID, 8000, policyTestTime().Add(-24*time.Hour))
	result, err := engine.EvaluateProject(version, ProjectContext{
		OperationID:    "op-suppression",
		ProjectID:      projectID,
		LifecycleState: domain.LifecycleCandidate,
		Evaluation:     &evaluation,
		EvaluatedAt:    policyTestTime().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("EvaluateProject() error = %v", err)
	}
	if !result.Outcome.Denied || !result.Outcome.SoftEffectsSuppressed {
		t.Fatalf("outcome = %#v, want denied with suppressed soft effects", result.Outcome)
	}
	if got, want := len(result.Outcome.MultiplierSteps), 2; got != want {
		t.Fatalf("multiplier steps = %d, want %d", got, want)
	}
	if got, want := len(result.Outcome.SuppressedDecisionIDs), 2; got != want {
		t.Fatalf("suppressed decision ids = %d, want %d", got, want)
	}
	for _, step := range result.Outcome.MultiplierSteps {
		if !step.Suppressed || step.ReasonCode != "suppressed_by_hard_policy" {
			t.Fatalf("multiplier step lacks suppression trace: %#v", step)
		}
		if step.Before.String() != "1" {
			t.Fatalf("suppressed step effective before = %q, want 1", step.Before.String())
		}
	}
	if got := result.Outcome.ScoreMultiplier.String(); got != "1" {
		t.Fatalf("effective multiplier after hard denial = %q, want 1", got)
	}
}

func TestExceptionValidationRequiresEvidence(t *testing.T) {
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
		ID:                 "exception-without-evidence",
		PolicySetVersionID: version.ID(),
		PolicyID:           confidence.PolicyID(),
		PolicyInstanceID:   confidence.ID(),
		EvaluatorVersion:   confidence.EvaluatorVersion(),
		ProjectID:          "project-1",
		Phase:              confidence.Phase(),
		PermittedDeviation: "accept after manual review",
		Actor:              mustMaintainer(t),
		Rationale:          "manual review performed",
		CreatedAt:          base,
		EffectiveAt:        base,
		ExpiresAt:          base.Add(24 * time.Hour),
		MaximumUses:        1,
	})
	if err != nil {
		t.Fatalf("construct incomplete exception for validation test: %v", err)
	}
	if err := engine.ValidateExceptionUse(version, exception, nil, nil, "project-1", confidence.Phase(), base.Add(time.Hour)); err == nil {
		t.Fatal("exception without supporting evidence passed validity check")
	}
}

func TestNonExceptionableHardPolicyCannotBeBypassed(t *testing.T) {
	engine := NewV0Engine()
	version := mustBaselinePolicySet(t)
	var nowCapacity domain.PolicyInstance
	for _, instance := range version.Instances() {
		if instance.PolicyID() == PolicyCapacityNow {
			nowCapacity = instance
			break
		}
	}
	if nowCapacity.ID() == "" {
		t.Fatal("now capacity policy not found")
	}
	base := policyTestTime()
	exception, err := domain.NewPolicyException(domain.PolicyExceptionInput{
		ID:                 "capacity-exception",
		PolicySetVersionID: version.ID(),
		PolicyID:           nowCapacity.PolicyID(),
		PolicyInstanceID:   nowCapacity.ID(),
		EvaluatorVersion:   nowCapacity.EvaluatorVersion(),
		ProjectID:          "project-1",
		Phase:              nowCapacity.Phase(),
		PermittedDeviation: "exceed now capacity",
		Actor:              mustMaintainer(t),
		Rationale:          "test attempted hard-policy bypass",
		EvidenceIDs:        []domain.EvidenceReferenceID{"evidence-1"},
		CreatedAt:          base,
		EffectiveAt:        base,
		ExpiresAt:          base.Add(24 * time.Hour),
		MaximumUses:        1,
	})
	if err != nil {
		t.Fatalf("NewPolicyException() error = %v", err)
	}
	if err := engine.ValidateExceptionUse(version, exception, nil, nil, "project-1", nowCapacity.Phase(), base.Add(time.Hour)); err == nil {
		t.Fatal("non-exceptionable hard capacity policy accepted an exception")
	}
}

func TestExceptionableHardPolicyCanUseValidScopedException(t *testing.T) {
	baseline := mustBaselinePolicySet(t)
	instances := make([]domain.PolicyInstance, 0, len(baseline.Instances()))
	for _, instance := range baseline.Instances() {
		if instance.PolicyID() != PolicyEvaluationRequired {
			instances = append(instances, instance)
		}
	}
	required, err := domain.NewRequiredEvaluationPolicyInstance(
		"exceptionable-required-evaluation",
		PolicyEvaluationRequired,
		20,
		domain.PolicyMissingExcludeSubject,
		domain.PolicyExceptionableWithReview,
		"accepted evaluation normally required, with explicitly reviewed exception support",
	)
	if err != nil {
		t.Fatalf("NewRequiredEvaluationPolicyInstance() error = %v", err)
	}
	instances = append(instances, required)
	set, _ := domain.NewPolicySet(baseline.PolicySetID())
	version, err := domain.NewPolicySetVersion(domain.PolicySetVersionInput{
		ID:                             "policy-set-exceptionable-hard",
		PolicySet:                      set,
		CreatedAt:                      policyTestTime(),
		Instances:                      instances,
		MinimumScoreMultiplierBasisPts: 10000,
		MaximumScoreMultiplierBasisPts: 10000,
	})
	if err != nil {
		t.Fatalf("NewPolicySetVersion() error = %v", err)
	}
	engine := NewV0Engine()
	if err := engine.ValidatePolicySetVersion(version); err != nil {
		t.Fatalf("ValidatePolicySetVersion() error = %v", err)
	}
	base := policyTestTime()
	exception, err := domain.NewPolicyException(domain.PolicyExceptionInput{
		ID:                 "required-evaluation-exception",
		PolicySetVersionID: version.ID(),
		PolicyID:           required.PolicyID(),
		PolicyInstanceID:   required.ID(),
		EvaluatorVersion:   required.EvaluatorVersion(),
		ProjectID:          "project-1",
		Phase:              required.Phase(),
		PermittedDeviation: "allow reviewed project without accepted evaluation for one orientation",
		Actor:              mustMaintainer(t),
		Rationale:          "evidence reviewed manually",
		EvidenceIDs:        []domain.EvidenceReferenceID{"evidence-1"},
		CreatedAt:          base,
		EffectiveAt:        base,
		ExpiresAt:          base.Add(24 * time.Hour),
		MaximumUses:        1,
	})
	if err != nil {
		t.Fatalf("NewPolicyException() error = %v", err)
	}
	if err := engine.ValidateExceptionUse(version, exception, nil, nil, "project-1", required.Phase(), base.Add(time.Hour)); err != nil {
		t.Fatalf("valid exceptionable hard-policy exception rejected: %v", err)
	}
}
