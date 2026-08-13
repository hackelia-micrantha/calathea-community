package domain

import (
	"fmt"
	"time"
)

const PolicyEvaluatorSemanticVersionV1 = "1"

type PolicyEvaluatorType string

const (
	PolicyEvaluatorCapacityLimit        PolicyEvaluatorType = "capacity_limit"
	PolicyEvaluatorRequiredEvaluation   PolicyEvaluatorType = "required_evaluation"
	PolicyEvaluatorLifecycleEligibility PolicyEvaluatorType = "lifecycle_eligibility"
	PolicyEvaluatorConfidenceGate       PolicyEvaluatorType = "confidence_gate"
	PolicyEvaluatorFreshnessRule        PolicyEvaluatorType = "freshness_rule"
	PolicyEvaluatorScoreMultiplier      PolicyEvaluatorType = "score_multiplier"
)

func (t PolicyEvaluatorType) Valid() bool {
	switch t {
	case PolicyEvaluatorCapacityLimit, PolicyEvaluatorRequiredEvaluation,
		PolicyEvaluatorLifecycleEligibility, PolicyEvaluatorConfidenceGate,
		PolicyEvaluatorFreshnessRule, PolicyEvaluatorScoreMultiplier:
		return true
	default:
		return false
	}
}

type PolicyEffectClass string

const (
	PolicyEffectHard           PolicyEffectClass = "hard"
	PolicyEffectSoft           PolicyEffectClass = "soft"
	PolicyEffectReviewRequired PolicyEffectClass = "review_required"
	PolicyEffectAdvisory       PolicyEffectClass = "advisory"
)

func (c PolicyEffectClass) Valid() bool {
	switch c {
	case PolicyEffectHard, PolicyEffectSoft, PolicyEffectReviewRequired, PolicyEffectAdvisory:
		return true
	default:
		return false
	}
}

type PolicyPhase string

const (
	PolicyPhaseCandidateEligibility  PolicyPhase = "candidate_eligibility"
	PolicyPhaseCandidateAdjustment   PolicyPhase = "candidate_adjustment"
	PolicyPhaseSetConstraints        PolicyPhase = "set_constraints"
	PolicyPhaseResultValidation      PolicyPhase = "result_validation"
	PolicyPhaseDispositionValidation PolicyPhase = "disposition_validation"
	PolicyPhaseReviewDiagnostics     PolicyPhase = "review_diagnostics"
)

func (p PolicyPhase) Valid() bool {
	switch p {
	case PolicyPhaseCandidateEligibility, PolicyPhaseCandidateAdjustment,
		PolicyPhaseSetConstraints, PolicyPhaseResultValidation,
		PolicyPhaseDispositionValidation, PolicyPhaseReviewDiagnostics:
		return true
	default:
		return false
	}
}

func (p PolicyPhase) order() int {
	switch p {
	case PolicyPhaseCandidateEligibility:
		return 1
	case PolicyPhaseCandidateAdjustment:
		return 2
	case PolicyPhaseSetConstraints:
		return 3
	case PolicyPhaseResultValidation:
		return 4
	case PolicyPhaseDispositionValidation:
		return 5
	case PolicyPhaseReviewDiagnostics:
		return 6
	default:
		return 100
	}
}

type PolicyMissingInputBehavior string

const (
	PolicyMissingDeny           PolicyMissingInputBehavior = "deny"
	PolicyMissingExcludeSubject PolicyMissingInputBehavior = "exclude_subject"
	PolicyMissingRequireReview  PolicyMissingInputBehavior = "require_review"
	PolicyMissingFailOperation  PolicyMissingInputBehavior = "fail_operation"
	PolicyMissingDiagnosticOnly PolicyMissingInputBehavior = "diagnostic_only"
)

func (b PolicyMissingInputBehavior) Valid() bool {
	switch b {
	case PolicyMissingDeny, PolicyMissingExcludeSubject, PolicyMissingRequireReview,
		PolicyMissingFailOperation, PolicyMissingDiagnosticOnly:
		return true
	default:
		return false
	}
}

type PolicyExceptionability string

