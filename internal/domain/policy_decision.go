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

// PolicyInputReference retains the exact deterministic value/reference consumed
// for one declared policy input kind. Values are intentionally opaque to the
// durable decision record; the evaluator version owns their interpretation.
type PolicyInputReference struct {
	kind  PolicyInputKind
	value string
}

func NewPolicyInputReference(kind PolicyInputKind, value string) (PolicyInputReference, error) {
	if !kind.Valid() {
		return PolicyInputReference{}, fmt.Errorf("invalid policy input kind %q", kind)
	}
	if err := requireText("policy input reference value", value); err != nil {
		return PolicyInputReference{}, err
	}
	return PolicyInputReference{kind: kind, value: value}, nil
}

func (r PolicyInputReference) Kind() PolicyInputKind { return r.kind }
func (r PolicyInputReference) Value() string         { return r.value }

// PolicyDecision is the immutable deterministic result of applying one exact
// configured policy instance to one typed subject in an operation. It retains
// enough evaluator/configuration/input identity for later composition and replay
// without consulting mutable policy defaults.
type PolicyDecision struct {
	id                   PolicyDecisionID
	policySetVersionID   PolicySetVersionID
	policyID             PolicyID
	policyInstanceID     PolicyInstanceID
	evaluatorType        PolicyEvaluatorType
	evaluatorVersion     string
	phase                PolicyPhase
	effectClass          PolicyEffectClass
	missingInputBehavior PolicyMissingInputBehavior
	subject              PolicySubject
	operationID          OperationID
	result               PolicyDecisionResult
	reasonCode           string
	requiredInputs       []PolicyInputKind
	inputReferences      []PolicyInputReference
	missingInputs        []PolicyInputKind
	evidenceIDs          []EvidenceReferenceID
	effects              []PolicyEffect
	priority             int
	rationale            string
	createdAt            time.Time
}

