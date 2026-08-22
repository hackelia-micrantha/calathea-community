package domain

import "fmt"

// PolicyInstanceID identifies one configured policy instance within a PolicySetVersion.
type PolicyInstanceID string

// PolicyEvaluatorType identifies a built-in deterministic evaluator family.
type PolicyEvaluatorType string

const (
	PolicyEvaluatorLifecycleEligibility PolicyEvaluatorType = "lifecycle_eligibility"
	PolicyEvaluatorRequiredEvaluation   PolicyEvaluatorType = "required_evaluation"
	PolicyEvaluatorCapacityLimit        PolicyEvaluatorType = "capacity_limit"
	PolicyEvaluatorConfidenceGate       PolicyEvaluatorType = "confidence_gate"
	PolicyEvaluatorFreshnessRule        PolicyEvaluatorType = "freshness_rule"
)

func (t PolicyEvaluatorType) Valid() bool {
	switch t {
	case PolicyEvaluatorLifecycleEligibility,
		PolicyEvaluatorRequiredEvaluation,
		PolicyEvaluatorCapacityLimit,
		PolicyEvaluatorConfidenceGate,
		PolicyEvaluatorFreshnessRule:
		return true
	default:
		return false
	}
}

// PolicyPhase identifies when an evaluator participates in an orientation workflow.
type PolicyPhase string

const (
	PolicyPhaseCandidateEligibility PolicyPhase = "candidate_eligibility"
	PolicyPhaseCandidateAdjustment  PolicyPhase = "candidate_adjustment"
	PolicyPhaseSetConstraints       PolicyPhase = "set_constraints"
)

func (p PolicyPhase) Valid() bool {
	switch p {
	case PolicyPhaseCandidateEligibility, PolicyPhaseCandidateAdjustment, PolicyPhaseSetConstraints:
		return true
	default:
		return false
	}
}

// PolicyEffectClass distinguishes legal constraints from bounded preferences and diagnostics.
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

// PolicyMissingInputBehavior describes the configured response to an applicable
// policy whose required input is unavailable. The policy engine retains the
// indeterminate decision; composition interprets this behavior in a later phase.
type PolicyMissingInputBehavior string

const (
	PolicyMissingInputDeny            PolicyMissingInputBehavior = "deny"
	PolicyMissingInputExcludeSubject  PolicyMissingInputBehavior = "exclude_subject"
	PolicyMissingInputRequireReview   PolicyMissingInputBehavior = "require_review"
	PolicyMissingInputFailOperation   PolicyMissingInputBehavior = "fail_operation"
	PolicyMissingInputDiagnosticOnly  PolicyMissingInputBehavior = "diagnostic_only"
)

func (b PolicyMissingInputBehavior) Valid() bool {
	switch b {
	case PolicyMissingInputDeny,
		PolicyMissingInputExcludeSubject,
		PolicyMissingInputRequireReview,
		PolicyMissingInputFailOperation,
		PolicyMissingInputDiagnosticOnly:
		return true
	default:
		return false
	}
}

// PolicyExceptionability records whether a policy may ever be deviated from by
// a separately authorized PolicyException.
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

// PolicySubjectType identifies the deterministic subject presented to an evaluator.
type PolicySubjectType string

const (
	PolicySubjectProject      PolicySubjectType = "project"
	PolicySubjectPlacementSet PolicySubjectType = "placement_set"
)

func (t PolicySubjectType) Valid() bool {
	return t == PolicySubjectProject || t == PolicySubjectPlacementSet
}

// PolicyInputKind names an input required by an evaluator without coupling the
// domain contract to an application persistence representation.
type PolicyInputKind string

const (
	PolicyInputLifecycleState    PolicyInputKind = "lifecycle_state"
	PolicyInputAcceptedEvaluation PolicyInputKind = "accepted_evaluation"
	PolicyInputSelectedCount     PolicyInputKind = "selected_count"
	PolicyInputAsOfTime          PolicyInputKind = "as_of_time"
)

