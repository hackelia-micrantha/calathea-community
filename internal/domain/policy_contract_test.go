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

func TestPolicyInstanceExposesConstrainedReplayMetadata(t *testing.T) {
	instance := mustPolicyInstance(t, "required-evaluation", 10)
	if got := instance.ConfigurationSchemaVersion(); got != PolicyConfigurationSchemaVersionV1 {
		t.Fatalf("configuration schema = %q, want %q", got, PolicyConfigurationSchemaVersionV1)
	}
	if got := instance.Workflow(); got != PolicyWorkflowOrientation {
		t.Fatalf("workflow = %q, want orientation", got)
	}
	if got := instance.SubjectType(); got != PolicySubjectProject {
		t.Fatalf("subject type = %q, want project", got)
	}
	if got := instance.RequiredInputs(); len(got) != 1 || got[0] != "accepted_evaluation" {
		t.Fatalf("required inputs = %#v, want accepted_evaluation", got)
	}
	if !instance.SubjectSelector().MatchesProject("project-1") {
		t.Fatal("v0 all-project selector did not match project")
	}
}

func TestProjectSubjectSelectorIsExactAndDefensive(t *testing.T) {
	input := []ProjectID{"project-1", "project-2"}
	selector, err := NewProjectPolicySubjectSelector(input)
	if err != nil {
		t.Fatalf("NewProjectPolicySubjectSelector() error = %v", err)
	}
	input[0] = "mutated"
	if !selector.MatchesProject("project-1") || selector.MatchesProject("project-3") {
		t.Fatalf("selector matching is incorrect: %#v", selector.ProjectIDs())
	}
	returned := selector.ProjectIDs()
	returned[0] = "mutated-again"
	if !selector.MatchesProject("project-1") {
		t.Fatal("mutating returned selector values changed selector")
	}
	if _, err := NewProjectPolicySubjectSelector([]ProjectID{"project-1", "project-1"}); err == nil {
		t.Fatal("duplicate selector project id succeeded")
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
		ID:                         "decision-1",
		SchemaVersion:              PolicyDecisionSchemaVersionV1,
		PolicySetVersionID:         "policy-set-v1",
		PolicyID:                   "orientation.evaluation.confidence",
		PolicyInstanceID:           "confidence",
		EvaluatorType:              PolicyEvaluatorConfidenceGate,
		EvaluatorVersion:           PolicyEvaluatorSemanticVersionV1,
		ConfigurationSchemaVersion: PolicyConfigurationSchemaVersionV1,
		Workflow:                   PolicyWorkflowOrientation,
		Phase:                      PolicyPhaseCandidateAdjustment,
		SubjectType:                PolicySubjectProject,
		ProjectID:                  "project-1",
		OperationID:                "operation-1",
		Applicability:              PolicyApplicable,
		Result:                     PolicyDecisionRequireReview,
		EffectClass:                PolicyEffectReviewRequired,
		Effects:                    effects,
		RequiredInputs:             []string{"accepted_evaluation", "confidence_band"},
		InputReferences:            inputReferences,
		EvidenceIDs:                evidenceIDs,
		MissingInputBehavior:       PolicyMissingRequireReview,
		Rationale:                  "weak evidence requires explicit review",
		ReasonCode:                 "confidence_below_threshold",
		Priority:                   10,
		CreatedAt:                  testTime(),
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
	if decision.SchemaVersion() != PolicyDecisionSchemaVersionV1 || decision.Workflow() != PolicyWorkflowOrientation || decision.Applicability() != PolicyApplicable {
		t.Fatalf("decision replay metadata incomplete: schema=%q workflow=%q applicability=%q", decision.SchemaVersion(), decision.Workflow(), decision.Applicability())
	}
}

func TestPolicyDecisionRejectsApplicabilityAndIndeterminateMismatches(t *testing.T) {
	base := PolicyDecisionInput{
		ID:                         "decision-1",
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
		Result:                     PolicyDecisionAllow,
		EffectClass:                PolicyEffectHard,
		RequiredInputs:             []string{"accepted_evaluation"},
		InputReferences:            []string{"evaluation_version:evaluation-v1"},
		MissingInputBehavior:       PolicyMissingExcludeSubject,
		Rationale:                  "accepted evaluation is required",
		ReasonCode:                 "accepted_evaluation_present",
		CreatedAt:                  testTime(),
	}

	badApplicability := base
	badApplicability.Result = PolicyDecisionNotApplicable
	if _, err := NewPolicyDecision(badApplicability); err == nil {
		t.Fatal("not_applicable result with applicable flag succeeded")
	}

	badIndeterminate := base
	badIndeterminate.Result = PolicyDecisionIndeterminate
	if _, err := NewPolicyDecision(badIndeterminate); err == nil {
		t.Fatal("indeterminate decision without missing/conflicting evidence succeeded")
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