const (
	PolicyNotExceptionable             PolicyExceptionability = "not_exceptionable"
	PolicyExceptionableWithReview      PolicyExceptionability = "exceptionable_with_review"
	PolicyExceptionableWithConstraints PolicyExceptionability = "exceptionable_with_constraints"
)

func (e PolicyExceptionability) Valid() bool {
	switch e {
	case PolicyNotExceptionable, PolicyExceptionableWithReview, PolicyExceptionableWithConstraints:
		return true
	default:
		return false
	}
}

type PolicyEffectKind string

const (
	PolicyEffectCapacityLimit        PolicyEffectKind = "capacity_limit"
	PolicyEffectScoreMultiplier      PolicyEffectKind = "score_multiplier"
	PolicyEffectDiagnosticAnnotation PolicyEffectKind = "diagnostic_annotation"
)

// PolicyEffect is a constrained typed effect. It deliberately has no generic
// map/expression payload: each supported effect has a fixed validated shape.
type PolicyEffect struct {
	kind                  PolicyEffectKind
	placement             Placement
	maximum               uint32
	multiplierBasisPoints uint32
	code                   string
	value                  string
}

func NewCapacityLimitPolicyEffect(placement Placement, maximum int) (PolicyEffect, error) {
	if !placement.Valid() || placement == PlacementKill {
		return PolicyEffect{}, fmt.Errorf("capacity placement %q is invalid", placement)
	}
	if maximum < 0 {
		return PolicyEffect{}, fmt.Errorf("capacity maximum must not be negative")
	}
	if uint64(maximum) > uint64(^uint32(0)) {
		return PolicyEffect{}, fmt.Errorf("capacity maximum %d exceeds supported uint32 range", maximum)
	}
	return PolicyEffect{kind: PolicyEffectCapacityLimit, placement: placement, maximum: uint32(maximum)}, nil
}

func NewScoreMultiplierPolicyEffect(basisPoints int) (PolicyEffect, error) {
	if basisPoints <= 0 || basisPoints > 100000 {
		return PolicyEffect{}, fmt.Errorf("score multiplier basis points %d are outside 1-100000", basisPoints)
	}
	return PolicyEffect{kind: PolicyEffectScoreMultiplier, multiplierBasisPoints: uint32(basisPoints)}, nil
}

func NewDiagnosticPolicyEffect(code, value string) (PolicyEffect, error) {
	if err := requireText("diagnostic code", code); err != nil {
		return PolicyEffect{}, err
	}
	if err := requireText("diagnostic value", value); err != nil {
		return PolicyEffect{}, err
	}
	return PolicyEffect{kind: PolicyEffectDiagnosticAnnotation, code: code, value: value}, nil
}

func (e PolicyEffect) Kind() PolicyEffectKind          { return e.kind }
func (e PolicyEffect) Placement() Placement            { return e.placement }
func (e PolicyEffect) Maximum() uint32                 { return e.maximum }
func (e PolicyEffect) MultiplierBasisPoints() uint32   { return e.multiplierBasisPoints }
func (e PolicyEffect) Code() string                    { return e.code }
func (e PolicyEffect) Value() string                   { return e.value }

type CapacityLimitParameters struct {
	placement Placement
	maximum   uint32
}

func NewCapacityLimitParameters(placement Placement, maximum int) (CapacityLimitParameters, error) {
	effect, err := NewCapacityLimitPolicyEffect(placement, maximum)
	if err != nil {
		return CapacityLimitParameters{}, err
	}
	return CapacityLimitParameters{placement: effect.Placement(), maximum: effect.Maximum()}, nil
}

func (p CapacityLimitParameters) Placement() Placement { return p.placement }
func (p CapacityLimitParameters) Maximum() uint32      { return p.maximum }

type LifecycleEligibilityParameters struct {
	allowProposed bool
}

func NewLifecycleEligibilityParameters(allowProposed bool) LifecycleEligibilityParameters {
	return LifecycleEligibilityParameters{allowProposed: allowProposed}
}

func (p LifecycleEligibilityParameters) AllowProposed() bool { return p.allowProposed }

type ConfidenceGateParameters struct {
	minimumBand ConfidenceBand
}

