package policy

import (
	"errors"
	"testing"
	"time"

	"github.com/hackelia-micrantha/calathea-community/internal/domain"
)

func policyTestTime() time.Time {
	return time.Date(2026, time.August, 12, 12, 0, 0, 0, time.UTC)
}

func mustPolicyInstance(t *testing.T, input domain.PolicyInstanceInput) domain.PolicyInstance {
	t.Helper()
	instance, err := domain.NewPolicyInstance(input)
	if err != nil {
		t.Fatalf("NewPolicyInstance() error = %v", err)
	}
	return instance
}

func baselineInstances(t *testing.T) []domain.PolicyInstance {
	t.Helper()
	nowParams, err := domain.NewCapacityLimitParameters(domain.PlacementNow, 3)
	if err != nil {
		t.Fatalf("NewCapacityLimitParameters(now) error = %v", err)
	}
	nextParams, err := domain.NewCapacityLimitParameters(domain.PlacementNext, 10)
	if err != nil {
		t.Fatalf("NewCapacityLimitParameters(next) error = %v", err)
	}
	confidenceParams, err := domain.NewConfidenceGateParameters(4000)
	if err != nil {
		t.Fatalf("NewConfidenceGateParameters() error = %v", err)
	}
	freshnessParams, err := domain.NewFreshnessRuleParameters(90)
	if err != nil {
		t.Fatalf("NewFreshnessRuleParameters() error = %v", err)
	}

	return []domain.PolicyInstance{
		mustPolicyInstance(t, domain.PolicyInstanceInput{
			ID:                   domain.PolicyInstanceID("lifecycle-default"),
			PolicyID:             PolicyLifecycleEligibility,
			EvaluatorType:        domain.PolicyEvaluatorLifecycleEligibility,
			EvaluatorVersion:     EvaluatorSemanticVersionV1,
			Phase:                domain.PolicyPhaseCandidateEligibility,
			EffectClass:          domain.PolicyEffectHard,
			SubjectType:          domain.PolicySubjectProject,
			RequiredInputs:       []domain.PolicyInputKind{domain.PolicyInputLifecycleState},
			MissingInputBehavior: domain.PolicyMissingInputExcludeSubject,
			Priority:             10,
			Exceptionability:     domain.PolicyNotExceptionable,
			Parameters:           domain.NewLifecycleEligibilityParameters(false),
			Rationale:            "exclude projects outside active orientation lifecycle states",
		}),
		mustPolicyInstance(t, domain.PolicyInstanceInput{
			ID:                   domain.PolicyInstanceID("evaluation-required-default"),
			PolicyID:             PolicyRequiredEvaluation,
			EvaluatorType:        domain.PolicyEvaluatorRequiredEvaluation,
			EvaluatorVersion:     EvaluatorSemanticVersionV1,
			Phase:                domain.PolicyPhaseCandidateEligibility,
			EffectClass:          domain.PolicyEffectHard,
			SubjectType:          domain.PolicySubjectProject,
			RequiredInputs:       []domain.PolicyInputKind{domain.PolicyInputAcceptedEvaluation},
			MissingInputBehavior: domain.PolicyMissingInputExcludeSubject,
			Priority:             20,
			Exceptionability:     domain.PolicyNotExceptionable,
			Parameters:           domain.NewNoPolicyParameters(),
			Rationale:            "require an accepted evaluation for active orientation",
		}),
		mustPolicyInstance(t, domain.PolicyInstanceInput{
			ID:                   domain.PolicyInstanceID("capacity-now-default"),
			PolicyID:             PolicyCapacityNow,
			EvaluatorType:        domain.PolicyEvaluatorCapacityLimit,
			EvaluatorVersion:     EvaluatorSemanticVersionV1,
			Phase:                domain.PolicyPhaseSetConstraints,
			EffectClass:          domain.PolicyEffectHard,
			SubjectType:          domain.PolicySubjectPlacementSet,
			RequiredInputs:       []domain.PolicyInputKind{domain.PolicyInputSelectedCount},
			MissingInputBehavior: domain.PolicyMissingInputFailOperation,
			Priority:             30,
			Exceptionability:     domain.PolicyNotExceptionable,
			Parameters:           nowParams,
			Rationale:            "bound immediate work in progress",
		}),
		mustPolicyInstance(t, domain.PolicyInstanceInput{
			ID:                   domain.PolicyInstanceID("capacity-next-default"),
			PolicyID:             PolicyCapacityNext,
			EvaluatorType:        domain.PolicyEvaluatorCapacityLimit,
			EvaluatorVersion:     EvaluatorSemanticVersionV1,
			Phase:                domain.PolicyPhaseSetConstraints,
			EffectClass:          domain.PolicyEffectHard,
			SubjectType:          domain.PolicySubjectPlacementSet,
			RequiredInputs:       []domain.PolicyInputKind{domain.PolicyInputSelectedCount},
			MissingInputBehavior: domain.PolicyMissingInputFailOperation,
			Priority:             40,
			Exceptionability:     domain.PolicyNotExceptionable,
			Parameters:           nextParams,
			Rationale:            "bound near-term ready queue",
		}),
		mustPolicyInstance(t, domain.PolicyInstanceInput{
			ID:                   domain.PolicyInstanceID("confidence-default"),
			PolicyID:             PolicyConfidence,
			EvaluatorType:        domain.PolicyEvaluatorConfidenceGate,
			EvaluatorVersion:     EvaluatorSemanticVersionV1,
			Phase:                domain.PolicyPhaseCandidateAdjustment,
			EffectClass:          domain.PolicyEffectReviewRequired,
			SubjectType:          domain.PolicySubjectProject,
			RequiredInputs:       []domain.PolicyInputKind{domain.PolicyInputAcceptedEvaluation},
			MissingInputBehavior: domain.PolicyMissingInputRequireReview,
			Priority:             50,
			Exceptionability:     domain.PolicyNotExceptionable,
			Parameters:           confidenceParams,
			Rationale:            "surface weak evaluation evidence",
		}),
		mustPolicyInstance(t, domain.PolicyInstanceInput{
			ID:                   domain.PolicyInstanceID("freshness-default"),
			PolicyID:             PolicyFreshness,
			EvaluatorType:        domain.PolicyEvaluatorFreshnessRule,
			EvaluatorVersion:     EvaluatorSemanticVersionV1,
			Phase:                domain.PolicyPhaseCandidateAdjustment,
			EffectClass:          domain.PolicyEffectReviewRequired,
			SubjectType:          domain.PolicySubjectProject,
			RequiredInputs:       []domain.PolicyInputKind{domain.PolicyInputAcceptedEvaluation, domain.PolicyInputAsOfTime},
			MissingInputBehavior: domain.PolicyMissingInputRequireReview,
			Priority:             60,
			Exceptionability:     domain.PolicyNotExceptionable,
			Parameters:           freshnessParams,
			Rationale:            "surface stale evaluation evidence",
		}),
	}
}

