package domain

import (
	"fmt"
	"time"
)

// PolicyDecisionResult is the semantic outcome of one policy evaluator.
// Operational evaluator failure is deliberately not a result value.
type PolicyDecisionResult string

const (
	PolicyDecisionAllow         PolicyDecisionResult = "allow"
	PolicyDecisionDeny          PolicyDecisionResult = "deny"
	PolicyDecisionAdjust        PolicyDecisionResult = "adjust"
	PolicyDecisionRequireReview PolicyDecisionResult = "require_review"
	PolicyDecisionNotApplicable PolicyDecisionResult = "not_applicable"
	PolicyDecisionIndeterminate PolicyDecisionResult = "indeterminate"
)

func (r PolicyDecisionResult) Valid() bool {
	switch r {
	case PolicyDecisionAllow, PolicyDecisionDeny, PolicyDecisionAdjust,
		PolicyDecisionRequireReview, PolicyDecisionNotApplicable, PolicyDecisionIndeterminate:
		return true
	default:
		return false
	}
}

type PolicyApplicability string

const (
	PolicyApplicable    PolicyApplicability = "applicable"
	PolicyNotApplicable PolicyApplicability = "not_applicable"
)

func (a PolicyApplicability) Valid() bool {
	return a == PolicyApplicable || a == PolicyNotApplicable
}

// PolicyDecision is the immutable deterministic result of applying one policy
// instance to one project in one operation. It is self-describing enough for
// replay/explanation without consulting mutable policy defaults or ambient
// evaluator state.
type PolicyDecision struct {
	id                         PolicyDecisionID
	schemaVersion              string
	policySetVersionID         PolicySetVersionID
	policyID                   PolicyID
	policyInstanceID           PolicyInstanceID
	evaluatorType              PolicyEvaluatorType
	evaluatorVersion           string
	configurationSchemaVersion string
	workflow                   PolicyWorkflow
	phase                      PolicyPhase
	subjectType                PolicySubjectType
	projectID                  ProjectID
	operationID                OperationID
	applicability              PolicyApplicability
	result                     PolicyDecisionResult
	effectClass                PolicyEffectClass
	effects                    []PolicyEffect
	requiredInputs             []string
	inputReferences            []string
	evidenceIDs                []EvidenceReferenceID
	conflictingEvidenceIDs     []EvidenceReferenceID
	missingInputs              []string
	missingInputBehavior       PolicyMissingInputBehavior
	rationale                  string
	reasonCode                 string
	priority                   int
	conflictKey                string
	exceptionID                *PolicyExceptionID
	createdAt                  time.Time
}

type PolicyDecisionInput struct {
	ID                         PolicyDecisionID
	SchemaVersion              string
	PolicySetVersionID         PolicySetVersionID
	PolicyID                   PolicyID
	PolicyInstanceID           PolicyInstanceID
	EvaluatorType              PolicyEvaluatorType
	EvaluatorVersion           string
	ConfigurationSchemaVersion string
	Workflow                   PolicyWorkflow
	Phase                      PolicyPhase
	SubjectType                PolicySubjectType
	ProjectID                  ProjectID
	OperationID                OperationID
	Applicability              PolicyApplicability
	Result                     PolicyDecisionResult
	EffectClass                PolicyEffectClass
	Effects                    []PolicyEffect
	RequiredInputs             []string
	InputReferences            []string
	EvidenceIDs                []EvidenceReferenceID
	ConflictingEvidenceIDs     []EvidenceReferenceID
	MissingInputs              []string
	MissingInputBehavior       PolicyMissingInputBehavior
	Rationale                  string
	ReasonCode                 string
	Priority                   int
	ConflictKey                string
	ExceptionID                *PolicyExceptionID
	CreatedAt                  time.Time
}