func NewConfidenceGateParameters(minimumBand ConfidenceBand) (ConfidenceGateParameters, error) {
	if !minimumBand.Valid() {
		return ConfidenceGateParameters{}, fmt.Errorf("invalid minimum confidence band %q", minimumBand)
	}
	return ConfidenceGateParameters{minimumBand: minimumBand}, nil
}

func (p ConfidenceGateParameters) MinimumBand() ConfidenceBand { return p.minimumBand }

type FreshnessRuleParameters struct {
	maximumAge time.Duration
}

func NewFreshnessRuleParameters(maximumAge time.Duration) (FreshnessRuleParameters, error) {
	if maximumAge <= 0 {
		return FreshnessRuleParameters{}, fmt.Errorf("freshness maximum age must be positive")
	}
	return FreshnessRuleParameters{maximumAge: maximumAge}, nil
}

func (p FreshnessRuleParameters) MaximumAge() time.Duration { return p.maximumAge }

type ScoreMultiplierParameters struct {
	basisPoints uint32
}

func NewScoreMultiplierParameters(basisPoints int) (ScoreMultiplierParameters, error) {
	effect, err := NewScoreMultiplierPolicyEffect(basisPoints)
	if err != nil {
		return ScoreMultiplierParameters{}, err
	}
	return ScoreMultiplierParameters{basisPoints: effect.MultiplierBasisPoints()}, nil
}

func (p ScoreMultiplierParameters) BasisPoints() uint32 { return p.basisPoints }

type policyParameters struct {
	capacity   *CapacityLimitParameters
	lifecycle  *LifecycleEligibilityParameters
	confidence *ConfidenceGateParameters
	freshness  *FreshnessRuleParameters
	multiplier *ScoreMultiplierParameters
}

// PolicyInstance is one immutable configured evaluator use within a
// PolicySetVersion. Constructors are evaluator-specific to prevent unsupported
// parameter/effect combinations from entering canonical policy configuration.
type PolicyInstance struct {
	id                   PolicyInstanceID
	policyID             PolicyID
	evaluatorType        PolicyEvaluatorType
	evaluatorVersion     string
	effectClass          PolicyEffectClass
	phase                PolicyPhase
	priority             int
	missingInputBehavior PolicyMissingInputBehavior
	exceptionability     PolicyExceptionability
	enabled              bool
	rationale            string
	parameters           policyParameters
}

func newPolicyInstance(id PolicyInstanceID, policyID PolicyID, evaluatorType PolicyEvaluatorType, effectClass PolicyEffectClass, phase PolicyPhase, priority int, missing PolicyMissingInputBehavior, exceptionability PolicyExceptionability, enabled bool, rationale string, parameters policyParameters) (PolicyInstance, error) {
	if err := requireIdentifier("policy instance id", string(id)); err != nil {
		return PolicyInstance{}, err
	}
	if err := requireIdentifier("policy id", string(policyID)); err != nil {
		return PolicyInstance{}, err
	}
	if !evaluatorType.Valid() {
		return PolicyInstance{}, fmt.Errorf("invalid policy evaluator type %q", evaluatorType)
	}
	if !effectClass.Valid() {
		return PolicyInstance{}, fmt.Errorf("invalid policy effect class %q", effectClass)
	}
	if !phase.Valid() {
		return PolicyInstance{}, fmt.Errorf("invalid policy phase %q", phase)
	}
	if priority < 0 {
		return PolicyInstance{}, fmt.Errorf("policy priority must not be negative")
	}
	if !missing.Valid() {
		return PolicyInstance{}, fmt.Errorf("invalid missing-input behavior %q", missing)
	}
	if missing == PolicyMissingDiagnosticOnly && effectClass != PolicyEffectAdvisory {
		return PolicyInstance{}, fmt.Errorf("diagnostic_only missing-input behavior requires advisory effect class")
	}
	if !exceptionability.Valid() {
		return PolicyInstance{}, fmt.Errorf("invalid policy exceptionability %q", exceptionability)
	}
	if err := requireText("policy rationale", rationale); err != nil {
		return PolicyInstance{}, err
	}
	return PolicyInstance{
		id:                   id,
		policyID:             policyID,
		evaluatorType:        evaluatorType,
		evaluatorVersion:     PolicyEvaluatorSemanticVersionV1,
		effectClass:          effectClass,
		phase:                phase,
		priority:             priority,
		missingInputBehavior: missing,
		exceptionability:     exceptionability,
		enabled:              enabled,
		rationale:            rationale,
		parameters:           parameters,
	}, nil
}