func baselinePolicySet(t *testing.T) domain.PolicySetVersion {
	t.Helper()
	version, err := domain.NewPolicySetVersion(
		domain.PolicySetVersionID("policy-set-v1"),
		domain.PolicySetID("policy-set"),
		policyTestTime(),
		baselineInstances(t)...,
	)
	if err != nil {
		t.Fatalf("NewPolicySetVersion() error = %v", err)
	}
	return version
}

func projectSubject(t *testing.T) domain.PolicySubject {
	t.Helper()
	subject, err := domain.NewProjectPolicySubject(domain.ProjectID("project-1"))
	if err != nil {
		t.Fatalf("NewProjectPolicySubject() error = %v", err)
	}
	return subject
}

func placementSubject(t *testing.T, placement domain.Placement) domain.PolicySubject {
	t.Helper()
	subject, err := domain.NewPlacementSetPolicySubject(placement)
	if err != nil {
		t.Fatalf("NewPlacementSetPolicySubject() error = %v", err)
	}
	return subject
}

func testEvaluation(t *testing.T, confidenceBasisPoints int, evidenceAsOf time.Time) domain.EvaluationVersion {
	t.Helper()
	projectID := domain.ProjectID("project-1")
	evaluation, err := domain.NewEvaluation(domain.EvaluationID("evaluation-1"), projectID)
	if err != nil {
		t.Fatalf("NewEvaluation() error = %v", err)
	}
	projectVersion, err := domain.NewProjectVersion(domain.ProjectVersionID("project-v1"), projectID, "Test project", policyTestTime(), nil)
	if err != nil {
		t.Fatalf("NewProjectVersion() error = %v", err)
	}
	axis := func(value int, rationale string) domain.AxisAssessment {
		assessment, axisErr := domain.NewAxisAssessment(value, rationale)
		if axisErr != nil {
			t.Fatalf("NewAxisAssessment() error = %v", axisErr)
		}
		return assessment
	}
	axes, err := domain.NewEvaluationAxes(
		axis(3, "material impact"),
		axis(3, "bounded effort"),
		axis(3, "known risk reduction"),
		axis(1, "no additional future path claimed"),
		axis(3, "planning-horizon relevance"),
		nil,
	)
	if err != nil {
		t.Fatalf("NewEvaluationAxes() error = %v", err)
	}
	confidence, err := domain.NewConfidence(confidenceBasisPoints, "test evidence quality")
	if err != nil {
		t.Fatalf("NewConfidence() error = %v", err)
	}
	freshness, err := domain.NewFreshnessMetadata(evidenceAsOf)
	if err != nil {
		t.Fatalf("NewFreshnessMetadata() error = %v", err)
	}
	actor, err := domain.NewMaintainerActor(domain.ActorID("maintainer"))
	if err != nil {
		t.Fatalf("NewMaintainerActor() error = %v", err)
	}
	version, err := domain.NewEvaluationVersion(domain.EvaluationVersionInput{
		ID:                     domain.EvaluationVersionID("evaluation-v1"),
		Evaluation:             evaluation,
		ProjectVersion:         projectVersion,
		EvaluatedAt:            policyTestTime(),
		PlanningHorizon:        "next 90 days",
		Freshness:              freshness,
		Axes:                   axes,
		Confidence:             confidence,
		Derivation:             domain.EvaluationDerivationAuthored,
		SemanticVersion:        domain.EvaluationSemanticVersionV1,
		FormulaSemanticVersion: domain.BaseScoreFormulaSemanticVersionV1,
		AcceptedBy:             actor,
		AcceptedAt:             policyTestTime(),
	})
	if err != nil {
		t.Fatalf("NewEvaluationVersion() error = %v", err)
	}
	return version
}

