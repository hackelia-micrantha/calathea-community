package domain

import "testing"

func TestPolicyDecisionRejectsUndeclaredMissingInput(t *testing.T) {
	_, err := NewPolicyDecision(PolicyDecisionInput{
		ID:                         "decision-undeclared-missing",
		SchemaVersion:              PolicyDecisionSchemaVersionV1,
		PolicySetVersionID:         "policy-set-v1",
		PolicyID:                   "orientation.evaluation.required",
		PolicyInstanceID:           "required-evaluation",
		EvaluatorType:              PolicyEvaluatorRequiredEvaluation,
		EvaluatorVersion:           PolicyEvaluatorSemanticVersionV1,
		ConfigurationSchemaVersion: PolicyConfigurationSchemaVersionV1,
		Workflow:                   PolicyWorkflowOrientation,
		Phase:                      PolicyPhaseCandidateEligibility,
		SubjectType:                PolicySubjectProject,
		ProjectID:                  "project-1",
		OperationID:                "operation-1",
		Applicability:              PolicyApplicable,
		Result:                     PolicyDecisionIndeterminate,
		EffectClass:                PolicyEffectHard,
		RequiredInputs:             []string{"accepted_evaluation"},
		MissingInputs:              []string{"undeclared_input"},
		MissingInputBehavior:       PolicyMissingExcludeSubject,
		Rationale:                  "accepted evaluation is required",
		ReasonCode:                 "accepted_evaluation_missing",
		CreatedAt:                  testTime(),
	})
	if err == nil {
		t.Fatal("NewPolicyDecision() accepted a missing input not declared by required inputs")
	}
}
