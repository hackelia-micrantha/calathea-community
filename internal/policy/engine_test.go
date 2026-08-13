package policy

import (
	"testing"
	"time"

	"github.com/hackelia-micrantha/calathea-community/internal/domain"
)

func policyTestTime() time.Time {
	return time.Date(2026, time.August, 12, 12, 0, 0, 0, time.UTC)
}

func intPointer(value int) *int { return &value }

func mustPolicySet(t *testing.T) domain.PolicySet {
	t.Helper()
	set, err := domain.NewPolicySet("policy-set")
	if err != nil {
		t.Fatalf("NewPolicySet() error = %v", err)
	}
	return set
}

func mustBaselinePolicySet(t *testing.T) domain.PolicySetVersion {
	t.Helper()
	version, err := NewBaselinePolicySetVersion(BaselinePolicySetConfig{
		ID:                  "policy-set-v1",
		PolicySet:           mustPolicySet(t),
		CreatedAt:           policyTestTime(),
		MaxNext:             intPointer(10),
		MinimumConfidence:   domain.ConfidenceVisibleUncertainty,
		FreshnessMaximumAge: 30 * 24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("NewBaselinePolicySetVersion() error = %v", err)
	}
	return version
}

func mustMaintainer(t *testing.T) domain.Actor {
	t.Helper()
	actor, err := domain.NewMaintainerActor("maintainer")
	if err != nil {
		t.Fatalf("NewMaintainerActor() error = %v", err)
	}
	return actor
}

func mustEvaluation(t *testing.T, projectID domain.ProjectID, confidenceBasisPoints int, evidenceAsOf time.Time) domain.EvaluationVersion {
	t.Helper()
	projectVersion, err := domain.NewProjectVersion("project-v1", projectID, "Project", policyTestTime().Add(-48*time.Hour), nil)
	if err != nil {
		t.Fatalf("NewProjectVersion() error = %v", err)
	}
	evaluation, err := domain.NewEvaluation("evaluation", projectID)
	if err != nil {
		t.Fatalf("NewEvaluation() error = %v", err)
	}
	axis := func(value int, rationale string) domain.AxisAssessment {
		assessment, axisErr := domain.NewAxisAssessment(value, rationale)
		if axisErr != nil {
			t.Fatalf("NewAxisAssessment() error = %v", axisErr)
		}
		return assessment
	}
	axes, err := domain.NewEvaluationAxes(
		axis(4, "impact"),
		axis(2, "effort"),
		axis(3, "risk reduction"),
		axis(2, "optionality"),
		axis(3, "urgency"),
		[]string{"follow-on"},
	)
	if err != nil {
		t.Fatalf("NewEvaluationAxes() error = %v", err)
	}
	confidence, err := domain.NewConfidence(confidenceBasisPoints, "policy test confidence")
	if err != nil {
		t.Fatalf("NewConfidence() error = %v", err)
	}
	freshness, err := domain.NewFreshnessMetadata(evidenceAsOf)
	if err != nil {
		t.Fatalf("NewFreshnessMetadata() error = %v", err)
	}
	version, err := domain.NewEvaluationVersion(domain.EvaluationVersionInput{
		ID:                     "evaluation-v1",
		Evaluation:             evaluation,
		ProjectVersion:         projectVersion,
		EvaluatedAt:            policyTestTime(),
		PlanningHorizon:        "next 90 days",
		Freshness:              freshness,
		Axes:                   axes,
		Confidence:             confidence,
		Derivation:             domain.EvaluationDerivationAuthored,
		EvidenceIDs:            []domain.EvidenceReferenceID{"evidence-1"},
		SemanticVersion:        domain.EvaluationSemanticVersionV1,
		FormulaSemanticVersion: domain.BaseScoreFormulaSemanticVersionV1,
		AcceptedBy:             mustMaintainer(t),
		AcceptedAt:             policyTestTime(),
	})
	if err != nil {
		t.Fatalf("NewEvaluationVersion() error = %v", err)
	}
	return version
}

func TestBaselinePolicySetHasRequiredDeterministicPolicies(t *testing.T) {
	version := mustBaselinePolicySet(t)
	engine := NewV0Engine()
	if err := engine.ValidatePolicySetVersion(version); err != nil {
		t.Fatalf("ValidatePolicySetVersion() error = %v", err)
	}
	if got, want := len(version.Instances()), 6; got != want {
		t.Fatalf("len(Instances()) = %d, want %d", got, want)
	}

	var nowMaximum, nextMaximum uint32
	for _, instance := range version.Instances() {
		parameters, ok := instance.CapacityLimitParameters()
		if !ok {
			continue
		}
		switch parameters.Placement() {
		case domain.PlacementNow:
			nowMaximum = parameters.Maximum()
		case domain.PlacementNext:
			nextMaximum = parameters.Maximum()
		}
	}
	if nowMaximum != 3 || nextMaximum != 10 {
		t.Fatalf("capacity defaults = now:%d next:%d, want now:3 next:10", nowMaximum, nextMaximum)
	}
}

func TestBaselinePolicySetPreservesExplicitZeroNextCapacity(t *testing.T) {
	version, err := NewBaselinePolicySetVersion(BaselinePolicySetConfig{
		ID:                  "policy-set-zero-next",
		PolicySet:           mustPolicySet(t),
		CreatedAt:           policyTestTime(),
		MaxNext:             intPointer(0),
		FreshnessMaximumAge: 30 * 24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("NewBaselinePolicySetVersion() error = %v", err)
	}
	for _, instance := range version.Instances() {
		parameters, ok := instance.CapacityLimitParameters()
		if ok && parameters.Placement() == domain.PlacementNext {
			if parameters.Maximum() != 0 {
				t.Fatalf("next capacity = %d, want explicit zero", parameters.Maximum())
			}
			return
		}
	}
	t.Fatal("next capacity policy not found")
}

func TestProjectEvaluationIsDeterministicAndExplainable(t *testing.T) {
	engine := NewV0Engine()
	version := mustBaselinePolicySet(t)
	projectID := domain.ProjectID("project-1")
	evaluation := mustEvaluation(t, projectID, 8000, policyTestTime().Add(-24*time.Hour))
	context := ProjectContext{
		OperationID:    "op-1",
		ProjectID:      projectID,
		LifecycleState: domain.LifecycleApproved,
		Evaluation:     &evaluation,
		EvaluatedAt:    policyTestTime().Add(24 * time.Hour),
	}

	first, err := engine.EvaluateProject(version, context)
	if err != nil {
		t.Fatalf("EvaluateProject() error = %v", err)
	}
	second, err := engine.EvaluateProject(version, context)
	if err != nil {
		t.Fatalf("second EvaluateProject() error = %v", err)
	}
	if first.Outcome.Denied || first.Outcome.Excluded || first.Outcome.ReviewRequired || first.Outcome.Indeterminate {
		t.Fatalf("unexpected baseline outcome: %#v", first.Outcome)
	}
	if got, want := len(first.Decisions), 4; got != want {
		t.Fatalf("len(Decisions) = %d, want %d", got, want)
	}
	for i := range first.Decisions {
		if first.Decisions[i].ID() != second.Decisions[i].ID() {
			t.Fatalf("decision %d id changed across identical evaluation: %q != %q", i, first.Decisions[i].ID(), second.Decisions[i].ID())
		}
		if first.Decisions[i].ReasonCode() == "" {
			t.Fatalf("decision %d lacks reason code", i)
		}
	}
	if len(first.Decisions[0].InputReferences()) == 0 {
		t.Fatal("first policy decision lacks deterministic input references")
	}
}

func TestMissingEvaluationFailsClosedWithoutGuessing(t *testing.T) {
	result, err := NewV0Engine().EvaluateProject(mustBaselinePolicySet(t), ProjectContext{
		OperationID:    "op-missing-evaluation",
		ProjectID:      "project-1",
		LifecycleState: domain.LifecycleApproved,
		EvaluatedAt:    policyTestTime().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("EvaluateProject() error = %v", err)
	}
	if !result.Outcome.Excluded || !result.Outcome.ReviewRequired || !result.Outcome.Indeterminate {
		t.Fatalf("missing evaluation outcome = %#v, want excluded + review + indeterminate", result.Outcome)
	}
	for _, decision := range result.Decisions {
		if decision.Result() == domain.PolicyDecisionIndeterminate && len(decision.MissingInputs()) == 0 {
			t.Fatalf("indeterminate decision %q lacks missing-input trace", decision.ID())
		}
	}
}

func TestLifecycleAndConfidencePoliciesRemainDistinct(t *testing.T) {
	engine := NewV0Engine()
	version := mustBaselinePolicySet(t)
	projectID := domain.ProjectID("project-1")
	weakEvaluation := mustEvaluation(t, projectID, 3000, policyTestTime().Add(-24*time.Hour))
	result, err := engine.EvaluateProject(version, ProjectContext{
		OperationID:    "op-candidate",
		ProjectID:      projectID,
		LifecycleState: domain.LifecycleCandidate,
		Evaluation:     &weakEvaluation,
		EvaluatedAt:    policyTestTime().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("EvaluateProject() error = %v", err)
	}
	if !result.Outcome.Denied {
		t.Fatal("candidate lifecycle state was not denied")
	}
	if !result.Outcome.ReviewRequired {
		t.Fatal("weak confidence diagnostic disappeared behind lifecycle denial")
	}
}

func TestFreshnessRuleRequiresReviewWithoutMutatingEvaluation(t *testing.T) {
	engine := NewV0Engine()
	version := mustBaselinePolicySet(t)
	projectID := domain.ProjectID("project-1")
	stale := mustEvaluation(t, projectID, 8000, policyTestTime().Add(-45*24*time.Hour))
	result, err := engine.EvaluateProject(version, ProjectContext{
		OperationID:    "op-stale",
		ProjectID:      projectID,
		LifecycleState: domain.LifecycleApproved,
		Evaluation:     &stale,
		EvaluatedAt:    policyTestTime().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("EvaluateProject() error = %v", err)
	}
	if !result.Outcome.ReviewRequired {
		t.Fatal("stale evaluation did not require review")
	}
	if got, want := stale.Freshness().EvidenceAsOf(), policyTestTime().Add(-45*24*time.Hour); !got.Equal(want) {
		t.Fatalf("freshness policy mutated evaluation: got %v want %v", got, want)
	}
}

func TestCapacityPoliciesEnforceNowAndNextBounds(t *testing.T) {
	engine := NewV0Engine()
	version := mustBaselinePolicySet(t)
	cases := []struct {
		name          string
		placement     domain.Placement
		selectedCount int
		wantDenied    bool
	}{
		{"now available", domain.PlacementNow, 2, false},
		{"now full", domain.PlacementNow, 3, true},
		{"next available", domain.PlacementNext, 9, false},
		{"next full", domain.PlacementNext, 10, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.EvaluateCapacity(version, CapacityContext{
				OperationID:   domain.OperationID("op-capacity-" + tc.name),
				ProjectID:     "project-1",
				Placement:     tc.placement,
				SelectedCount: tc.selectedCount,
				EvaluatedAt:   policyTestTime(),
			})
			if err != nil {
				t.Fatalf("EvaluateCapacity() error = %v", err)
			}
			if result.Outcome.Denied != tc.wantDenied {
				t.Fatalf("Denied = %v, want %v", result.Outcome.Denied, tc.wantDenied)
			}
		})
	}
}

func TestSoftMultiplierCompositionIsExactAndHardDenialSuppressesIt(t *testing.T) {
	baseline := mustBaselinePolicySet(t)
	instances := baseline.Instances()
	firstParameters, _ := domain.NewScoreMultiplierParameters(12000)
	secondParameters, _ := domain.NewScoreMultiplierParameters(8000)
	first, _ := domain.NewScoreMultiplierPolicyInstance("multiplier-1", "orientation.test.multiplier.one", firstParameters, 50, "test bounded multiplier")
	second, _ := domain.NewScoreMultiplierPolicyInstance("multiplier-2", "orientation.test.multiplier.two", secondParameters, 60, "test bounded multiplier")
	instances = append(instances, first, second)
	set, _ := domain.NewPolicySet(baseline.PolicySetID())
	version, err := domain.NewPolicySetVersion(domain.PolicySetVersionInput{
		ID:                             "policy-set-with-multipliers",
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
	approved, err := engine.EvaluateProject(version, ProjectContext{
		OperationID:    "op-multipliers-approved",
		ProjectID:      projectID,
		LifecycleState: domain.LifecycleApproved,
		Evaluation:     &evaluation,
		EvaluatedAt:    policyTestTime().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("EvaluateProject(approved) error = %v", err)
	}
	if got, want := approved.Outcome.ScoreMultiplier.String(), "24/25"; got != want {
		t.Fatalf("ScoreMultiplier = %q, want %q", got, want)
	}

	candidate, err := engine.EvaluateProject(version, ProjectContext{
		OperationID:    "op-multipliers-candidate",
		ProjectID:      projectID,
		LifecycleState: domain.LifecycleCandidate,
		Evaluation:     &evaluation,
		EvaluatedAt:    policyTestTime().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("EvaluateProject(candidate) error = %v", err)
	}
	if !candidate.Outcome.Denied || !candidate.Outcome.SoftEffectsSuppressed {
		t.Fatalf("candidate outcome = %#v, want denied with soft effects suppressed", candidate.Outcome)
	}
	if got := candidate.Outcome.ScoreMultiplier.String(); got != "1" {
		t.Fatalf("suppressed multiplier = %q, want 1", got)
	}
}

func TestPolicySetActivationRejectsConflictsAndCumulativeBoundViolation(t *testing.T) {
	engine := NewV0Engine()
	baseline := mustBaselinePolicySet(t)
	set, _ := domain.NewPolicySet(baseline.PolicySetID())

	duplicateNowParameters, _ := domain.NewCapacityLimitParameters(domain.PlacementNow, 4)
	duplicateNow, _ := domain.NewCapacityLimitPolicyInstance("duplicate-now", "orientation.capacity.extra", duplicateNowParameters, 30, "conflicting now capacity")
	conflicting, err := domain.NewPolicySetVersion(domain.PolicySetVersionInput{
		ID:                             "policy-set-conflicting-capacity",
		PolicySet:                      set,
		CreatedAt:                      policyTestTime(),
		Instances:                      append(baseline.Instances(), duplicateNow),
		MinimumScoreMultiplierBasisPts: 10000,
		MaximumScoreMultiplierBasisPts: 10000,
	})
	if err != nil {
		t.Fatalf("construct conflicting policy set: %v", err)
	}
	if err := engine.ValidatePolicySetVersion(conflicting); err == nil {
		t.Fatal("duplicate now-capacity policy passed activation validation")
	}

	firstParameters, _ := domain.NewScoreMultiplierParameters(12000)
	secondParameters, _ := domain.NewScoreMultiplierParameters(12000)
	first, _ := domain.NewScoreMultiplierPolicyInstance("multiplier-a", "orientation.test.multiplier.a", firstParameters, 50, "test multiplier")
	second, _ := domain.NewScoreMultiplierPolicyInstance("multiplier-b", "orientation.test.multiplier.b", secondParameters, 60, "test multiplier")
	outsideBounds, err := domain.NewPolicySetVersion(domain.PolicySetVersionInput{
		ID:                             "policy-set-outside-bounds",
		PolicySet:                      set,
		CreatedAt:                      policyTestTime(),
		Instances:                      append(baseline.Instances(), first, second),
		MinimumScoreMultiplierBasisPts: 9000,
		MaximumScoreMultiplierBasisPts: 12500,
	})
	if err != nil {
		t.Fatalf("construct multiplier policy set: %v", err)
	}
	if err := engine.ValidatePolicySetVersion(outsideBounds); err == nil {
		t.Fatal("cumulative score multiplier outside policy-set bounds passed activation")
	}
}

func TestPolicyExceptionUseIsScopedExpiringRevocableAndUseLimited(t *testing.T) {
	engine := NewV0Engine()
	version := mustBaselinePolicySet(t)
	var confidenceInstance domain.PolicyInstance
	for _, instance := range version.Instances() {
		if instance.PolicyID() == PolicyEvaluationConfidence {
			confidenceInstance = instance
			break
		}
	}
	if confidenceInstance.ID() == "" {
		t.Fatal("baseline confidence policy not found")
	}
	base := policyTestTime()
	exception, err := domain.NewPolicyException(domain.PolicyExceptionInput{
		ID:                 "exception-1",
		PolicySetVersionID: version.ID(),
		PolicyID:           confidenceInstance.PolicyID(),
		PolicyInstanceID:   confidenceInstance.ID(),
		EvaluatorVersion:   confidenceInstance.EvaluatorVersion(),
		ProjectID:          "project-1",
		Phase:              confidenceInstance.Phase(),
		PermittedDeviation: "accept weak-confidence recommendation after explicit review",
		Actor:              mustMaintainer(t),
		Rationale:          "reviewed evidence manually",
		EvidenceIDs:        []domain.EvidenceReferenceID{"exception-evidence"},
		CreatedAt:          base,
		EffectiveAt:        base,
		ExpiresAt:          base.Add(48 * time.Hour),
		MaximumUses:        1,
	})
	if err != nil {
		t.Fatalf("NewPolicyException() error = %v", err)
	}
	at := base.Add(time.Hour)
	if err := engine.ValidateExceptionUse(version, exception, nil, nil, "project-1", confidenceInstance.Phase(), at); err != nil {
		t.Fatalf("valid exception rejected: %v", err)
	}

	application, err := domain.NewPolicyExceptionApplication("exception-application-1", exception, "project-1", "op-1", "policy-decision-1", at)
	if err != nil {
		t.Fatalf("NewPolicyExceptionApplication() error = %v", err)
	}
	if err := engine.ValidateExceptionUse(version, exception, []domain.PolicyExceptionApplication{application}, nil, "project-1", confidenceInstance.Phase(), at.Add(time.Minute)); err == nil {
		t.Fatal("one-shot exception remained reusable after application")
	}
	if err := engine.ValidateExceptionUse(version, exception, nil, nil, "project-2", confidenceInstance.Phase(), at); err == nil {
		t.Fatal("project-scoped exception applied to another project")
	}
	if err := engine.ValidateExceptionUse(version, exception, nil, nil, "project-1", confidenceInstance.Phase(), base.Add(49*time.Hour)); err == nil {
		t.Fatal("expired exception remained usable")
	}
	revocation, err := domain.NewPolicyExceptionRevocation("revocation-1", exception.ID(), mustMaintainer(t), "evidence changed", base.Add(30*time.Minute))
	if err != nil {
		t.Fatalf("NewPolicyExceptionRevocation() error = %v", err)
	}
	if err := engine.ValidateExceptionUse(version, exception, nil, []domain.PolicyExceptionRevocation{revocation}, "project-1", confidenceInstance.Phase(), at); err == nil {
		t.Fatal("revoked exception remained usable")
	}
}
