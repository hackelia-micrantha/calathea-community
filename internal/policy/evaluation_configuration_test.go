package policy

import (
	"errors"
	"testing"

	"github.com/hackelia-micrantha/calathea-community/internal/domain"
)

func TestEvaluateRejectsPolicyEvaluatorMismatchBeforeDecision(t *testing.T) {
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
		Rationale:            "mismatched evaluator must fail before a decision is emitted",
	})
	set, err := domain.NewPolicySetVersion(
		domain.PolicySetVersionID("policy-set-mismatch"),
		domain.PolicySetID("policy-set"),
		policyTestTime(),
		instances...,
	)
	if err != nil {
		t.Fatalf("NewPolicySetVersion() error = %v", err)
	}

	state := domain.LifecycleApproved
	req := request(t, set, domain.PolicyInstanceID("lifecycle-default"), projectSubject(t))
	req.Context.LifecycleState = &state
	if _, err := Evaluate(req); err == nil {
		t.Fatal("Evaluate() emitted a decision from mismatched policy configuration")
	} else {
		var failure *EvaluationFailure
		if !errors.As(err, &failure) {
			t.Fatalf("Evaluate() error = %T %v, want EvaluationFailure", err, err)
		}
		if got, want := failure.Code, "invalid_policy_configuration"; got != want {
			t.Fatalf("failure code = %q, want %q", got, want)
		}
	}
}
