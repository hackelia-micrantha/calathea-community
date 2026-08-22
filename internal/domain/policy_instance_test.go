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

func TestPolicySetVersionCopiesInstancesRejectsDuplicatesAndOrdersDeterministically(t *testing.T) {
	lifecycle := mustPolicyInstance(t, PolicyInstanceInput{
		ID:                   PolicyInstanceID("lifecycle"),
		PolicyID:             PolicyID("orientation.lifecycle.eligibility"),
		EvaluatorType:        PolicyEvaluatorLifecycleEligibility,
		EvaluatorVersion:     "1",
		Phase:                PolicyPhaseCandidateEligibility,
		EffectClass:          PolicyEffectHard,
		SubjectType:          PolicySubjectProject,
		RequiredInputs:       []PolicyInputKind{PolicyInputLifecycleState},
		MissingInputBehavior: PolicyMissingInputExcludeSubject,
		Priority:             20,
		Exceptionability:     PolicyNotExceptionable,
		Parameters:           NewLifecycleEligibilityParameters(false),
		Rationale:            "lifecycle eligibility",
	})
	required := mustPolicyInstance(t, PolicyInstanceInput{
		ID:                   PolicyInstanceID("evaluation-required"),
		PolicyID:             PolicyID("orientation.evaluation.required"),
		EvaluatorType:        PolicyEvaluatorRequiredEvaluation,
		EvaluatorVersion:     "1",
		Phase:                PolicyPhaseCandidateEligibility,
		EffectClass:          PolicyEffectHard,
		SubjectType:          PolicySubjectProject,
		RequiredInputs:       []PolicyInputKind{PolicyInputAcceptedEvaluation},
		MissingInputBehavior: PolicyMissingInputExcludeSubject,
		Priority:             10,
		Exceptionability:     PolicyNotExceptionable,
		Parameters:           NewNoPolicyParameters(),
		Rationale:            "accepted evaluation required",
	})

	if _, err := NewPolicySetVersion(
		PolicySetVersionID("policy-set-v1"),
		PolicySetID("policy-set"),
		testTime(),
		lifecycle,
		lifecycle,
	); err == nil {
		t.Fatal("policy set version accepted duplicate policy instance id")
	}

	version, err := NewPolicySetVersion(
		PolicySetVersionID("policy-set-v1"),
		PolicySetID("policy-set"),
		testTime(),
		lifecycle,
		required,
	)
	if err != nil {
		t.Fatalf("NewPolicySetVersion() error = %v", err)
	}

	instances := version.Instances()
	if got, want := instances[0].ID(), PolicyInstanceID("evaluation-required"); got != want {
		t.Fatalf("first ordered instance = %q, want %q", got, want)
	}
	instances[0] = PolicyInstance{}
	if got := version.Instances()[0].ID(); got != PolicyInstanceID("evaluation-required") {
		t.Fatalf("mutating returned instances changed version: ID() = %q", got)
	}

	inputs := version.Instances()[0].RequiredInputs()
	inputs[0] = PolicyInputSelectedCount
	if got := version.Instances()[0].RequiredInputs()[0]; got != PolicyInputAcceptedEvaluation {
		t.Fatalf("mutating returned required inputs changed version: got %q", got)
	}
}