func (k PolicyInputKind) Valid() bool {
	switch k {
	case PolicyInputLifecycleState, PolicyInputAcceptedEvaluation, PolicyInputSelectedCount, PolicyInputAsOfTime:
		return true
	default:
		return false
	}
}

// PolicyParameters is the closed v0 parameter surface for built-in evaluators.
// New evaluator parameters require an explicit domain/schema change rather than
// an untyped map or executable expression.
type PolicyParameters struct {
	placement                  *Placement
	maximum                    *int
	allowProposed              *bool
	confidenceReviewBelowBPS   *int
	freshnessMaxAgeDays        *int
}

func NewNoPolicyParameters() PolicyParameters { return PolicyParameters{} }

func NewLifecycleEligibilityParameters(allowProposed bool) PolicyParameters {
	return PolicyParameters{allowProposed: cloneBool(&allowProposed)}
}

func NewCapacityLimitParameters(placement Placement, maximum int) (PolicyParameters, error) {
	if !placement.Valid() || placement == PlacementKill {
		return PolicyParameters{}, fmt.Errorf("capacity policy requires now, next, or later placement, got %q", placement)
	}
	if maximum < 0 {
		return PolicyParameters{}, fmt.Errorf("capacity maximum %d must not be negative", maximum)
	}
	return PolicyParameters{placement: clonePlacement(&placement), maximum: cloneInt(&maximum)}, nil
}

func NewConfidenceGateParameters(reviewBelowBasisPoints int) (PolicyParameters, error) {
	if reviewBelowBasisPoints < 0 || reviewBelowBasisPoints > 10000 {
		return PolicyParameters{}, fmt.Errorf("confidence review threshold %d is outside 0-10000", reviewBelowBasisPoints)
	}
	return PolicyParameters{confidenceReviewBelowBPS: cloneInt(&reviewBelowBasisPoints)}, nil
}

func NewFreshnessRuleParameters(maxAgeDays int) (PolicyParameters, error) {
	if maxAgeDays < 0 {
		return PolicyParameters{}, fmt.Errorf("freshness maximum age %d must not be negative", maxAgeDays)
	}
	return PolicyParameters{freshnessMaxAgeDays: cloneInt(&maxAgeDays)}, nil
}

func (p PolicyParameters) Placement() (Placement, bool) {
	if p.placement == nil {
		return "", false
	}
	return *p.placement, true
}

func (p PolicyParameters) Maximum() (int, bool) {
	if p.maximum == nil {
		return 0, false
	}
	return *p.maximum, true
}

func (p PolicyParameters) AllowProposed() (bool, bool) {
	if p.allowProposed == nil {
		return false, false
	}
	return *p.allowProposed, true
}

func (p PolicyParameters) ConfidenceReviewBelowBasisPoints() (int, bool) {
	if p.confidenceReviewBelowBPS == nil {
		return 0, false
	}
	return *p.confidenceReviewBelowBPS, true
}

func (p PolicyParameters) FreshnessMaxAgeDays() (int, bool) {
	if p.freshnessMaxAgeDays == nil {
		return 0, false
	}
	return *p.freshnessMaxAgeDays, true
}

func (p PolicyParameters) clone() PolicyParameters {
	return PolicyParameters{
		placement:                clonePlacement(p.placement),
		maximum:                  cloneInt(p.maximum),
		allowProposed:            cloneBool(p.allowProposed),
		confidenceReviewBelowBPS: cloneInt(p.confidenceReviewBelowBPS),
		freshnessMaxAgeDays:      cloneInt(p.freshnessMaxAgeDays),
	}
}

