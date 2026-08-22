package domain

import "testing"

func TestPolicyV1WorkflowMetadataIsExplicit(t *testing.T) {
	instance := mustPolicyInstance(t, PolicyInstanceInput{
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
		Parameters:           NewNoPolicyParameters(),
		Rationale:            "accepted evaluation required",
	})
	if got := instance.ConfigurationSchemaVersion(); got != PolicyConfigurationSchemaVersionV1 {
		t.Fatalf("ConfigurationSchemaVersion() = %q, want %q", got, PolicyConfigurationSchemaVersionV1)
	}
	if got := instance.Workflow(); got != PolicyWorkflowOrientation {
		t.Fatalf("Workflow() = %q, want %q", got, PolicyWorkflowOrientation)
	}
}

func TestPolicyDecisionApplicabilityIsDistinctFromAllow(t *testing.T) {
	subject, err := NewPlacementSetPolicySubject(PlacementNext)
	if err != nil {
		t.Fatalf("NewPlacementSetPolicySubject() error = %v", err)
	}
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
		Result:               PolicyDecisionNotApplicable,
		ReasonCode:           "placement_set_not_applicable",
		RequiredInputs:       []PolicyInputKind{PolicyInputSelectedCount},
		Priority:             30,
		Rationale:            "bound immediate work in progress",
		CreatedAt:            testTime(),
	})
	if err != nil {
		t.Fatalf("NewPolicyDecision() error = %v", err)
	}
	if got := decision.Applicability(); got != PolicyNotApplicable {
		t.Fatalf("Applicability() = %q, want %q", got, PolicyNotApplicable)
	}
	if got := decision.Workflow(); got != PolicyWorkflowOrientation {
		t.Fatalf("Workflow() = %q, want %q", got, PolicyWorkflowOrientation)
	}
}