func NewCapacityLimitPolicyInstance(id PolicyInstanceID, policyID PolicyID, parameters CapacityLimitParameters, priority int, rationale string) (PolicyInstance, error) {
	return newPolicyInstance(id, policyID, PolicyEvaluatorCapacityLimit, PolicyEffectHard, PolicyPhaseSetConstraints, priority, PolicyMissingFailOperation, PolicyNotExceptionable, true, rationale, policyParameters{capacity: &parameters})
}

func NewRequiredEvaluationPolicyInstance(id PolicyInstanceID, policyID PolicyID, priority int, missing PolicyMissingInputBehavior, exceptionability PolicyExceptionability, rationale string) (PolicyInstance, error) {
	if missing == PolicyMissingDiagnosticOnly {
		return PolicyInstance{}, fmt.Errorf("required evaluation cannot use diagnostic-only missing behavior")
	}
	return newPolicyInstance(id, policyID, PolicyEvaluatorRequiredEvaluation, PolicyEffectHard, PolicyPhaseCandidateEligibility, priority, missing, exceptionability, true, rationale, policyParameters{})
}

func NewLifecycleEligibilityPolicyInstance(id PolicyInstanceID, policyID PolicyID, parameters LifecycleEligibilityParameters, priority int, rationale string) (PolicyInstance, error) {
	return newPolicyInstance(id, policyID, PolicyEvaluatorLifecycleEligibility, PolicyEffectHard, PolicyPhaseCandidateEligibility, priority, PolicyMissingFailOperation, PolicyNotExceptionable, true, rationale, policyParameters{lifecycle: &parameters})
}

func NewConfidenceGatePolicyInstance(id PolicyInstanceID, policyID PolicyID, parameters ConfidenceGateParameters, priority int, missing PolicyMissingInputBehavior, rationale string) (PolicyInstance, error) {
	return newPolicyInstance(id, policyID, PolicyEvaluatorConfidenceGate, PolicyEffectReviewRequired, PolicyPhaseCandidateAdjustment, priority, missing, PolicyExceptionableWithReview, true, rationale, policyParameters{confidence: &parameters})
}

func NewFreshnessRulePolicyInstance(id PolicyInstanceID, policyID PolicyID, parameters FreshnessRuleParameters, priority int, missing PolicyMissingInputBehavior, rationale string) (PolicyInstance, error) {
	return newPolicyInstance(id, policyID, PolicyEvaluatorFreshnessRule, PolicyEffectReviewRequired, PolicyPhaseCandidateAdjustment, priority, missing, PolicyExceptionableWithReview, true, rationale, policyParameters{freshness: &parameters})
}

func NewScoreMultiplierPolicyInstance(id PolicyInstanceID, policyID PolicyID, parameters ScoreMultiplierParameters, priority int, rationale string) (PolicyInstance, error) {
	return newPolicyInstance(id, policyID, PolicyEvaluatorScoreMultiplier, PolicyEffectSoft, PolicyPhaseCandidateAdjustment, priority, PolicyMissingFailOperation, PolicyNotExceptionable, true, rationale, policyParameters{multiplier: &parameters})
}

func (p PolicyInstance) ID() PolicyInstanceID                               { return p.id }
func (p PolicyInstance) PolicyID() PolicyID                                 { return p.policyID }
func (p PolicyInstance) EvaluatorType() PolicyEvaluatorType                 { return p.evaluatorType }
func (p PolicyInstance) EvaluatorVersion() string                           { return p.evaluatorVersion }
func (p PolicyInstance) EffectClass() PolicyEffectClass                     { return p.effectClass }
func (p PolicyInstance) Phase() PolicyPhase                                 { return p.phase }
func (p PolicyInstance) Priority() int                                      { return p.priority }
func (p PolicyInstance) MissingInputBehavior() PolicyMissingInputBehavior   { return p.missingInputBehavior }
func (p PolicyInstance) Exceptionability() PolicyExceptionability           { return p.exceptionability }
func (p PolicyInstance) Enabled() bool                                      { return p.enabled }
func (p PolicyInstance) Rationale() string                                  { return p.rationale }