// PolicyInstanceInput contains the complete immutable configuration for one
// built-in policy evaluator.
type PolicyInstanceInput struct {
	ID                   PolicyInstanceID
	PolicyID             PolicyID
	EvaluatorType        PolicyEvaluatorType
	EvaluatorVersion     string
	Phase                PolicyPhase
	EffectClass          PolicyEffectClass
	SubjectType          PolicySubjectType
	RequiredInputs       []PolicyInputKind
	MissingInputBehavior PolicyMissingInputBehavior
	Priority             int
	Exceptionability     PolicyExceptionability
	Parameters           PolicyParameters
	Rationale            string
}

// PolicyInstance is one immutable configured evaluator within a PolicySetVersion.
type PolicyInstance struct {
	id                   PolicyInstanceID
	policyID             PolicyID
	evaluatorType        PolicyEvaluatorType
	evaluatorVersion     string
	phase                PolicyPhase
	effectClass          PolicyEffectClass
	subjectType          PolicySubjectType
	requiredInputs       []PolicyInputKind
	missingInputBehavior PolicyMissingInputBehavior
	priority             int
	exceptionability     PolicyExceptionability
	parameters           PolicyParameters
	rationale            string
}

func NewPolicyInstance(input PolicyInstanceInput) (PolicyInstance, error) {
	instance := PolicyInstance{
		id:                   input.ID,
		policyID:             input.PolicyID,
		evaluatorType:        input.EvaluatorType,
		evaluatorVersion:     input.EvaluatorVersion,
		phase:                input.Phase,
		effectClass:          input.EffectClass,
		subjectType:          input.SubjectType,
		requiredInputs:       clonePolicyInputKinds(input.RequiredInputs),
		missingInputBehavior: input.MissingInputBehavior,
		priority:             input.Priority,
		exceptionability:     input.Exceptionability,
		parameters:           input.Parameters.clone(),
		rationale:            input.Rationale,
	}
	if err := validatePolicyInstance(instance); err != nil {
		return PolicyInstance{}, err
	}
	return instance, nil
}

func validatePolicyInstance(instance PolicyInstance) error {
	if err := requireIdentifier("policy instance id", string(instance.id)); err != nil {
		return err
	}
	if err := requireIdentifier("policy id", string(instance.policyID)); err != nil {
		return err
	}
	if !instance.evaluatorType.Valid() {
		return fmt.Errorf("unsupported policy evaluator type %q", instance.evaluatorType)
	}
	if err := requireIdentifier("policy evaluator version", instance.evaluatorVersion); err != nil {
		return err
	}
	if !instance.phase.Valid() {
		return fmt.Errorf("invalid policy phase %q", instance.phase)
	}
	if !instance.effectClass.Valid() {
		return fmt.Errorf("invalid policy effect class %q", instance.effectClass)
	}
	if !instance.subjectType.Valid() {
		return fmt.Errorf("invalid policy subject type %q", instance.subjectType)
	}
	if !instance.missingInputBehavior.Valid() {
		return fmt.Errorf("invalid missing-input behavior %q", instance.missingInputBehavior)
	}
	if !instance.exceptionability.Valid() {
		return fmt.Errorf("invalid policy exceptionability %q", instance.exceptionability)
	}
	if instance.priority < 0 {
		return fmt.Errorf("policy priority %d must not be negative", instance.priority)
	}
	if err := requireText("policy rationale", instance.rationale); err != nil {
		return err
	}
	seenInputs := make(map[PolicyInputKind]struct{}, len(instance.requiredInputs))
	for _, kind := range instance.requiredInputs {
		if !kind.Valid() {
			return fmt.Errorf("invalid required policy input %q", kind)
		}
		if _, exists := seenInputs[kind]; exists {
			return fmt.Errorf("required policy input %q is duplicated", kind)
		}
		seenInputs[kind] = struct{}{}
	}
	return validatePolicyParameters(instance.evaluatorType, instance.parameters)
}