func NewPolicyDecision(input PolicyDecisionInput) (PolicyDecision, error) {
	if err := requireIdentifier("policy decision id", string(input.ID)); err != nil {
		return PolicyDecision{}, err
	}
	if input.SchemaVersion != PolicyDecisionSchemaVersionV1 {
		return PolicyDecision{}, fmt.Errorf("unsupported policy decision schema version %q", input.SchemaVersion)
	}
	if err := requireIdentifier("policy set version id", string(input.PolicySetVersionID)); err != nil {
		return PolicyDecision{}, err
	}
	if err := requireIdentifier("policy id", string(input.PolicyID)); err != nil {
		return PolicyDecision{}, err
	}
	if err := requireIdentifier("policy instance id", string(input.PolicyInstanceID)); err != nil {
		return PolicyDecision{}, err
	}
	if !input.EvaluatorType.Valid() {
		return PolicyDecision{}, fmt.Errorf("invalid policy evaluator type %q", input.EvaluatorType)
	}
	if err := requireText("policy evaluator version", input.EvaluatorVersion); err != nil {
		return PolicyDecision{}, err
	}
	if input.ConfigurationSchemaVersion != PolicyConfigurationSchemaVersionV1 {
		return PolicyDecision{}, fmt.Errorf("unsupported policy configuration schema version %q", input.ConfigurationSchemaVersion)
	}
	if !input.Workflow.Valid() {
		return PolicyDecision{}, fmt.Errorf("invalid policy workflow %q", input.Workflow)
	}
	if !input.Phase.Valid() {
		return PolicyDecision{}, fmt.Errorf("invalid policy phase %q", input.Phase)
	}
	if !input.SubjectType.Valid() {
		return PolicyDecision{}, fmt.Errorf("invalid policy subject type %q", input.SubjectType)
	}
	if err := requireIdentifier("project id", string(input.ProjectID)); err != nil {
		return PolicyDecision{}, err
	}
	if err := requireIdentifier("operation id", string(input.OperationID)); err != nil {
		return PolicyDecision{}, err
	}
	if !input.Applicability.Valid() {
		return PolicyDecision{}, fmt.Errorf("invalid policy applicability %q", input.Applicability)
	}
	if !input.Result.Valid() {
		return PolicyDecision{}, fmt.Errorf("invalid policy decision result %q", input.Result)
	}
	if input.Result == PolicyDecisionNotApplicable && input.Applicability != PolicyNotApplicable {
		return PolicyDecision{}, fmt.Errorf("not_applicable result requires not_applicable applicability")
	}
	if input.Result != PolicyDecisionNotApplicable && input.Applicability != PolicyApplicable {
		return PolicyDecision{}, fmt.Errorf("result %q requires applicable applicability", input.Result)
	}
	if !input.EffectClass.Valid() {
		return PolicyDecision{}, fmt.Errorf("invalid policy effect class %q", input.EffectClass)
	}
	if input.Priority < 0 {
		return PolicyDecision{}, fmt.Errorf("policy decision priority must not be negative")
	}
	if !input.MissingInputBehavior.Valid() {
		return PolicyDecision{}, fmt.Errorf("invalid missing-input behavior %q", input.MissingInputBehavior)
	}
	if err := requireText("policy rationale", input.Rationale); err != nil {
		return PolicyDecision{}, err
	}
	if err := requireText("policy reason code", input.ReasonCode); err != nil {
		return PolicyDecision{}, err
	}
	if input.CreatedAt.IsZero() {
		return PolicyDecision{}, errZeroTime("policy decision time")
	}
	if len(input.RequiredInputs) == 0 {
		return PolicyDecision{}, fmt.Errorf("policy decision requires explicit required-input declarations")
	}
	if err := validateUniqueStrings("policy required input", input.RequiredInputs); err != nil {
		return PolicyDecision{}, err
	}
	if err := validateUniqueStrings("policy decision input reference", input.InputReferences); err != nil {
		return PolicyDecision{}, err
	}
	if err := validateUniqueEvidenceIDs("policy decision evidence reference id", input.EvidenceIDs); err != nil {
		return PolicyDecision{}, err
	}
	if err := validateUniqueEvidenceIDs("policy decision conflicting evidence reference id", input.ConflictingEvidenceIDs); err != nil {
		return PolicyDecision{}, err
	}
	if err := validateUniqueStrings("policy missing input", input.MissingInputs); err != nil {
		return PolicyDecision{}, err
	}
	requiredSet := make(map[string]struct{}, len(input.RequiredInputs))
	for _, required := range input.RequiredInputs {
		requiredSet[required] = struct{}{}
	}
	for _, missing := range input.MissingInputs {
		if _, ok := requiredSet[missing]; !ok {
			return PolicyDecision{}, fmt.Errorf("policy missing input %q is not declared in required inputs", missing)
		}
	}
	if input.Result == PolicyDecisionIndeterminate && len(input.MissingInputs) == 0 && len(input.ConflictingEvidenceIDs) == 0 {
		return PolicyDecision{}, fmt.Errorf("indeterminate policy decision requires missing or conflicting input evidence")
	}
	for _, effect := range input.Effects {
		switch effect.Kind() {
		case PolicyEffectCapacityLimit:
			if input.EffectClass != PolicyEffectHard || input.Phase != PolicyPhaseSetConstraints {
				return PolicyDecision{}, fmt.Errorf("capacity effect requires hard set-constraint policy")
			}
		case PolicyEffectScoreMultiplier:
			if input.EffectClass != PolicyEffectSoft || input.Result != PolicyDecisionAdjust {
				return PolicyDecision{}, fmt.Errorf("score multiplier effect requires soft adjust result")
			}
		case PolicyEffectDiagnosticAnnotation:
			// Diagnostic annotations may accompany any visible result.
		default:
			return PolicyDecision{}, fmt.Errorf("unsupported policy effect kind %q", effect.Kind())
		}
	}
	if input.Result == PolicyDecisionAdjust && len(input.Effects) == 0 {
		return PolicyDecision{}, fmt.Errorf("adjust policy decision requires at least one typed effect")
	}
	if input.Result == PolicyDecisionDeny && input.EffectClass != PolicyEffectHard {
		return PolicyDecision{}, fmt.Errorf("deny policy decision requires hard effect class")
	}
	if input.Result == PolicyDecisionRequireReview && input.EffectClass != PolicyEffectReviewRequired {
		return PolicyDecision{}, fmt.Errorf("require_review decision requires review_required effect class")
	}
	if input.ConflictKey != "" {
		if err := requireText("policy conflict key", input.ConflictKey); err != nil {
			return PolicyDecision{}, err
		}
	}
	if input.ExceptionID != nil {
		if err := requireIdentifier("policy exception id", string(*input.ExceptionID)); err != nil {
			return PolicyDecision{}, err
		}
	}
	return PolicyDecision{
		id:                         input.ID,
		schemaVersion:              input.SchemaVersion,
		policySetVersionID:         input.PolicySetVersionID,
		policyID:                   input.PolicyID,
		policyInstanceID:           input.PolicyInstanceID,
		evaluatorType:              input.EvaluatorType,
		evaluatorVersion:           input.EvaluatorVersion,
		configurationSchemaVersion: input.ConfigurationSchemaVersion,
		workflow:                   input.Workflow,
		phase:                      input.Phase,
		subjectType:                input.SubjectType,
		projectID:                  input.ProjectID,
		operationID:                input.OperationID,
		applicability:              input.Applicability,
		result:                     input.Result,
		effectClass:                input.EffectClass,
		effects:                    clonePolicyEffects(input.Effects),
		requiredInputs:             cloneStrings(input.RequiredInputs),
		inputReferences:            cloneStrings(input.InputReferences),
		evidenceIDs:                cloneEvidenceIDs(input.EvidenceIDs),
		conflictingEvidenceIDs:     cloneEvidenceIDs(input.ConflictingEvidenceIDs),
		missingInputs:              cloneStrings(input.MissingInputs),
		missingInputBehavior:       input.MissingInputBehavior,
		rationale:                  input.Rationale,
		reasonCode:                 input.ReasonCode,
		priority:                   input.Priority,
		conflictKey:                input.ConflictKey,
		exceptionID:                clonePolicyExceptionID(input.ExceptionID),
		createdAt:                  input.CreatedAt,
	}, nil
}