func TestPolicyDecisionCopiesEffectsEvidenceAndInputTrace(t *testing.T) {
	subject, err := NewPlacementSetPolicySubject(PlacementNow)
	if err != nil {
		t.Fatalf("NewPlacementSetPolicySubject() error = %v", err)
	}
	effect, err := NewCapacityPolicyEffect(PlacementNow, 3)
	if err != nil {
		t.Fatalf("NewCapacityPolicyEffect() error = %v", err)
	}
	selectedCount, err := NewPolicyInputReference(PolicyInputSelectedCount, "3")
	if err != nil {
		t.Fatalf("NewPolicyInputReference() error = %v", err)
	}
	evidence := []EvidenceReferenceID{"evidence-1"}
	effects := []PolicyEffect{effect}
	requiredInputs := []PolicyInputKind{PolicyInputSelectedCount}
	inputReferences := []PolicyInputReference{selectedCount}

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
		RequiredInputs:       requiredInputs,
		InputReferences:      inputReferences,
		EvidenceIDs:          evidence,
		Effects:              effects,
		Priority:             30,
		Rationale:            "bound immediate work in progress",
		CreatedAt:            testTime(),
	})
	if err != nil {
		t.Fatalf("NewPolicyDecision() error = %v", err)
	}

	evidence[0] = EvidenceReferenceID("changed")
	effects[0] = PolicyEffect{}
	requiredInputs[0] = PolicyInputLifecycleState
	inputReferences[0] = PolicyInputReference{}
	returnedEffects := decision.Effects()
	returnedEffects[0] = PolicyEffect{}
	returnedReferences := decision.InputReferences()
	returnedReferences[0] = PolicyInputReference{}

	if got := decision.EvidenceIDs()[0]; got != EvidenceReferenceID("evidence-1") {
		t.Fatalf("stored evidence changed to %q", got)
	}
	if got := decision.RequiredInputs()[0]; got != PolicyInputSelectedCount {
		t.Fatalf("stored required input changed to %q", got)
	}
	if got := decision.InputReferences()[0]; got.Kind() != PolicyInputSelectedCount || got.Value() != "3" {
		t.Fatalf("stored input reference = %#v, want selected_count=3", got)
	}
	stored := decision.Effects()[0]
	if got, want := stored.Type(), PolicyEffectCapacityLimit; got != want {
		t.Fatalf("stored effect type = %q, want %q", got, want)
	}
	if maximum, ok := stored.Maximum(); !ok || maximum != 3 {
		t.Fatalf("stored capacity maximum = %d, ok=%v, want 3,true", maximum, ok)
	}
}

func TestPolicyDecisionRejectsUndeclaredMissingInput(t *testing.T) {
	subject, err := NewProjectPolicySubject(ProjectID("project-1"))
	if err != nil {
		t.Fatalf("NewProjectPolicySubject() error = %v", err)
	}
	_, err = NewPolicyDecision(PolicyDecisionInput{
		ID:                   PolicyDecisionID("decision-1"),
		PolicySetVersionID:   PolicySetVersionID("policy-set-v1"),
		PolicyID:             PolicyID("orientation.evaluation.required"),
		PolicyInstanceID:     PolicyInstanceID("evaluation-required"),
		EvaluatorType:        PolicyEvaluatorRequiredEvaluation,
		EvaluatorVersion:     "1",
		Phase:                PolicyPhaseCandidateEligibility,
		EffectClass:          PolicyEffectHard,
		MissingInputBehavior: PolicyMissingInputExcludeSubject,
		Subject:              subject,
		OperationID:          OperationID("op-1"),
		Result:               PolicyDecisionIndeterminate,
		ReasonCode:           "missing_accepted_evaluation",
		RequiredInputs:       []PolicyInputKind{PolicyInputAcceptedEvaluation},
		MissingInputs:        []PolicyInputKind{PolicyInputLifecycleState},
		Priority:             20,
		Rationale:            "accepted evaluation required",
		CreatedAt:            testTime(),
	})
	if err == nil {
		t.Fatal("policy decision accepted a missing input not declared by the evaluator contract")
	}
}

func TestPolicyDecisionRejectsApplicableResultWithoutRequiredInputReference(t *testing.T) {
	subject, err := NewProjectPolicySubject(ProjectID("project-1"))
	if err != nil {
		t.Fatalf("NewProjectPolicySubject() error = %v", err)
	}
	_, err = NewPolicyDecision(PolicyDecisionInput{
		ID:                   PolicyDecisionID("decision-1"),
		PolicySetVersionID:   PolicySetVersionID("policy-set-v1"),
		PolicyID:             PolicyID("orientation.evaluation.required"),
		PolicyInstanceID:     PolicyInstanceID("evaluation-required"),
		EvaluatorType:        PolicyEvaluatorRequiredEvaluation,
		EvaluatorVersion:     "1",
		Phase:                PolicyPhaseCandidateEligibility,
		EffectClass:          PolicyEffectHard,
		MissingInputBehavior: PolicyMissingInputExcludeSubject,
		Subject:              subject,
		OperationID:          OperationID("op-1"),
		Result:               PolicyDecisionAllow,
		ReasonCode:           "accepted_evaluation_present",
		RequiredInputs:       []PolicyInputKind{PolicyInputAcceptedEvaluation},
		Priority:             20,
		Rationale:            "accepted evaluation required",
		CreatedAt:            testTime(),
	})
	if err == nil {
		t.Fatal("policy decision accepted applicable result without required input reference")
	}
}