func validatePolicyParameters(evaluator PolicyEvaluatorType, parameters PolicyParameters) error {
	placement, hasPlacement := parameters.Placement()
	maximum, hasMaximum := parameters.Maximum()
	_, hasAllowProposed := parameters.AllowProposed()
	confidenceThreshold, hasConfidence := parameters.ConfidenceReviewBelowBasisPoints()
	freshnessDays, hasFreshness := parameters.FreshnessMaxAgeDays()

	switch evaluator {
	case PolicyEvaluatorLifecycleEligibility:
		if !hasAllowProposed || hasPlacement || hasMaximum || hasConfidence || hasFreshness {
			return fmt.Errorf("lifecycle eligibility policy has invalid parameter shape")
		}
	case PolicyEvaluatorRequiredEvaluation:
		if hasAllowProposed || hasPlacement || hasMaximum || hasConfidence || hasFreshness {
			return fmt.Errorf("required evaluation policy does not accept parameters")
		}
	case PolicyEvaluatorCapacityLimit:
		if !hasPlacement || !hasMaximum || hasAllowProposed || hasConfidence || hasFreshness {
			return fmt.Errorf("capacity policy requires only placement and maximum parameters")
		}
		if !placement.Valid() || placement == PlacementKill || maximum < 0 {
			return fmt.Errorf("capacity policy parameters are invalid")
		}
	case PolicyEvaluatorConfidenceGate:
		if !hasConfidence || hasAllowProposed || hasPlacement || hasMaximum || hasFreshness {
			return fmt.Errorf("confidence gate requires only a confidence threshold")
		}
		if confidenceThreshold < 0 || confidenceThreshold > 10000 {
			return fmt.Errorf("confidence review threshold %d is outside 0-10000", confidenceThreshold)
		}
	case PolicyEvaluatorFreshnessRule:
		if !hasFreshness || hasAllowProposed || hasPlacement || hasMaximum || hasConfidence {
			return fmt.Errorf("freshness rule requires only maximum age days")
		}
		if freshnessDays < 0 {
			return fmt.Errorf("freshness maximum age %d must not be negative", freshnessDays)
		}
	default:
		return fmt.Errorf("unsupported policy evaluator type %q", evaluator)
	}
	return nil
}

func (p PolicyInstance) ID() PolicyInstanceID                         { return p.id }
func (p PolicyInstance) PolicyID() PolicyID                           { return p.policyID }
func (p PolicyInstance) EvaluatorType() PolicyEvaluatorType           { return p.evaluatorType }
func (p PolicyInstance) EvaluatorVersion() string                     { return p.evaluatorVersion }
func (p PolicyInstance) Phase() PolicyPhase                           { return p.phase }
func (p PolicyInstance) EffectClass() PolicyEffectClass               { return p.effectClass }
func (p PolicyInstance) SubjectType() PolicySubjectType               { return p.subjectType }
func (p PolicyInstance) RequiredInputs() []PolicyInputKind             { return clonePolicyInputKinds(p.requiredInputs) }
func (p PolicyInstance) MissingInputBehavior() PolicyMissingInputBehavior { return p.missingInputBehavior }
func (p PolicyInstance) Priority() int                                { return p.priority }
func (p PolicyInstance) Exceptionability() PolicyExceptionability     { return p.exceptionability }
func (p PolicyInstance) Parameters() PolicyParameters                 { return p.parameters.clone() }
func (p PolicyInstance) Rationale() string                            { return p.rationale }

func clonePolicyInputKinds(values []PolicyInputKind) []PolicyInputKind {
	if len(values) == 0 {
		return nil
	}
	return append([]PolicyInputKind(nil), values...)
}

// PolicySubject is a typed deterministic evaluator subject. Capacity policies
// target a placement set rather than pretending that every policy is project-scoped.
type PolicySubject struct {
	typeName PolicySubjectType
	id       string
}

func NewProjectPolicySubject(projectID ProjectID) (PolicySubject, error) {
	if err := requireIdentifier("project id", string(projectID)); err != nil {
		return PolicySubject{}, err
	}
	return PolicySubject{typeName: PolicySubjectProject, id: string(projectID)}, nil
}