type PolicyDecisionInput struct {
	ID                   PolicyDecisionID
	PolicySetVersionID   PolicySetVersionID
	PolicyID             PolicyID
	PolicyInstanceID     PolicyInstanceID
	EvaluatorType        PolicyEvaluatorType
	EvaluatorVersion     string
	Phase                PolicyPhase
	EffectClass          PolicyEffectClass
	MissingInputBehavior PolicyMissingInputBehavior
	Subject              PolicySubject
	OperationID          OperationID
	Result               PolicyDecisionResult
	ReasonCode           string
	RequiredInputs       []PolicyInputKind
	InputReferences      []PolicyInputReference
	MissingInputs        []PolicyInputKind
	EvidenceIDs          []EvidenceReferenceID
	Effects              []PolicyEffect
	Priority             int
	Rationale            string
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
	if err := requireIdentifier("policy evaluator version", input.EvaluatorVersion); err != nil {
		return PolicyDecision{}, err
	}
	if !input.Phase.Valid() {
		return PolicyDecision{}, fmt.Errorf("invalid policy phase %q", input.Phase)
	}
	if !input.EffectClass.Valid() {
		return PolicyDecision{}, fmt.Errorf("invalid policy effect class %q", input.EffectClass)
	}
	if !input.MissingInputBehavior.Valid() {
		return PolicyDecision{}, fmt.Errorf("invalid missing-input behavior %q", input.MissingInputBehavior)
	}
	if !input.Subject.valid() {
		return PolicyDecision{}, fmt.Errorf("invalid policy subject")
	}
	if err := requireIdentifier("operation id", string(input.OperationID)); err != nil {
		return PolicyDecision{}, err
	}
	if !input.Result.Valid() {
		return PolicyDecision{}, fmt.Errorf("invalid policy decision result %q", input.Result)
	}
	if err := requireText("policy reason code", input.ReasonCode); err != nil {
		return PolicyDecision{}, err
	}
	if input.Priority < 0 {
		return PolicyDecision{}, fmt.Errorf("policy decision priority %d must not be negative", input.Priority)
	}
	if err := requireText("policy rationale", input.Rationale); err != nil {
		return PolicyDecision{}, err
	}
	if input.CreatedAt.IsZero() {
		return PolicyDecision{}, errZeroTime("policy decision time")
	}
	if err := validatePolicyInputKinds("required policy input", input.RequiredInputs); err != nil {
		return PolicyDecision{}, err
	}
	if err := validatePolicyInputKinds("missing policy input", input.MissingInputs); err != nil {
		return PolicyDecision{}, err
	}

	required := make(map[PolicyInputKind]struct{}, len(input.RequiredInputs))
	for _, kind := range input.RequiredInputs {
		required[kind] = struct{}{}
	}
	missing := make(map[PolicyInputKind]struct{}, len(input.MissingInputs))
	for _, kind := range input.MissingInputs {
		if _, ok := required[kind]; !ok {
			return PolicyDecision{}, fmt.Errorf("missing policy input %q was not declared required", kind)
		}
		missing[kind] = struct{}{}
	}

	referenced := make(map[PolicyInputKind]struct{}, len(input.InputReferences))
	for _, reference := range input.InputReferences {
		if !reference.kind.Valid() {
			return PolicyDecision{}, fmt.Errorf("invalid policy input reference kind %q", reference.kind)
		}
		if err := requireText("policy input reference value", reference.value); err != nil {
			return PolicyDecision{}, err
		}
		if _, ok := required[reference.kind]; !ok {
			return PolicyDecision{}, fmt.Errorf("policy input reference %q was not declared required", reference.kind)
		}
		if _, exists := referenced[reference.kind]; exists {
			return PolicyDecision{}, fmt.Errorf("policy input reference %q is duplicated", reference.kind)
		}
		if _, isMissing := missing[reference.kind]; isMissing {
			return PolicyDecision{}, fmt.Errorf("policy input %q cannot be both referenced and missing", reference.kind)
		}
		referenced[reference.kind] = struct{}{}
	}

	if input.Result == PolicyDecisionIndeterminate && len(input.MissingInputs) == 0 {
		return PolicyDecision{}, fmt.Errorf("indeterminate policy decision requires at least one missing input")
	}
	if input.Result != PolicyDecisionIndeterminate && len(input.MissingInputs) != 0 {
		return PolicyDecision{}, fmt.Errorf("non-indeterminate policy decision must not report missing inputs")
	}
	if input.Result != PolicyDecisionNotApplicable {
		for _, kind := range input.RequiredInputs {
			_, hasReference := referenced[kind]
			_, isMissing := missing[kind]
			if !hasReference && !isMissing {
				return PolicyDecision{}, fmt.Errorf("required policy input %q lacks a retained reference or missing marker", kind)
			}
		}
	}
	if err := validateEvidenceIDs(input.EvidenceIDs); err != nil {
		return PolicyDecision{}, err
	}
	for _, effect := range input.Effects {
		if !effect.valid() {
			return PolicyDecision{}, fmt.Errorf("invalid policy effect")
		}
	}
	if input.Result == PolicyDecisionDeny && input.EffectClass != PolicyEffectHard {
		return PolicyDecision{}, fmt.Errorf("deny policy decision requires hard effect class")
	}
	if input.Result == PolicyDecisionRequireReview && input.EffectClass != PolicyEffectReviewRequired {
		return PolicyDecision{}, fmt.Errorf("require_review decision requires review_required effect class")
	}
	if input.Result == PolicyDecisionAdjust && len(input.Effects) == 0 {
		return PolicyDecision{}, fmt.Errorf("adjust policy decision requires a typed effect")
	}

	return PolicyDecision{
		id:                   input.ID,
		policySetVersionID:   input.PolicySetVersionID,
		policyID:             input.PolicyID,
		policyInstanceID:     input.PolicyInstanceID,
		evaluatorType:        input.EvaluatorType,
		evaluatorVersion:     input.EvaluatorVersion,
		phase:                input.Phase,
		effectClass:          input.EffectClass,
		missingInputBehavior: input.MissingInputBehavior,
		subject:              input.Subject,
		operationID:          input.OperationID,
		result:               input.Result,
		reasonCode:           input.ReasonCode,
		requiredInputs:       clonePolicyInputKinds(input.RequiredInputs),
		inputReferences:      clonePolicyInputReferences(input.InputReferences),
		missingInputs:        clonePolicyInputKinds(input.MissingInputs),
		evidenceIDs:          cloneEvidenceIDs(input.EvidenceIDs),
		effects:              clonePolicyEffects(input.Effects),
		priority:             input.Priority,
		rationale:            input.Rationale,
		createdAt:            input.CreatedAt,
	}, nil
}