func request(t *testing.T, policySet domain.PolicySetVersion, instanceID domain.PolicyInstanceID, subject domain.PolicySubject) Request {
	t.Helper()
	return Request{
		DecisionID:  domain.PolicyDecisionID("decision-1"),
		PolicySet:   policySet,
		InstanceID:  instanceID,
		OperationID: domain.OperationID("op-1"),
		CreatedAt:   policyTestTime(),
		Context: Context{
			Subject: subject,
		},
	}
}

func TestValidateForActivationAcceptsCompleteBaseline(t *testing.T) {
	if err := ValidateForActivation(baselinePolicySet(t)); err != nil {
		t.Fatalf("ValidateForActivation() error = %v", err)
	}
}

func TestValidateForActivationRejectsMissingBaselinePolicy(t *testing.T) {
	instances := baselineInstances(t)
	instances = instances[:len(instances)-1]
	version, err := domain.NewPolicySetVersion(domain.PolicySetVersionID("policy-set-v1"), domain.PolicySetID("policy-set"), policyTestTime(), instances...)
	if err != nil {
		t.Fatalf("NewPolicySetVersion() error = %v", err)
	}
	if err := ValidateForActivation(version); err == nil {
		t.Fatal("incomplete policy set activated")
	}
}

func TestValidateForActivationRejectsPolicyEvaluatorMismatch(t *testing.T) {
	instances := baselineInstances(t)
	instances[0] = mustPolicyInstance(t, domain.PolicyInstanceInput{
		ID:                   domain.PolicyInstanceID("lifecycle-default"),
		PolicyID:             PolicyLifecycleEligibility,
		EvaluatorType:        domain.PolicyEvaluatorRequiredEvaluation,
		EvaluatorVersion:     EvaluatorSemanticVersionV1,
		Phase:                domain.PolicyPhaseCandidateEligibility,
		EffectClass:          domain.PolicyEffectHard,
		SubjectType:          domain.PolicySubjectProject,
		RequiredInputs:       []domain.PolicyInputKind{domain.PolicyInputAcceptedEvaluation},
		MissingInputBehavior: domain.PolicyMissingInputExcludeSubject,
		Exceptionability:     domain.PolicyNotExceptionable,
		Parameters:           domain.NewNoPolicyParameters(),
		Rationale:            "mismatched evaluator should fail activation",
	})
	version, err := domain.NewPolicySetVersion(domain.PolicySetVersionID("policy-set-v1"), domain.PolicySetID("policy-set"), policyTestTime(), instances...)
	if err != nil {
		t.Fatalf("NewPolicySetVersion() error = %v", err)
	}
	if err := ValidateForActivation(version); err == nil {
		t.Fatal("policy set with evaluator mismatch activated")
	}
}

