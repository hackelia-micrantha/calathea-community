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

// PolicyDecision is the immutable deterministic result of applying one policy
// instance to one project in one operation. It captures evaluator identity,
// effect class/phase, typed effects, exact deterministic inputs, missing-input
// behavior, and evidence references so replay does not depend on ambient state.
type PolicyDecision struct {
	id                   PolicyDecisionID
	policySetVersionID   PolicySetVersionID
	policyID             PolicyID
	policyInstanceID     PolicyInstanceID
	evaluatorType        PolicyEvaluatorType
	evaluatorVersion     string
	projectID            ProjectID
	operationID          OperationID
	result               PolicyDecisionResult
	effectClass          PolicyEffectClass
	phase                PolicyPhase
	effects              []PolicyEffect
	reasonCode           string
	inputReferences      []string
	evidenceIDs          []EvidenceReferenceID
	missingInputs        []string
	missingInputBehavior PolicyMissingInputBehavior
	priority             int
	exceptionID          *PolicyExceptionID
	createdAt            time.Time
}

type PolicyDecisionInput struct {
	ID                   PolicyDecisionID
	PolicySetVersionID   PolicySetVersionID
	PolicyID             PolicyID
	PolicyInstanceID     PolicyInstanceID
	EvaluatorType        PolicyEvaluatorType
	EvaluatorVersion     string
	ProjectID            ProjectID
	OperationID          OperationID
	Result               PolicyDecisionResult
	EffectClass          PolicyEffectClass
	Phase                PolicyPhase
	Effects              []PolicyEffect
	ReasonCode           string
	InputReferences      []string
	EvidenceIDs          []EvidenceReferenceID
	MissingInputs        []string
	MissingInputBehavior PolicyMissingInputBehavior
	Priority             int
	ExceptionID          *PolicyExceptionID
	CreatedAt            time.Time
}

func NewPolicyDecision(input PolicyDecisionInput) (PolicyDecision, error) {
	if err := requireIdentifier("policy decision id", string(input.ID)); err != nil {
		return PolicyDecision{}, err
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
	if err := requireIdentifier("project id", string(input.ProjectID)); err != nil {
		return PolicyDecision{}, err
	}
	if err := requireIdentifier("operation id", string(input.OperationID)); err != nil {
		return PolicyDecision{}, err
	}
	if !input.Result.Valid() {
		return PolicyDecision{}, fmt.Errorf("invalid policy decision result %q", input.Result)
	}
	if !input.EffectClass.Valid() {
		return PolicyDecision{}, fmt.Errorf("invalid policy effect class %q", input.EffectClass)
	}
	if !input.Phase.Valid() {
		return PolicyDecision{}, fmt.Errorf("invalid policy phase %q", input.Phase)
	}
	if input.Priority < 0 {
		return PolicyDecision{}, fmt.Errorf("policy decision priority must not be negative")
	}
	if !input.MissingInputBehavior.Valid() {
		return PolicyDecision{}, fmt.Errorf("invalid missing-input behavior %q", input.MissingInputBehavior)
	}
	if err := requireText("policy reason code", input.ReasonCode); err != nil {
		return PolicyDecision{}, err
	}
	if input.CreatedAt.IsZero() {
		return PolicyDecision{}, errZeroTime("policy decision time")
	}
	for _, reference := range input.InputReferences {
		if err := requireText("policy decision input reference", reference); err != nil {
			return PolicyDecision{}, err
		}
	}
	for _, evidenceID := range input.EvidenceIDs {
		if err := requireIdentifier("policy decision evidence reference id", string(evidenceID)); err != nil {
			return PolicyDecision{}, err
		}
	}
	for _, missing := range input.MissingInputs {
		if err := requireText("policy missing input", missing); err != nil {
			return PolicyDecision{}, err
		}
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
	if input.ExceptionID != nil {
		if err := requireIdentifier("policy exception id", string(*input.ExceptionID)); err != nil {
			return PolicyDecision{}, err
		}
	}
	return PolicyDecision{
		id:                   input.ID,
		policySetVersionID:   input.PolicySetVersionID,
		policyID:             input.PolicyID,
		policyInstanceID:     input.PolicyInstanceID,
		evaluatorType:        input.EvaluatorType,
		evaluatorVersion:     input.EvaluatorVersion,
		projectID:            input.ProjectID,
		operationID:          input.OperationID,
		result:               input.Result,
		effectClass:          input.EffectClass,
		phase:                input.Phase,
		effects:              clonePolicyEffects(input.Effects),
		reasonCode:           input.ReasonCode,
		inputReferences:      cloneStrings(input.InputReferences),
		evidenceIDs:          cloneEvidenceIDs(input.EvidenceIDs),
		missingInputs:        cloneStrings(input.MissingInputs),
		missingInputBehavior: input.MissingInputBehavior,
		priority:             input.Priority,
		exceptionID:          clonePolicyExceptionID(input.ExceptionID),
		createdAt:            input.CreatedAt,
	}, nil
}

func (d PolicyDecision) ID() PolicyDecisionID                   { return d.id }
func (d PolicyDecision) PolicySetVersionID() PolicySetVersionID { return d.policySetVersionID }
func (d PolicyDecision) PolicyID() PolicyID                     { return d.policyID }
func (d PolicyDecision) PolicyInstanceID() PolicyInstanceID     { return d.policyInstanceID }
func (d PolicyDecision) EvaluatorType() PolicyEvaluatorType     { return d.evaluatorType }
func (d PolicyDecision) EvaluatorVersion() string               { return d.evaluatorVersion }
func (d PolicyDecision) ProjectID() ProjectID                   { return d.projectID }
func (d PolicyDecision) OperationID() OperationID               { return d.operationID }
func (d PolicyDecision) Result() PolicyDecisionResult           { return d.result }
func (d PolicyDecision) EffectClass() PolicyEffectClass         { return d.effectClass }
func (d PolicyDecision) Phase() PolicyPhase                     { return d.phase }
func (d PolicyDecision) Effects() []PolicyEffect                { return clonePolicyEffects(d.effects) }
func (d PolicyDecision) ReasonCode() string                     { return d.reasonCode }
func (d PolicyDecision) InputReferences() []string              { return cloneStrings(d.inputReferences) }
func (d PolicyDecision) EvidenceIDs() []EvidenceReferenceID     { return cloneEvidenceIDs(d.evidenceIDs) }
func (d PolicyDecision) MissingInputs() []string                { return cloneStrings(d.missingInputs) }
func (d PolicyDecision) MissingInputBehavior() PolicyMissingInputBehavior { return d.missingInputBehavior }
func (d PolicyDecision) Priority() int                          { return d.priority }
func (d PolicyDecision) ExceptionID() *PolicyExceptionID        { return clonePolicyExceptionID(d.exceptionID) }
func (d PolicyDecision) CreatedAt() time.Time                   { return d.createdAt }

func clonePolicyEffects(values []PolicyEffect) []PolicyEffect {
	if len(values) == 0 {
		return nil
	}
	return append([]PolicyEffect(nil), values...)
}