func (d PolicyDecision) ID() PolicyDecisionID                   { return d.id }
func (d PolicyDecision) PolicySetVersionID() PolicySetVersionID { return d.policySetVersionID }
func (d PolicyDecision) PolicyID() PolicyID                     { return d.policyID }
func (d PolicyDecision) PolicyInstanceID() PolicyInstanceID     { return d.policyInstanceID }
func (d PolicyDecision) EvaluatorType() PolicyEvaluatorType     { return d.evaluatorType }
func (d PolicyDecision) EvaluatorVersion() string               { return d.evaluatorVersion }
func (d PolicyDecision) Phase() PolicyPhase                     { return d.phase }
func (d PolicyDecision) EffectClass() PolicyEffectClass         { return d.effectClass }
func (d PolicyDecision) MissingInputBehavior() PolicyMissingInputBehavior {
	return d.missingInputBehavior
}
func (d PolicyDecision) Subject() PolicySubject { return d.subject }
func (d PolicyDecision) ProjectID() ProjectID {
	if d.subject.Type() != PolicySubjectProject {
		return ""
	}
	return ProjectID(d.subject.ID())
}
func (d PolicyDecision) OperationID() OperationID                  { return d.operationID }
func (d PolicyDecision) Result() PolicyDecisionResult              { return d.result }
func (d PolicyDecision) ReasonCode() string                        { return d.reasonCode }
func (d PolicyDecision) RequiredInputs() []PolicyInputKind         { return clonePolicyInputKinds(d.requiredInputs) }
func (d PolicyDecision) InputReferences() []PolicyInputReference   { return clonePolicyInputReferences(d.inputReferences) }
func (d PolicyDecision) MissingInputs() []PolicyInputKind          { return clonePolicyInputKinds(d.missingInputs) }
func (d PolicyDecision) EvidenceIDs() []EvidenceReferenceID        { return cloneEvidenceIDs(d.evidenceIDs) }
func (d PolicyDecision) Effects() []PolicyEffect                   { return clonePolicyEffects(d.effects) }
func (d PolicyDecision) Priority() int                             { return d.priority }
func (d PolicyDecision) Rationale() string                         { return d.rationale }
func (d PolicyDecision) CreatedAt() time.Time                      { return d.createdAt }

func validatePolicyInputKinds(kind string, values []PolicyInputKind) error {
	seen := make(map[PolicyInputKind]struct{}, len(values))
	for _, value := range values {
		if !value.Valid() {
			return fmt.Errorf("invalid %s %q", kind, value)
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("%s %q is duplicated", kind, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validateEvidenceIDs(values []EvidenceReferenceID) error {
	seen := make(map[EvidenceReferenceID]struct{}, len(values))
	for _, value := range values {
		if err := requireIdentifier("policy decision evidence reference id", string(value)); err != nil {
			return err
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("policy decision evidence reference id %q is duplicated", value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func clonePolicyInputReferences(values []PolicyInputReference) []PolicyInputReference {
	if len(values) == 0 {
		return nil
	}
	return append([]PolicyInputReference(nil), values...)
}