func (p PolicyInstance) CapacityLimitParameters() (CapacityLimitParameters, bool) {
	if p.parameters.capacity == nil {
		return CapacityLimitParameters{}, false
	}
	return *p.parameters.capacity, true
}

func (p PolicyInstance) LifecycleEligibilityParameters() (LifecycleEligibilityParameters, bool) {
	if p.parameters.lifecycle == nil {
		return LifecycleEligibilityParameters{}, false
	}
	return *p.parameters.lifecycle, true
}

func (p PolicyInstance) ConfidenceGateParameters() (ConfidenceGateParameters, bool) {
	if p.parameters.confidence == nil {
		return ConfidenceGateParameters{}, false
	}
	return *p.parameters.confidence, true
}

func (p PolicyInstance) FreshnessRuleParameters() (FreshnessRuleParameters, bool) {
	if p.parameters.freshness == nil {
		return FreshnessRuleParameters{}, false
	}
	return *p.parameters.freshness, true
}

func (p PolicyInstance) ScoreMultiplierParameters() (ScoreMultiplierParameters, bool) {
	if p.parameters.multiplier == nil {
		return ScoreMultiplierParameters{}, false
	}
	return *p.parameters.multiplier, true
}

// PolicyException is an immutable maintainer-authorized deviation record. It is
// narrowly scoped to one configured policy instance and one project for v0.
type PolicyException struct {
	id                 PolicyExceptionID
	policySetVersionID PolicySetVersionID
	policyID           PolicyID
	policyInstanceID   PolicyInstanceID
	evaluatorVersion   string
	projectID          ProjectID
	phase              PolicyPhase
	permittedDeviation string
	actor              Actor
	rationale          string
	evidenceIDs        []EvidenceReferenceID
	createdAt          time.Time
	effectiveAt        time.Time
	expiresAt          time.Time
	maximumUses        uint32
	supersedes         *PolicyExceptionID
}

type PolicyExceptionInput struct {
	ID                 PolicyExceptionID
	PolicySetVersionID PolicySetVersionID
	PolicyID           PolicyID
	PolicyInstanceID   PolicyInstanceID
	EvaluatorVersion   string
	ProjectID          ProjectID
	Phase              PolicyPhase
	PermittedDeviation string
	Actor              Actor
	Rationale          string
	EvidenceIDs        []EvidenceReferenceID
	CreatedAt          time.Time
	EffectiveAt        time.Time
	ExpiresAt          time.Time
	MaximumUses        uint32
	Supersedes         *PolicyExceptionID
}