func NewPlacementSetPolicySubject(placement Placement) (PolicySubject, error) {
	if !placement.Valid() || placement == PlacementKill {
		return PolicySubject{}, fmt.Errorf("placement-set subject requires now, next, or later placement, got %q", placement)
	}
	return PolicySubject{typeName: PolicySubjectPlacementSet, id: string(placement)}, nil
}

func (s PolicySubject) Type() PolicySubjectType { return s.typeName }
func (s PolicySubject) ID() string              { return s.id }
func (s PolicySubject) valid() bool {
	return s.typeName.Valid() && requireIdentifier("policy subject id", s.id) == nil
}

// PolicyEffectType identifies a bounded structured effect produced by an evaluator.
type PolicyEffectType string

const (
	PolicyEffectCapacityLimit PolicyEffectType = "capacity_limit"
	PolicyEffectRequireReview PolicyEffectType = "require_review"
	PolicyEffectDiagnostic    PolicyEffectType = "diagnostic"
)

func (t PolicyEffectType) Valid() bool {
	switch t {
	case PolicyEffectCapacityLimit, PolicyEffectRequireReview, PolicyEffectDiagnostic:
		return true
	default:
		return false
	}
}

// PolicyEffect is a closed structured effect descriptor retained in PolicyDecision.
type PolicyEffect struct {
	typeName PolicyEffectType
	placement *Placement
	maximum   *int
	code      string
}

func NewCapacityPolicyEffect(placement Placement, maximum int) (PolicyEffect, error) {
	params, err := NewCapacityLimitParameters(placement, maximum)
	if err != nil {
		return PolicyEffect{}, err
	}
	placementCopy, _ := params.Placement()
	maximumCopy, _ := params.Maximum()
	return PolicyEffect{typeName: PolicyEffectCapacityLimit, placement: clonePlacement(&placementCopy), maximum: cloneInt(&maximumCopy)}, nil
}

func NewRequireReviewPolicyEffect(code string) (PolicyEffect, error) {
	if err := requireText("policy review code", code); err != nil {
		return PolicyEffect{}, err
	}
	return PolicyEffect{typeName: PolicyEffectRequireReview, code: code}, nil
}

func NewDiagnosticPolicyEffect(code string) (PolicyEffect, error) {
	if err := requireText("policy diagnostic code", code); err != nil {
		return PolicyEffect{}, err
	}
	return PolicyEffect{typeName: PolicyEffectDiagnostic, code: code}, nil
}

func (e PolicyEffect) Type() PolicyEffectType { return e.typeName }
func (e PolicyEffect) Placement() (Placement, bool) {
	if e.placement == nil {
		return "", false
	}
	return *e.placement, true
}
func (e PolicyEffect) Maximum() (int, bool) {
	if e.maximum == nil {
		return 0, false
	}
	return *e.maximum, true
}
func (e PolicyEffect) Code() string { return e.code }

func (e PolicyEffect) valid() bool {
	switch e.typeName {
	case PolicyEffectCapacityLimit:
		placement, hasPlacement := e.Placement()
		maximum, hasMaximum := e.Maximum()
		return hasPlacement && hasMaximum && placement.Valid() && placement != PlacementKill && maximum >= 0 && e.code == ""
	case PolicyEffectRequireReview, PolicyEffectDiagnostic:
		return e.placement == nil && e.maximum == nil && requireText("policy effect code", e.code) == nil
	default:
		return false
	}
}

func clonePolicyEffects(values []PolicyEffect) []PolicyEffect {
	if len(values) == 0 {
		return nil
	}
	result := make([]PolicyEffect, len(values))
	for i, value := range values {
		result[i] = PolicyEffect{
			typeName:  value.typeName,
			placement: clonePlacement(value.placement),
			maximum:   cloneInt(value.maximum),
			code:      value.code,
		}
	}
	return result
}

func clonePlacement(value *Placement) *Placement {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}

func cloneInt(value *int) *int {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}

func cloneBool(value *bool) *bool {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}