func TestLifecycleEvaluator(t *testing.T) {
	set := baselinePolicySet(t)
	cases := []struct {
		name  string
		state domain.LifecycleState
		want  domain.PolicyDecisionResult
	}{
		{name: "approved", state: domain.LifecycleApproved, want: domain.PolicyDecisionAllow},
		{name: "active", state: domain.LifecycleActive, want: domain.PolicyDecisionAllow},
		{name: "paused", state: domain.LifecyclePaused, want: domain.PolicyDecisionAllow},
		{name: "proposed", state: domain.LifecycleProposed, want: domain.PolicyDecisionDeny},
		{name: "candidate", state: domain.LifecycleCandidate, want: domain.PolicyDecisionDeny},
		{name: "completed", state: domain.LifecycleCompleted, want: domain.PolicyDecisionDeny},
		{name: "cancelled", state: domain.LifecycleCancelled, want: domain.PolicyDecisionDeny},
		{name: "archived", state: domain.LifecycleArchived, want: domain.PolicyDecisionDeny},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := request(t, set, domain.PolicyInstanceID("lifecycle-default"), projectSubject(t))
			req.Context.LifecycleState = &tc.state
			decision, err := Evaluate(req)
			if err != nil {
				t.Fatalf("Evaluate() error = %v", err)
			}
			if got := decision.Result(); got != tc.want {
				t.Fatalf("Result() = %q, want %q", got, tc.want)
			}
		})
	}

	req := request(t, set, domain.PolicyInstanceID("lifecycle-default"), projectSubject(t))
	decision, err := Evaluate(req)
	if err != nil {
		t.Fatalf("Evaluate(missing lifecycle) error = %v", err)
	}
	if got := decision.Result(); got != domain.PolicyDecisionIndeterminate {
		t.Fatalf("missing lifecycle Result() = %q, want %q", got, domain.PolicyDecisionIndeterminate)
	}
	if got := decision.MissingInputs(); len(got) != 1 || got[0] != domain.PolicyInputLifecycleState {
		t.Fatalf("missing lifecycle inputs = %#v", got)
	}
}