func NewPolicyException(input PolicyExceptionInput) (PolicyException, error) {
	if err := requireIdentifier("policy exception id", string(input.ID)); err != nil {
		return PolicyException{}, err
	}
	if err := requireIdentifier("policy set version id", string(input.PolicySetVersionID)); err != nil {
		return PolicyException{}, err
	}
	if err := requireIdentifier("policy id", string(input.PolicyID)); err != nil {
		return PolicyException{}, err
	}
	if err := requireIdentifier("policy instance id", string(input.PolicyInstanceID)); err != nil {
		return PolicyException{}, err
	}
	if err := requireText("policy evaluator version", input.EvaluatorVersion); err != nil {
		return PolicyException{}, err
	}
	if err := requireIdentifier("project id", string(input.ProjectID)); err != nil {
		return PolicyException{}, err
	}
	if !input.Phase.Valid() {
		return PolicyException{}, fmt.Errorf("invalid policy exception phase %q", input.Phase)
	}
	if err := requireText("permitted policy deviation", input.PermittedDeviation); err != nil {
		return PolicyException{}, err
	}
	if !input.Actor.IsMaintainer() {
		return PolicyException{}, errMaintainerAuthority("policy exception")
	}
	if err := requireText("policy exception rationale", input.Rationale); err != nil {
		return PolicyException{}, err
	}
	if len(input.EvidenceIDs) == 0 {
		return PolicyException{}, fmt.Errorf("policy exception requires supporting evidence/provenance")
	}
	seenEvidence := make(map[EvidenceReferenceID]struct{}, len(input.EvidenceIDs))
	for _, evidenceID := range input.EvidenceIDs {
		if err := requireIdentifier("policy exception evidence reference id", string(evidenceID)); err != nil {
			return PolicyException{}, err
		}
		if _, exists := seenEvidence[evidenceID]; exists {
			return PolicyException{}, fmt.Errorf("policy exception evidence reference id %q is duplicated", evidenceID)
		}
		seenEvidence[evidenceID] = struct{}{}
	}
	if input.CreatedAt.IsZero() || input.EffectiveAt.IsZero() || input.ExpiresAt.IsZero() {
		return PolicyException{}, fmt.Errorf("policy exception times must not be zero")
	}
	if input.EffectiveAt.Before(input.CreatedAt) {
		return PolicyException{}, fmt.Errorf("policy exception effective time must not precede creation")
	}
	if !input.ExpiresAt.After(input.EffectiveAt) {
		return PolicyException{}, fmt.Errorf("policy exception expiry must follow effective time")
	}
	if input.MaximumUses == 0 {
		return PolicyException{}, fmt.Errorf("policy exception maximum uses must be positive")
	}
	if input.Supersedes != nil {
		if err := requireIdentifier("superseded policy exception id", string(*input.Supersedes)); err != nil {
			return PolicyException{}, err
		}
		if *input.Supersedes == input.ID {
			return PolicyException{}, errSelfReference("policy exception supersedes")
		}
	}
	return PolicyException{
		id:                 input.ID,
		policySetVersionID: input.PolicySetVersionID,
		policyID:           input.PolicyID,
		policyInstanceID:   input.PolicyInstanceID,
		evaluatorVersion:   input.EvaluatorVersion,
		projectID:          input.ProjectID,
		phase:              input.Phase,
		permittedDeviation: input.PermittedDeviation,
		actor:              input.Actor,
		rationale:          input.Rationale,
		evidenceIDs:        cloneEvidenceIDs(input.EvidenceIDs),
		createdAt:          input.CreatedAt,
		effectiveAt:        input.EffectiveAt,
		expiresAt:          input.ExpiresAt,
		maximumUses:        input.MaximumUses,
		supersedes:         clonePolicyExceptionID(input.Supersedes),
	}, nil
}

func (e PolicyException) ID() PolicyExceptionID                  { return e.id }
func (e PolicyException) PolicySetVersionID() PolicySetVersionID { return e.policySetVersionID }
func (e PolicyException) PolicyID() PolicyID                     { return e.policyID }
func (e PolicyException) PolicyInstanceID() PolicyInstanceID     { return e.policyInstanceID }
func (e PolicyException) EvaluatorVersion() string               { return e.evaluatorVersion }
func (e PolicyException) ProjectID() ProjectID                   { return e.projectID }
func (e PolicyException) Phase() PolicyPhase                     { return e.phase }
func (e PolicyException) PermittedDeviation() string             { return e.permittedDeviation }
func (e PolicyException) Actor() Actor                           { return e.actor }
func (e PolicyException) Rationale() string                      { return e.rationale }
func (e PolicyException) EvidenceIDs() []EvidenceReferenceID     { return cloneEvidenceIDs(e.evidenceIDs) }
func (e PolicyException) CreatedAt() time.Time                   { return e.createdAt }
func (e PolicyException) EffectiveAt() time.Time                 { return e.effectiveAt }
func (e PolicyException) ExpiresAt() time.Time                   { return e.expiresAt }
func (e PolicyException) MaximumUses() uint32                    { return e.maximumUses }
func (e PolicyException) Supersedes() *PolicyExceptionID         { return clonePolicyExceptionID(e.supersedes) }
func (e PolicyException) EffectiveAtTime(at time.Time) bool {
	return !at.Before(e.effectiveAt) && at.Before(e.expiresAt)
}

type PolicyExceptionRevocation struct {
	id          PolicyExceptionRevocationID
	exceptionID PolicyExceptionID
	actor       Actor
	rationale   string
	revokedAt   time.Time
}