func (d PolicyDecision) ID() PolicyDecisionID                   { return d.id }
func (d PolicyDecision) SchemaVersion() string                  { return d.schemaVersion }
func (d PolicyDecision) PolicySetVersionID() PolicySetVersionID { return d.policySetVersionID }
func (d PolicyDecision) PolicyID() PolicyID                     { return d.policyID }
func (d PolicyDecision) PolicyInstanceID() PolicyInstanceID     { return d.policyInstanceID }
func (d PolicyDecision) EvaluatorType() PolicyEvaluatorType     { return d.evaluatorType }
func (d PolicyDecision) EvaluatorVersion() string               { return d.evaluatorVersion }
func (d PolicyDecision) ConfigurationSchemaVersion() string     { return d.configurationSchemaVersion }
func (d PolicyDecision) Workflow() PolicyWorkflow               { return d.workflow }
func (d PolicyDecision) Phase() PolicyPhase                     { return d.phase }
func (d PolicyDecision) SubjectType() PolicySubjectType         { return d.subjectType }
func (d PolicyDecision) ProjectID() ProjectID                   { return d.projectID }
func (d PolicyDecision) OperationID() OperationID               { return d.operationID }
func (d PolicyDecision) Applicability() PolicyApplicability     { return d.applicability }
func (d PolicyDecision) Result() PolicyDecisionResult           { return d.result }
func (d PolicyDecision) EffectClass() PolicyEffectClass         { return d.effectClass }
func (d PolicyDecision) Effects() []PolicyEffect                { return clonePolicyEffects(d.effects) }
func (d PolicyDecision) RequiredInputs() []string               { return cloneStrings(d.requiredInputs) }
func (d PolicyDecision) InputReferences() []string              { return cloneStrings(d.inputReferences) }
func (d PolicyDecision) EvidenceIDs() []EvidenceReferenceID     { return cloneEvidenceIDs(d.evidenceIDs) }
func (d PolicyDecision) ConflictingEvidenceIDs() []EvidenceReferenceID {
	return cloneEvidenceIDs(d.conflictingEvidenceIDs)
}
func (d PolicyDecision) MissingInputs() []string { return cloneStrings(d.missingInputs) }
func (d PolicyDecision) MissingInputBehavior() PolicyMissingInputBehavior {
	return d.missingInputBehavior
}
func (d PolicyDecision) Rationale() string               { return d.rationale }
func (d PolicyDecision) ReasonCode() string              { return d.reasonCode }
func (d PolicyDecision) Priority() int                   { return d.priority }
func (d PolicyDecision) ConflictKey() string             { return d.conflictKey }
func (d PolicyDecision) ExceptionID() *PolicyExceptionID { return clonePolicyExceptionID(d.exceptionID) }
func (d PolicyDecision) CreatedAt() time.Time            { return d.createdAt }

func clonePolicyEffects(values []PolicyEffect) []PolicyEffect {
	if len(values) == 0 {
		return nil
	}
	return append([]PolicyEffect(nil), values...)
}

func validateUniqueStrings(kind string, values []string) error {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if err := requireText(kind, value); err != nil {
			return err
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("%s %q is duplicated", kind, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validateUniqueEvidenceIDs(kind string, values []EvidenceReferenceID) error {
	seen := make(map[EvidenceReferenceID]struct{}, len(values))
	for _, value := range values {
		if err := requireIdentifier(kind, string(value)); err != nil {
			return err
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("%s %q is duplicated", kind, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}