func TestRequiredEvaluationEvaluator(t *testing.T) {
	set := baselinePolicySet(t)
	req := request(t, set, domain.PolicyInstanceID("evaluation-required-default"), projectSubject(t))
	decision, err := Evaluate(req)
	if err != nil {
		t.Fatalf("Evaluate(missing evaluation) error = %v", err)
	}
	if got := decision.Result(); got != domain.PolicyDecisionIndeterminate {
		t.Fatalf("missing evaluation Result() = %q, want %q", got, domain.PolicyDecisionIndeterminate)
	}
	if got := decision.MissingInputBehavior(); got != domain.PolicyMissingInputExcludeSubject {
		t.Fatalf("missing-input behavior = %q, want %q", got, domain.PolicyMissingInputExcludeSubject)
	}
	if got := decision.MissingInputs(); len(got) != 1 || got[0] != domain.PolicyInputAcceptedEvaluation {
		t.Fatalf("missing evaluation inputs = %#v", got)
	}

	evaluation := testEvaluation(t, 7000, policyTestTime())
	req.Context.Evaluation = &evaluation
	decision, err = Evaluate(req)
	if err != nil {
		t.Fatalf("Evaluate(present evaluation) error = %v", err)
	}
	if got := decision.Result(); got != domain.PolicyDecisionAllow {
		t.Fatalf("present evaluation Result() = %q, want %q", got, domain.PolicyDecisionAllow)
	}
}

func TestCapacityEvaluator(t *testing.T) {
	set := baselinePolicySet(t)
	req := request(t, set, domain.PolicyInstanceID("capacity-now-default"), placementSubject(t, domain.PlacementNow))
	count := 2
	req.Context.SelectedCount = &count
	decision, err := Evaluate(req)
	if err != nil {
		t.Fatalf("Evaluate(capacity available) error = %v", err)
	}
	if got := decision.Result(); got != domain.PolicyDecisionAllow {
		t.Fatalf("capacity available Result() = %q, want %q", got, domain.PolicyDecisionAllow)
	}
	if len(decision.Effects()) != 1 {
		t.Fatalf("capacity effects = %d, want 1", len(decision.Effects()))
	}

	count = 3
	decision, err = Evaluate(req)
	if err != nil {
		t.Fatalf("Evaluate(capacity exhausted) error = %v", err)
	}
	if got := decision.Result(); got != domain.PolicyDecisionDeny {
		t.Fatalf("capacity exhausted Result() = %q, want %q", got, domain.PolicyDecisionDeny)
	}

	other := request(t, set, domain.PolicyInstanceID("capacity-now-default"), placementSubject(t, domain.PlacementNext))
	other.Context.SelectedCount = &count
	decision, err = Evaluate(other)
	if err != nil {
		t.Fatalf("Evaluate(other placement) error = %v", err)
	}
	if got := decision.Result(); got != domain.PolicyDecisionNotApplicable {
		t.Fatalf("other placement Result() = %q, want %q", got, domain.PolicyDecisionNotApplicable)
	}
}

func TestConfidenceEvaluatorRequiresReviewForWeakEvidence(t *testing.T) {
	set := baselinePolicySet(t)
	weak := testEvaluation(t, 3999, policyTestTime())
	req := request(t, set, domain.PolicyInstanceID("confidence-default"), projectSubject(t))
	req.Context.Evaluation = &weak
	decision, err := Evaluate(req)
	if err != nil {
		t.Fatalf("Evaluate(weak confidence) error = %v", err)
	}
	if got := decision.Result(); got != domain.PolicyDecisionRequireReview {
		t.Fatalf("weak confidence Result() = %q, want %q", got, domain.PolicyDecisionRequireReview)
	}

	req.Context.Evaluation = nil
	decision, err = Evaluate(req)
	if err != nil {
		t.Fatalf("Evaluate(missing confidence input) error = %v", err)
	}
	if got := decision.Result(); got != domain.PolicyDecisionIndeterminate {
		t.Fatalf("missing confidence input Result() = %q, want %q", got, domain.PolicyDecisionIndeterminate)
	}
}