func NewPolicyExceptionRevocation(id PolicyExceptionRevocationID, exceptionID PolicyExceptionID, actor Actor, rationale string, revokedAt time.Time) (PolicyExceptionRevocation, error) {
	if err := requireIdentifier("policy exception revocation id", string(id)); err != nil {
		return PolicyExceptionRevocation{}, err
	}
	if err := requireIdentifier("policy exception id", string(exceptionID)); err != nil {
		return PolicyExceptionRevocation{}, err
	}
	if !actor.IsMaintainer() {
		return PolicyExceptionRevocation{}, errMaintainerAuthority("policy exception revocation")
	}
	if err := requireText("policy exception revocation rationale", rationale); err != nil {
		return PolicyExceptionRevocation{}, err
	}
	if revokedAt.IsZero() {
		return PolicyExceptionRevocation{}, errZeroTime("policy exception revocation time")
	}
	return PolicyExceptionRevocation{id: id, exceptionID: exceptionID, actor: actor, rationale: rationale, revokedAt: revokedAt}, nil
}

func (r PolicyExceptionRevocation) ID() PolicyExceptionRevocationID { return r.id }
func (r PolicyExceptionRevocation) ExceptionID() PolicyExceptionID  { return r.exceptionID }
func (r PolicyExceptionRevocation) Actor() Actor                    { return r.actor }
func (r PolicyExceptionRevocation) Rationale() string               { return r.rationale }
func (r PolicyExceptionRevocation) RevokedAt() time.Time            { return r.revokedAt }

// PolicyExceptionApplication records that an already-authorized exception was
// validly used in one operation. Use-limit/revocation validation occurs before
// this immutable application record is created.
type PolicyExceptionApplication struct {
	id               PolicyExceptionApplicationID
	exceptionID      PolicyExceptionID
	projectID        ProjectID
	operationID      OperationID
	policyDecisionID PolicyDecisionID
	appliedAt        time.Time
}

func NewPolicyExceptionApplication(id PolicyExceptionApplicationID, exception PolicyException, projectID ProjectID, operationID OperationID, policyDecisionID PolicyDecisionID, appliedAt time.Time) (PolicyExceptionApplication, error) {
	if err := requireIdentifier("policy exception application id", string(id)); err != nil {
		return PolicyExceptionApplication{}, err
	}
	if exception.ID() == "" {
		return PolicyExceptionApplication{}, fmt.Errorf("policy exception is required")
	}
	if projectID != exception.ProjectID() {
		return PolicyExceptionApplication{}, fmt.Errorf("policy exception %q does not apply to project %q", exception.ID(), projectID)
	}
	if err := requireIdentifier("operation id", string(operationID)); err != nil {
		return PolicyExceptionApplication{}, err
	}
	if err := requireIdentifier("policy decision id", string(policyDecisionID)); err != nil {
		return PolicyExceptionApplication{}, err
	}
	if appliedAt.IsZero() {
		return PolicyExceptionApplication{}, errZeroTime("policy exception application time")
	}
	if !exception.EffectiveAtTime(appliedAt) {
		return PolicyExceptionApplication{}, fmt.Errorf("policy exception %q is not effective at application time", exception.ID())
	}
	return PolicyExceptionApplication{id: id, exceptionID: exception.ID(), projectID: projectID, operationID: operationID, policyDecisionID: policyDecisionID, appliedAt: appliedAt}, nil
}

func (a PolicyExceptionApplication) ID() PolicyExceptionApplicationID             { return a.id }
func (a PolicyExceptionApplication) ExceptionID() PolicyExceptionID               { return a.exceptionID }
func (a PolicyExceptionApplication) ProjectID() ProjectID                         { return a.projectID }
func (a PolicyExceptionApplication) OperationID() OperationID                     { return a.operationID }
func (a PolicyExceptionApplication) PolicyDecisionID() PolicyDecisionID           { return a.policyDecisionID }
func (a PolicyExceptionApplication) AppliedAt() time.Time                         { return a.appliedAt }

func clonePolicyExceptionID(value *PolicyExceptionID) *PolicyExceptionID {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}
