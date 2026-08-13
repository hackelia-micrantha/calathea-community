package domain

import "testing"

func mustPolicyInstance(t *testing.T, input PolicyInstanceInput) PolicyInstance {
	t.Helper()
	instance, err := NewPolicyInstance(input)
	if err != nil {
		t.Fatalf("NewPolicyInstance() error = %v", err)
	}
	return instance
}

func TestPolicyInstanceRejectsInvalidParameterShape(t *testing.T) {
	params, err := NewCapacityLimitParameters(PlacementNow, 3)
	if err != nil {
		t.Fatalf("NewCapacityLimitParameters() error = %v", err)
	}
	_, err = NewPolicyInstance(PolicyInstanceInput{
		ID:                   PolicyInstanceID("evaluation-required"),
		PolicyID:             PolicyID("orientation.evaluation.required"),
		EvaluatorType:        PolicyEvaluatorRequiredEvaluation,
		EvaluatorVersion:     "1",
		Phase:                PolicyPhaseCandidateEligibility,
		EffectClass:          PolicyEffectHard,
		SubjectType:          PolicySubjectProject,
		RequiredInputs:       []PolicyInputKind{PolicyInputAcceptedEvaluation},
		MissingInputBehavior: PolicyMissingInputExcludeSubject,
		Exceptionability:     PolicyNotExceptionable,
		Parameters:           params,
		Rationale:            "accepted evaluation is required",
	})
	if err == nil {
		t.Fatal("required-evaluation policy accepted capacity parameters")
	}
}

func TestPolicyInstanceRejectsDuplicateRequiredInput(t *testing.T) {
	_, err := NewPolicyInstance(PolicyInstanceInput{
		ID:                   PolicyInstanceID("lifecycle"),
		PolicyID:             PolicyID("orientation.lifecycle.eligibility"),
		EvaluatorType:        PolicyEvaluatorLifecycleEligibility,
		EvaluatorVersion:     "1",
		Phase:                PolicyPhaseCandidateEligibility,
		EffectClass:          PolicyEffectHard,
		SubjectType:          PolicySubjectProject,
		RequiredInputs:       []PolicyInputKind{PolicyInputLifecycleState, PolicyInputLifecycleState},
		MissingInputBehavior: PolicyMissingInputExcludeSubject,
		Exceptionability:     PolicyNotExceptionable,
		Parameters:           NewLifecycleEligibilityParameters(false),
		Rationale:            "only approved or active work is eligible",
	})
	if err == nil {
		t.Fatal("policy instance accepted duplicate required input")
	}
}

func TestPolicySetVersionCopiesInstancesAndRejectsDuplicateInstanceID(t *testing.T) {
	instance := mustPolicyInstance(t, PolicyInstanceInput{
		ID:                   PolicyInstanceID("lifecycle"),
		PolicyID:             PolicyID("orientation.lifecycle.eligibility"),
		EvaluatorType:        PolicyEvaluatorLifecycleEligibility,
		EvaluatorVersion:     "1",
		Phase:                PolicyPhaseCandidateEligibility,
		EffectClass:          PolicyEffectHard,
		SubjectType:          PolicySubjectProject,
		RequiredInputs:       []PolicyInputKind{PolicyInputLifecycleState},
		MissingInputBehavior: PolicyMissingInputExcludeSubject,
		Exceptionability:     PolicyNotExceptionable,
		Parameters:           NewLifecycleEligibilityParameters(false),
		Rationale:            "lifecycle eligibility",
	})

	if _, err := NewPolicySetVersion(
		PolicySetVersionID("policy-set-v1"),
		PolicySetID("policy-set"),
		testTime(),
		instance,
		instance,
	); err == nil {
		t.Fatal("policy set version accepted duplicate policy instance id")
	}

	version, err := NewPolicySetVersion(
		PolicySetVersionID("policy-set-v1"),
		PolicySetID("policy-set"),
		testTime(),
		instance,
	)
	if err != nil {
		t.Fatalf("NewPolicySetVersion() error = %v", err)
	}

	instances := version.Instances()
	instances[0] = PolicyInstance{}
	if got := version.Instances()[0].ID(); got != PolicyInstanceID("lifecycle") {
		t.Fatalf("mutating returned instances changed version: ID() = %q", got)
	}

	required := version.Instances()[0].RequiredInputs()
	required[0] = PolicyInputSelectedCount
	if got := version.Instances()[0].RequiredInputs()[0]; got != PolicyInputLifecycleState {
		t.Fatalf("mutating returned required inputs changed version: got %q", got)
	}
}

func TestPolicyDecisionCopiesEffectsAndEvidence(t *testing.T) {
	subject, err := NewPlacementSetPolicySubject(PlacementNow)
	if err != nil {
		t.Fatalf("NewPlacementSetPolicySubject() error = %v", err)
	}
	effect, err := NewCapacityPolicyEffect(PlacementNow, 3)
	if err != nil {
		t.Fatalf("NewCapacityPolicyEffect() error = %v", err)
	}
	evidence := []EvidenceReferenceID{"evidence-1"}
	effects := []PolicyEffect{effect}

	decision, err := NewPolicyDecision(PolicyDecisionInput{
		ID:                   PolicyDecisionID("decision-1"),
		PolicySetVersionID:   PolicySetVersionID("policy-set-v1"),
		PolicyID:             PolicyID("orientation.capacity.now"),
		PolicyInstanceID:     PolicyInstanceID("capacity-now"),
		EvaluatorType:        PolicyEvaluatorCapacityLimit,
		EvaluatorVersion:     "1",
		Phase:                PolicyPhaseSetConstraints,
		EffectClass:          PolicyEffectHard,
		MissingInputBehavior: PolicyMissingInputFailOperation,
		Subject:              subject,
		OperationID:          OperationID("op-1"),
		Result:               PolicyDecisionAllow,
		ReasonCode:           "capacity_satisfied",
		EvidenceIDs:          evidence,
		Effects:              effects,
		CreatedAt:            testTime(),
	})
	if err != nil {
		t.Fatalf("NewPolicyDecision() error = %v", err)
	}

	evidence[0] = EvidenceReferenceID("changed")
	effects[0] = PolicyEffect{}
	returned := decision.Effects()
	returned[0] = PolicyEffect{}

	if got := decision.EvidenceIDs()[0]; got != EvidenceReferenceID("evidence-1") {
		t.Fatalf("stored evidence changed to %q", got)
	}
	stored := decision.Effects()[0]
	if got, want := stored.Type(), PolicyEffectCapacityLimit; got != want {
		t.Fatalf("stored effect type = %q, want %q", got, want)
	}
	if maximum, ok := stored.Maximum(); !ok || maximum != 3 {
		t.Fatalf("stored capacity maximum = %d, ok=%v, want 3,true", maximum, ok)
	}
}