func TestFreshnessEvaluatorRequiresReviewForStaleEvidence(t *testing.T) {
	set := baselinePolicySet(t)
	stale := testEvaluation(t, 7000, policyTestTime().AddDate(0, 0, -91))
	req := request(t, set, domain.PolicyInstanceID("freshness-default"), projectSubject(t))
	req.Context.Evaluation = &stale
	req.Context.AsOf = policyTestTime()
	decision, err := Evaluate(req)
	if err != nil {
		t.Fatalf("Evaluate(stale) error = %v", err)
	}
	if got := decision.Result(); got != domain.PolicyDecisionRequireReview {
		t.Fatalf("stale Result() = %q, want %q", got, domain.PolicyDecisionRequireReview)
	}

	req.Context.AsOf = time.Time{}
	decision, err = Evaluate(req)
	if err != nil {
		t.Fatalf("Evaluate(missing as-of) error = %v", err)
	}
	if got := decision.Result(); got != domain.PolicyDecisionIndeterminate {
		t.Fatalf("missing as-of Result() = %q, want %q", got, domain.PolicyDecisionIndeterminate)
	}
	if got := decision.MissingInputs(); len(got) != 1 || got[0] != domain.PolicyInputAsOfTime {
		t.Fatalf("missing freshness inputs = %#v", got)
	}
}

func TestFreshnessEvaluatorRejectsImpossibleAsOfTime(t *testing.T) {
	set := baselinePolicySet(t)
	evaluation := testEvaluation(t, 7000, policyTestTime())
	req := request(t, set, domain.PolicyInstanceID("freshness-default"), projectSubject(t))
	req.Context.Evaluation = &evaluation
	req.Context.AsOf = policyTestTime().Add(-time.Second)

	_, err := Evaluate(req)
	var failure *EvaluationFailure
	if !errors.As(err, &failure) {
		t.Fatalf("Evaluate() error = %T %v, want EvaluationFailure", err, err)
	}
	if got, want := failure.Code, "invalid_freshness_context"; got != want {
		t.Fatalf("failure code = %q, want %q", got, want)
	}
}

func TestEvaluateRejectsUnsupportedEvaluatorVersionAsFailure(t *testing.T) {
	instances := baselineInstances(t)
	original := instances[0]
	instances[0] = mustPolicyInstance(t, domain.PolicyInstanceInput{
		ID:                   original.ID(),
		PolicyID:             original.PolicyID(),
		EvaluatorType:        original.EvaluatorType(),
		EvaluatorVersion:     "2",
		Phase:                original.Phase(),
		EffectClass:          original.EffectClass(),
		SubjectType:          original.SubjectType(),
		RequiredInputs:       original.RequiredInputs(),
		MissingInputBehavior: original.MissingInputBehavior(),
		Priority:             original.Priority(),
		Exceptionability:     original.Exceptionability(),
		Parameters:           original.Parameters(),
		Rationale:            original.Rationale(),
	})
	set, err := domain.NewPolicySetVersion(domain.PolicySetVersionID("policy-set-v2"), domain.PolicySetID("policy-set"), policyTestTime(), instances...)
	if err != nil {
		t.Fatalf("NewPolicySetVersion() error = %v", err)
	}
	if err := ValidateForActivation(set); err == nil {
		t.Fatal("unsupported evaluator version activated")
	}

	state := domain.LifecycleApproved
	req := request(t, set, original.ID(), projectSubject(t))
	req.Context.LifecycleState = &state
	_, err = Evaluate(req)
	var failure *EvaluationFailure
	if !errors.As(err, &failure) {
		t.Fatalf("Evaluate() error = %T %v, want EvaluationFailure", err, err)
	}
	if got, want := failure.Code, "unsupported_evaluator_version"; got != want {
		t.Fatalf("failure code = %q, want %q", got, want)
	}
}
