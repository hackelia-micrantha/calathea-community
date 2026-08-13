package policy

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/bits"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hackelia-micrantha/calathea-community/internal/domain"
)

const (
	PolicyCapacityNow           domain.PolicyID = "orientation.capacity.now"
	PolicyCapacityNext          domain.PolicyID = "orientation.capacity.next"
	PolicyEvaluationRequired   domain.PolicyID = "orientation.evaluation.required"
	PolicyLifecycleEligibility domain.PolicyID = "orientation.lifecycle.eligibility"
	PolicyEvaluationConfidence domain.PolicyID = "orientation.evaluation.confidence"
	PolicyEvaluationFreshness  domain.PolicyID = "orientation.evaluation.freshness"
)

type baselineRequirement struct {
	policyID      domain.PolicyID
	evaluatorType domain.PolicyEvaluatorType
}

var baselineRequirements = []baselineRequirement{
	{policyID: PolicyLifecycleEligibility, evaluatorType: domain.PolicyEvaluatorLifecycleEligibility},
	{policyID: PolicyEvaluationRequired, evaluatorType: domain.PolicyEvaluatorRequiredEvaluation},
	{policyID: PolicyEvaluationConfidence, evaluatorType: domain.PolicyEvaluatorConfidenceGate},
	{policyID: PolicyEvaluationFreshness, evaluatorType: domain.PolicyEvaluatorFreshnessRule},
	{policyID: PolicyCapacityNow, evaluatorType: domain.PolicyEvaluatorCapacityLimit},
	{policyID: PolicyCapacityNext, evaluatorType: domain.PolicyEvaluatorCapacityLimit},
}

type evaluatorKey struct {
	typeName domain.PolicyEvaluatorType
	version  string
}

type evaluationInput struct {
	asOf          time.Time
	lifecycle     *domain.LifecycleState
	evaluation    *domain.EvaluationVersion
	placement     *domain.Placement
	selectedCount *int
}

type evaluatorOutput struct {
	result                 domain.PolicyDecisionResult
	effects                []domain.PolicyEffect
	reasonCode             string
	inputReferences        []string
	evidenceIDs            []domain.EvidenceReferenceID
	conflictingEvidenceIDs []domain.EvidenceReferenceID
	missingInputs          []string
}

type evaluatorFunc func(domain.PolicyInstance, evaluationInput) (evaluatorOutput, error)

// Engine is a deterministic registry of built-in policy evaluators. Registry
// membership and evaluator semantic versions are explicit so unsupported
// historical policy sets fail rather than silently substituting behavior.
type Engine struct {
	evaluators map[evaluatorKey]evaluatorFunc
}

func NewV0Engine() *Engine {
	return &Engine{evaluators: map[evaluatorKey]evaluatorFunc{
		{domain.PolicyEvaluatorCapacityLimit, domain.PolicyEvaluatorSemanticVersionV1}:        evaluateCapacityLimit,
		{domain.PolicyEvaluatorRequiredEvaluation, domain.PolicyEvaluatorSemanticVersionV1}:   evaluateRequiredEvaluation,
		{domain.PolicyEvaluatorLifecycleEligibility, domain.PolicyEvaluatorSemanticVersionV1}: evaluateLifecycleEligibility,
		{domain.PolicyEvaluatorConfidenceGate, domain.PolicyEvaluatorSemanticVersionV1}:       evaluateConfidenceGate,
		{domain.PolicyEvaluatorFreshnessRule, domain.PolicyEvaluatorSemanticVersionV1}:        evaluateFreshnessRule,
		{domain.PolicyEvaluatorScoreMultiplier, domain.PolicyEvaluatorSemanticVersionV1}:      evaluateScoreMultiplier,
	}}
}

type BaselinePolicySetConfig struct {
	ID                  domain.PolicySetVersionID
	PolicySet           domain.PolicySet
	CreatedAt           time.Time
	MaxNext             *int
	MinimumConfidence   domain.ConfidenceBand
	FreshnessMaximumAge time.Duration
}

// NewBaselinePolicySetVersion constructs the offline UC-01 baseline. The now
// capacity default is fixed at 3 by the accepted RFC. A nil MaxNext uses the
// accepted initial default of 10; an explicit zero remains a valid zero-capacity
// queue rather than being reinterpreted as an omitted value.
func NewBaselinePolicySetVersion(config BaselinePolicySetConfig) (domain.PolicySetVersion, error) {
	maxNext := 10
	if config.MaxNext != nil {
		maxNext = *config.MaxNext
	}
	if maxNext < 0 {
		return domain.PolicySetVersion{}, fmt.Errorf("baseline next capacity must not be negative")
	}
	if uint64(maxNext) > uint64(^uint32(0)) {
		return domain.PolicySetVersion{}, fmt.Errorf("baseline next capacity %d exceeds supported maximum", maxNext)
	}
	minimumConfidence := config.MinimumConfidence
	if minimumConfidence == "" {
		minimumConfidence = domain.ConfidenceVisibleUncertainty
	}
	confidenceParameters, err := domain.NewConfidenceGateParameters(minimumConfidence)
	if err != nil {
		return domain.PolicySetVersion{}, err
	}
	freshnessParameters, err := domain.NewFreshnessRuleParameters(config.FreshnessMaximumAge)
	if err != nil {
		return domain.PolicySetVersion{}, err
	}
	nowParameters, err := domain.NewCapacityLimitParameters(domain.PlacementNow, 3)
	if err != nil {
		return domain.PolicySetVersion{}, err
	}
	nextParameters, err := domain.NewCapacityLimitParameters(domain.PlacementNext, maxNext)
	if err != nil {
		return domain.PolicySetVersion{}, err
	}

	instances := make([]domain.PolicyInstance, 0, 6)
	appendInstance := func(instance domain.PolicyInstance, instanceErr error) error {
		if instanceErr != nil {
			return instanceErr
		}
		instances = append(instances, instance)
		return nil
	}
	if err := appendInstance(domain.NewLifecycleEligibilityPolicyInstance(
		"baseline-lifecycle-eligibility",
		PolicyLifecycleEligibility,
		domain.NewLifecycleEligibilityParameters(false),
		10,
		"Normal active orientation excludes candidate and terminal lifecycle states.",
	)); err != nil {
		return domain.PolicySetVersion{}, err
	}
	if err := appendInstance(domain.NewRequiredEvaluationPolicyInstance(
		"baseline-evaluation-required",
		PolicyEvaluationRequired,
		20,
		domain.PolicyMissingExcludeSubject,
		domain.PolicyNotExceptionable,
		"UC-01 requires an accepted evaluation before active orientation.",
	)); err != nil {
		return domain.PolicySetVersion{}, err
	}
	if err := appendInstance(domain.NewConfidenceGatePolicyInstance(
		"baseline-confidence",
		PolicyEvaluationConfidence,
		confidenceParameters,
		30,
		domain.PolicyMissingRequireReview,
		"Weak evaluation evidence remains visible and requires review.",
	)); err != nil {
		return domain.PolicySetVersion{}, err
	}
	if err := appendInstance(domain.NewFreshnessRulePolicyInstance(
		"baseline-freshness",
		PolicyEvaluationFreshness,
		freshnessParameters,
		40,
		domain.PolicyMissingRequireReview,
		"Stale evaluation evidence requires explicit review.",
	)); err != nil {
		return domain.PolicySetVersion{}, err
	}
	if err := appendInstance(domain.NewCapacityLimitPolicyInstance(
		"baseline-now-capacity",
		PolicyCapacityNow,
		nowParameters,
		10,
		"Bound the now queue to the accepted default capacity.",
	)); err != nil {
		return domain.PolicySetVersion{}, err
	}
	if err := appendInstance(domain.NewCapacityLimitPolicyInstance(
		"baseline-next-capacity",
		PolicyCapacityNext,
		nextParameters,
		20,
		"Bound the next queue explicitly.",
	)); err != nil {
		return domain.PolicySetVersion{}, err
	}

	version, err := domain.NewPolicySetVersion(domain.PolicySetVersionInput{
		ID:                             config.ID,
		PolicySet:                      config.PolicySet,
		CreatedAt:                      config.CreatedAt,
		Instances:                      instances,
		MinimumScoreMultiplierBasisPts: 10000,
		MaximumScoreMultiplierBasisPts: 10000,
	})
	if err != nil {
		return domain.PolicySetVersion{}, err
	}
	if err := NewV0Engine().ValidatePolicySetVersion(version); err != nil {
		return domain.PolicySetVersion{}, err
	}
	return version, nil
}

// PolicyEvaluationFailure is an operational evaluator failure, not an
// indeterminate policy result.
type PolicyEvaluationFailure struct {
	InstanceID    domain.PolicyInstanceID
	EvaluatorType domain.PolicyEvaluatorType
	Version       string
	Cause         error
}

func (f *PolicyEvaluationFailure) Error() string {
	return fmt.Sprintf("policy evaluator %s@%s for instance %s failed: %v", f.EvaluatorType, f.Version, f.InstanceID, f.Cause)
}

func (f *PolicyEvaluationFailure) Unwrap() error { return f.Cause }

type ProjectContext struct {
	OperationID    domain.OperationID
	ProjectID      domain.ProjectID
	LifecycleState domain.LifecycleState
	Evaluation     *domain.EvaluationVersion
	EvaluatedAt    time.Time
}

type CapacityContext struct {
	OperationID   domain.OperationID
	ProjectID     domain.ProjectID
	Placement     domain.Placement
	SelectedCount int
	EvaluatedAt   time.Time
}

type ExactMultiplier struct {
	numerator   uint64
	denominator uint64
}

func (m ExactMultiplier) Numerator() uint64   { return m.numerator }
func (m ExactMultiplier) Denominator() uint64 { return m.denominator }
func (m ExactMultiplier) String() string {
	if m.denominator == 1 {
		return strconv.FormatUint(m.numerator, 10)
	}
	return fmt.Sprintf("%d/%d", m.numerator, m.denominator)
}

type MultiplierCompositionStep struct {
	DecisionID        domain.PolicyDecisionID
	Before            ExactMultiplier
	FactorBasisPoints uint32
	ProposedAfter     ExactMultiplier
	Suppressed        bool
	ReasonCode        string
}

type CompositeOutcome struct {
	Denied                bool
	Excluded              bool
	ReviewRequired        bool
	Indeterminate         bool
	SoftEffectsSuppressed bool
	ScoreMultiplier       ExactMultiplier
	MultiplierSteps       []MultiplierCompositionStep
	SuppressedDecisionIDs []domain.PolicyDecisionID
}

type EvaluationResult struct {
	Decisions []domain.PolicyDecision
	Outcome   CompositeOutcome
}

func (e *Engine) ValidatePolicySetVersion(version domain.PolicySetVersion) error {
	instances := version.Instances()
	if len(instances) == 0 {
		return fmt.Errorf("policy set version %q contains no policy instances", version.ID())
	}

	required := make(map[domain.PolicyID]domain.PolicyEvaluatorType, len(baselineRequirements))
	for _, requirement := range baselineRequirements {
		required[requirement.policyID] = requirement.evaluatorType
	}
	seenRequired := make(map[domain.PolicyID]int, len(required))
	conflictKeys := make(map[string]domain.PolicyInstanceID)
	multiplier := ExactMultiplier{numerator: 1, denominator: 1}

	for _, instance := range instances {
		key := evaluatorKey{instance.EvaluatorType(), instance.EvaluatorVersion()}
		if _, ok := e.evaluators[key]; !ok {
			return fmt.Errorf("unsupported evaluator %s@%s for instance %q", instance.EvaluatorType(), instance.EvaluatorVersion(), instance.ID())
		}
		if err := validateInstanceShape(instance); err != nil {
			return fmt.Errorf("invalid policy instance %q: %w", instance.ID(), err)
		}
		if expected, ok := required[instance.PolicyID()]; ok {
			if instance.EvaluatorType() != expected {
				return fmt.Errorf("baseline policy %q requires evaluator %q, got %q", instance.PolicyID(), expected, instance.EvaluatorType())
			}
			seenRequired[instance.PolicyID()]++
		}
		if conflictKey := instance.ConflictKey(); conflictKey != "" {
			if prior, exists := conflictKeys[conflictKey]; exists {
				return fmt.Errorf("policy conflict key %q is configured by both %q and %q", conflictKey, prior, instance.ID())
			}
			conflictKeys[conflictKey] = instance.ID()
		}
		if parameters, ok := instance.CapacityLimitParameters(); ok {
			if instance.PolicyID() == PolicyCapacityNow && parameters.Placement() != domain.PlacementNow {
				return fmt.Errorf("%s must constrain now placement", PolicyCapacityNow)
			}
			if instance.PolicyID() == PolicyCapacityNext && parameters.Placement() != domain.PlacementNext {
				return fmt.Errorf("%s must constrain next placement", PolicyCapacityNext)
			}
		}
		if parameters, ok := instance.ScoreMultiplierParameters(); ok {
			var err error
			multiplier, err = multiplyExact(multiplier, parameters.BasisPoints())
			if err != nil {
				return fmt.Errorf("score multiplier composition for instance %q: %w", instance.ID(), err)
			}
		}
	}
	for _, requirement := range baselineRequirements {
		if seenRequired[requirement.policyID] != 1 {
			return fmt.Errorf("baseline policy %q must appear exactly once; found %d", requirement.policyID, seenRequired[requirement.policyID])
		}
	}
	minimum, maximum := version.ScoreMultiplierBoundsBasisPoints()
	if ratioLessThanBasisPoints(multiplier, minimum) || ratioGreaterThanBasisPoints(multiplier, maximum) {
		return fmt.Errorf("cumulative score multiplier %s is outside policy-set bounds %d-%d basis points", multiplier.String(), minimum, maximum)
	}
	return nil
}

func validateInstanceShape(instance domain.PolicyInstance) error {
	if instance.EvaluatorVersion() != domain.PolicyEvaluatorSemanticVersionV1 {
		return fmt.Errorf("unsupported evaluator semantic version %q", instance.EvaluatorVersion())
	}
	if instance.ConfigurationSchemaVersion() != domain.PolicyConfigurationSchemaVersionV1 {
		return fmt.Errorf("unsupported policy configuration schema version %q", instance.ConfigurationSchemaVersion())
	}
	if instance.Workflow() != domain.PolicyWorkflowOrientation {
		return fmt.Errorf("unsupported policy workflow %q", instance.Workflow())
	}
	if instance.SubjectType() != domain.PolicySubjectProject {
		return fmt.Errorf("unsupported policy subject type %q", instance.SubjectType())
	}
	if len(instance.RequiredInputs()) == 0 {
		return fmt.Errorf("policy instance requires explicit required-input declarations")
	}
	switch instance.EvaluatorType() {
	case domain.PolicyEvaluatorCapacityLimit:
		if instance.EffectClass() != domain.PolicyEffectHard || instance.Phase() != domain.PolicyPhaseSetConstraints {
			return fmt.Errorf("capacity_limit requires hard set_constraints semantics")
		}
		if _, ok := instance.CapacityLimitParameters(); !ok {
			return fmt.Errorf("capacity_limit parameters missing")
		}
	case domain.PolicyEvaluatorRequiredEvaluation:
		if instance.EffectClass() != domain.PolicyEffectHard || instance.Phase() != domain.PolicyPhaseCandidateEligibility {
			return fmt.Errorf("required_evaluation requires hard candidate_eligibility semantics")
		}
	case domain.PolicyEvaluatorLifecycleEligibility:
		if instance.EffectClass() != domain.PolicyEffectHard || instance.Phase() != domain.PolicyPhaseCandidateEligibility {
			return fmt.Errorf("lifecycle_eligibility requires hard candidate_eligibility semantics")
		}
		if _, ok := instance.LifecycleEligibilityParameters(); !ok {
			return fmt.Errorf("lifecycle_eligibility parameters missing")
		}
	case domain.PolicyEvaluatorConfidenceGate:
		if instance.EffectClass() != domain.PolicyEffectReviewRequired || instance.Phase() != domain.PolicyPhaseCandidateAdjustment {
			return fmt.Errorf("confidence_gate requires review_required candidate_adjustment semantics")
		}
		if _, ok := instance.ConfidenceGateParameters(); !ok {
			return fmt.Errorf("confidence_gate parameters missing")
		}
	case domain.PolicyEvaluatorFreshnessRule:
		if instance.EffectClass() != domain.PolicyEffectReviewRequired || instance.Phase() != domain.PolicyPhaseCandidateAdjustment {
			return fmt.Errorf("freshness_rule requires review_required candidate_adjustment semantics")
		}
		if _, ok := instance.FreshnessRuleParameters(); !ok {
			return fmt.Errorf("freshness_rule parameters missing")
		}
	case domain.PolicyEvaluatorScoreMultiplier:
		if instance.EffectClass() != domain.PolicyEffectSoft || instance.Phase() != domain.PolicyPhaseCandidateAdjustment {
			return fmt.Errorf("score_multiplier requires soft candidate_adjustment semantics")
		}
		if _, ok := instance.ScoreMultiplierParameters(); !ok {
			return fmt.Errorf("score_multiplier parameters missing")
		}
	default:
		return fmt.Errorf("unsupported evaluator type %q", instance.EvaluatorType())
	}
	return nil
}

func (e *Engine) EvaluateProject(version domain.PolicySetVersion, context ProjectContext) (EvaluationResult, error) {
	if err := e.ValidatePolicySetVersion(version); err != nil {
		return EvaluationResult{}, err
	}
	if strings.TrimSpace(string(context.OperationID)) == "" {
		return EvaluationResult{}, fmt.Errorf("operation id must not be empty")
	}
	if strings.TrimSpace(string(context.ProjectID)) == "" {
		return EvaluationResult{}, fmt.Errorf("project id must not be empty")
	}
	if !context.LifecycleState.Valid() {
		return EvaluationResult{}, fmt.Errorf("invalid lifecycle state %q", context.LifecycleState)
	}
	if context.EvaluatedAt.IsZero() {
		return EvaluationResult{}, fmt.Errorf("policy evaluation time must not be zero")
	}
	if context.Evaluation != nil {
		if context.Evaluation.ProjectID() != context.ProjectID {
			return EvaluationResult{}, fmt.Errorf("evaluation version %q belongs to project %q, want %q", context.Evaluation.ID(), context.Evaluation.ProjectID(), context.ProjectID)
		}
		if context.Evaluation.AcceptedAt().After(context.EvaluatedAt) {
			return EvaluationResult{}, fmt.Errorf("evaluation version %q was accepted after policy evaluation time", context.Evaluation.ID())
		}
	}

	lifecycle := context.LifecycleState
	input := evaluationInput{asOf: context.EvaluatedAt, lifecycle: &lifecycle, evaluation: context.Evaluation}
	decisions := make([]domain.PolicyDecision, 0, len(version.Instances()))
	for _, instance := range version.Instances() {
		switch instance.Phase() {
		case domain.PolicyPhaseCandidateEligibility, domain.PolicyPhaseCandidateAdjustment, domain.PolicyPhaseReviewDiagnostics:
		default:
			continue
		}
		decision, err := e.evaluateInstance(version, instance, context.ProjectID, context.OperationID, context.EvaluatedAt, "project", input)
		if err != nil {
			return EvaluationResult{}, err
		}
		decisions = append(decisions, decision)
	}
	outcome, err := Compose(version, decisions)
	if err != nil {
		return EvaluationResult{}, err
	}
	return EvaluationResult{Decisions: append([]domain.PolicyDecision(nil), decisions...), Outcome: cloneOutcome(outcome)}, nil
}

func (e *Engine) EvaluateCapacity(version domain.PolicySetVersion, context CapacityContext) (EvaluationResult, error) {
	if err := e.ValidatePolicySetVersion(version); err != nil {
		return EvaluationResult{}, err
	}
	if strings.TrimSpace(string(context.OperationID)) == "" {
		return EvaluationResult{}, fmt.Errorf("operation id must not be empty")
	}
	if strings.TrimSpace(string(context.ProjectID)) == "" {
		return EvaluationResult{}, fmt.Errorf("project id must not be empty")
	}
	if !context.Placement.Valid() || context.Placement == domain.PlacementKill {
		return EvaluationResult{}, fmt.Errorf("invalid capacity placement %q", context.Placement)
	}
	if context.SelectedCount < 0 {
		return EvaluationResult{}, fmt.Errorf("selected count must not be negative")
	}
	if context.EvaluatedAt.IsZero() {
		return EvaluationResult{}, fmt.Errorf("policy evaluation time must not be zero")
	}
	placement := context.Placement
	selectedCount := context.SelectedCount
	input := evaluationInput{asOf: context.EvaluatedAt, placement: &placement, selectedCount: &selectedCount}
	decisions := make([]domain.PolicyDecision, 0, 2)
	for _, instance := range version.Instances() {
		if instance.Phase() != domain.PolicyPhaseSetConstraints {
			continue
		}
		discriminator := fmt.Sprintf("placement:%s:selected:%d", context.Placement, context.SelectedCount)
		decision, err := e.evaluateInstance(version, instance, context.ProjectID, context.OperationID, context.EvaluatedAt, discriminator, input)
		if err != nil {
			return EvaluationResult{}, err
		}
		decisions = append(decisions, decision)
	}
	outcome, err := Compose(version, decisions)
	if err != nil {
		return EvaluationResult{}, err
	}
	return EvaluationResult{Decisions: append([]domain.PolicyDecision(nil), decisions...), Outcome: cloneOutcome(outcome)}, nil
}

func (e *Engine) evaluateInstance(version domain.PolicySetVersion, instance domain.PolicyInstance, projectID domain.ProjectID, operationID domain.OperationID, evaluatedAt time.Time, discriminator string, input evaluationInput) (domain.PolicyDecision, error) {
	var output evaluatorOutput
	if !instance.SubjectSelector().MatchesProject(projectID) {
		output = evaluatorOutput{
			result:          domain.PolicyDecisionNotApplicable,
			reasonCode:      "subject_selector_not_matched",
			inputReferences: []string{"project_id:" + string(projectID)},
		}
	} else {
		key := evaluatorKey{instance.EvaluatorType(), instance.EvaluatorVersion()}
		evaluator, ok := e.evaluators[key]
		if !ok {
			return domain.PolicyDecision{}, &PolicyEvaluationFailure{
				InstanceID:    instance.ID(),
				EvaluatorType: instance.EvaluatorType(),
				Version:       instance.EvaluatorVersion(),
				Cause:         fmt.Errorf("evaluator unavailable"),
			}
		}
		var err error
		output, err = evaluator(instance, input)
		if err != nil {
			return domain.PolicyDecision{}, &PolicyEvaluationFailure{
				InstanceID:    instance.ID(),
				EvaluatorType: instance.EvaluatorType(),
				Version:       instance.EvaluatorVersion(),
				Cause:         err,
			}
		}
	}
	applicability := domain.PolicyApplicable
	if output.result == domain.PolicyDecisionNotApplicable {
		applicability = domain.PolicyNotApplicable
	}
	decisionID := deterministicDecisionID(operationID, instance.ID(), projectID, discriminator)
	decision, err := domain.NewPolicyDecision(domain.PolicyDecisionInput{
		ID:                         decisionID,
		SchemaVersion:              domain.PolicyDecisionSchemaVersionV1,
		PolicySetVersionID:         version.ID(),
		PolicyID:                   instance.PolicyID(),
		PolicyInstanceID:           instance.ID(),
		EvaluatorType:              instance.EvaluatorType(),
		EvaluatorVersion:           instance.EvaluatorVersion(),
		ConfigurationSchemaVersion: instance.ConfigurationSchemaVersion(),
		Workflow:                   instance.Workflow(),
		Phase:                      instance.Phase(),
		SubjectType:                instance.SubjectType(),
		ProjectID:                  projectID,
		OperationID:                operationID,
		Applicability:              applicability,
		Result:                     output.result,
		EffectClass:                instance.EffectClass(),
		Effects:                    output.effects,
		RequiredInputs:             instance.RequiredInputs(),
		InputReferences:            output.inputReferences,
		EvidenceIDs:                output.evidenceIDs,
		ConflictingEvidenceIDs:     output.conflictingEvidenceIDs,
		MissingInputs:              output.missingInputs,
		MissingInputBehavior:       instance.MissingInputBehavior(),
		Rationale:                  instance.Rationale(),
		ReasonCode:                 output.reasonCode,
		Priority:                   instance.Priority(),
		ConflictKey:                instance.ConflictKey(),
		CreatedAt:                  evaluatedAt,
	})
	if err != nil {
		return domain.PolicyDecision{}, &PolicyEvaluationFailure{
			InstanceID:    instance.ID(),
			EvaluatorType: instance.EvaluatorType(),
			Version:       instance.EvaluatorVersion(),
			Cause:         err,
		}
	}
	return decision, nil
}

func deterministicDecisionID(operationID domain.OperationID, instanceID domain.PolicyInstanceID, projectID domain.ProjectID, discriminator string) domain.PolicyDecisionID {
	payload := strings.Join([]string{string(operationID), string(instanceID), string(projectID), discriminator}, "\x00")
	digest := sha256.Sum256([]byte(payload))
	return domain.PolicyDecisionID("policy-decision-" + hex.EncodeToString(digest[:16]))
}

func Compose(version domain.PolicySetVersion, decisions []domain.PolicyDecision) (CompositeOutcome, error) {
	ordered := append([]domain.PolicyDecision(nil), decisions...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if phaseRank(ordered[i].Phase()) != phaseRank(ordered[j].Phase()) {
			return phaseRank(ordered[i].Phase()) < phaseRank(ordered[j].Phase())
		}
		if ordered[i].Priority() != ordered[j].Priority() {
			return ordered[i].Priority() < ordered[j].Priority()
		}
		if ordered[i].PolicyID() != ordered[j].PolicyID() {
			return ordered[i].PolicyID() < ordered[j].PolicyID()
		}
		return ordered[i].PolicyInstanceID() < ordered[j].PolicyInstanceID()
	})

	outcome := CompositeOutcome{ScoreMultiplier: ExactMultiplier{numerator: 1, denominator: 1}}
	for _, decision := range ordered {
		switch decision.Result() {
		case domain.PolicyDecisionAllow, domain.PolicyDecisionNotApplicable:
		case domain.PolicyDecisionDeny:
			outcome.Denied = true
		case domain.PolicyDecisionRequireReview:
			outcome.ReviewRequired = true
		case domain.PolicyDecisionAdjust:
			for _, effect := range decision.Effects() {
				if effect.Kind() != domain.PolicyEffectScoreMultiplier {
					continue
				}
				before := outcome.ScoreMultiplier
				proposedAfter, err := multiplyExact(before, effect.MultiplierBasisPoints())
				if err != nil {
					return CompositeOutcome{}, fmt.Errorf("compose score multiplier for decision %q: %w", decision.ID(), err)
				}
				step := MultiplierCompositionStep{
					DecisionID:        decision.ID(),
					Before:            before,
					FactorBasisPoints: effect.MultiplierBasisPoints(),
					ProposedAfter:     proposedAfter,
				}
				if outcome.Denied || outcome.Excluded {
					step.Suppressed = true
					step.ReasonCode = "suppressed_by_hard_policy"
					outcome.SoftEffectsSuppressed = true
					outcome.SuppressedDecisionIDs = appendUniqueDecisionID(outcome.SuppressedDecisionIDs, decision.ID())
				} else {
					outcome.ScoreMultiplier = proposedAfter
				}
				outcome.MultiplierSteps = append(outcome.MultiplierSteps, step)
			}
		case domain.PolicyDecisionIndeterminate:
			outcome.Indeterminate = true
			switch decision.MissingInputBehavior() {
			case domain.PolicyMissingDeny:
				outcome.Denied = true
			case domain.PolicyMissingExcludeSubject:
				outcome.Excluded = true
			case domain.PolicyMissingRequireReview:
				outcome.ReviewRequired = true
			case domain.PolicyMissingFailOperation:
				return CompositeOutcome{}, fmt.Errorf("policy decision %q requires operation failure because required input is indeterminate", decision.ID())
			case domain.PolicyMissingDiagnosticOnly:
				if decision.EffectClass() != domain.PolicyEffectAdvisory {
					return CompositeOutcome{}, fmt.Errorf("policy decision %q uses diagnostic_only outside advisory policy", decision.ID())
				}
			default:
				return CompositeOutcome{}, fmt.Errorf("policy decision %q has unsupported missing-input behavior %q", decision.ID(), decision.MissingInputBehavior())
			}
		default:
			return CompositeOutcome{}, fmt.Errorf("policy decision %q has unsupported result %q", decision.ID(), decision.Result())
		}
	}

	if (outcome.Denied || outcome.Excluded) && outcome.ScoreMultiplier.numerator != outcome.ScoreMultiplier.denominator {
		outcome.SoftEffectsSuppressed = true
		for i := range outcome.MultiplierSteps {
			if outcome.MultiplierSteps[i].Suppressed {
				continue
			}
			outcome.MultiplierSteps[i].Suppressed = true
			outcome.MultiplierSteps[i].ReasonCode = "suppressed_by_hard_policy"
			outcome.SuppressedDecisionIDs = appendUniqueDecisionID(outcome.SuppressedDecisionIDs, outcome.MultiplierSteps[i].DecisionID)
		}
		outcome.ScoreMultiplier = ExactMultiplier{numerator: 1, denominator: 1}
	}
	minimum, maximum := version.ScoreMultiplierBoundsBasisPoints()
	if ratioLessThanBasisPoints(outcome.ScoreMultiplier, minimum) || ratioGreaterThanBasisPoints(outcome.ScoreMultiplier, maximum) {
		return CompositeOutcome{}, fmt.Errorf("composed score multiplier %s is outside policy-set bounds %d-%d basis points", outcome.ScoreMultiplier.String(), minimum, maximum)
	}
	return cloneOutcome(outcome), nil
}

func cloneOutcome(outcome CompositeOutcome) CompositeOutcome {
	outcome.MultiplierSteps = append([]MultiplierCompositionStep(nil), outcome.MultiplierSteps...)
	outcome.SuppressedDecisionIDs = append([]domain.PolicyDecisionID(nil), outcome.SuppressedDecisionIDs...)
	return outcome
}

func appendUniqueDecisionID(values []domain.PolicyDecisionID, value domain.PolicyDecisionID) []domain.PolicyDecisionID {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func phaseRank(phase domain.PolicyPhase) int {
	switch phase {
	case domain.PolicyPhaseCandidateEligibility:
		return 1
	case domain.PolicyPhaseCandidateAdjustment:
		return 2
	case domain.PolicyPhaseSetConstraints:
		return 3
	case domain.PolicyPhaseResultValidation:
		return 4
	case domain.PolicyPhaseDispositionValidation:
		return 5
	case domain.PolicyPhaseReviewDiagnostics:
		return 6
	default:
		return 100
	}
}

func (e *Engine) ValidateExceptionUse(version domain.PolicySetVersion, exception domain.PolicyException, applications []domain.PolicyExceptionApplication, revocations []domain.PolicyExceptionRevocation, projectID domain.ProjectID, phase domain.PolicyPhase, at time.Time) error {
	if at.IsZero() {
		return fmt.Errorf("exception validation time must not be zero")
	}
	if len(exception.EvidenceIDs()) == 0 {
		return fmt.Errorf("policy exception %q requires supporting evidence/provenance", exception.ID())
	}
	var instance *domain.PolicyInstance
	for _, candidate := range version.Instances() {
		if candidate.ID() == exception.PolicyInstanceID() {
			copyCandidate := candidate
			instance = &copyCandidate
			break
		}
	}
	if instance == nil {
		return fmt.Errorf("policy exception %q references unknown instance %q", exception.ID(), exception.PolicyInstanceID())
	}
	if instance.Exceptionability() == domain.PolicyNotExceptionable {
		return fmt.Errorf("policy instance %q is not exceptionable", instance.ID())
	}
	if exception.PolicySetVersionID() != version.ID() || exception.PolicyID() != instance.PolicyID() || exception.EvaluatorVersion() != instance.EvaluatorVersion() {
		return fmt.Errorf("policy exception %q is incompatible with policy instance %q", exception.ID(), instance.ID())
	}
	if exception.ProjectID() != projectID {
		return fmt.Errorf("policy exception %q does not apply to project %q", exception.ID(), projectID)
	}
	if exception.Phase() != phase || instance.Phase() != phase {
		return fmt.Errorf("policy exception %q does not apply to phase %q", exception.ID(), phase)
	}
	if !exception.EffectiveAtTime(at) {
		return fmt.Errorf("policy exception %q is not effective at %s", exception.ID(), at.UTC().Format(time.RFC3339))
	}
	var earliestRevocation *time.Time
	for _, revocation := range revocations {
		if revocation.ExceptionID() != exception.ID() || revocation.RevokedAt().After(at) {
			continue
		}
		revokedAt := revocation.RevokedAt()
		if earliestRevocation == nil || revokedAt.Before(*earliestRevocation) {
			copyTime := revokedAt
			earliestRevocation = &copyTime
		}
	}
	if earliestRevocation != nil {
		return fmt.Errorf("policy exception %q was revoked at %s", exception.ID(), earliestRevocation.UTC().Format(time.RFC3339))
	}
	var uses uint32
	for _, application := range applications {
		if application.ExceptionID() == exception.ID() && !application.AppliedAt().After(at) {
			uses++
		}
	}
	if uses >= exception.MaximumUses() {
		return fmt.Errorf("policy exception %q exhausted its %d allowed uses", exception.ID(), exception.MaximumUses())
	}
	return nil
}

func evaluateCapacityLimit(instance domain.PolicyInstance, input evaluationInput) (evaluatorOutput, error) {
	parameters, ok := instance.CapacityLimitParameters()
	if !ok {
		return evaluatorOutput{}, fmt.Errorf("capacity parameters missing")
	}
	if input.placement == nil || input.selectedCount == nil {
		missing := make([]string, 0, 2)
		if input.placement == nil {
			missing = append(missing, "placement")
		}
		if input.selectedCount == nil {
			missing = append(missing, "selected_count")
		}
		return evaluatorOutput{
			result:        domain.PolicyDecisionIndeterminate,
			reasonCode:    "capacity_input_missing",
			missingInputs: missing,
		}, nil
	}
	if *input.placement != parameters.Placement() {
		return evaluatorOutput{
			result:          domain.PolicyDecisionNotApplicable,
			reasonCode:      "capacity_other_placement",
			inputReferences: []string{"placement:" + string(*input.placement)},
		}, nil
	}
	effect, err := domain.NewCapacityLimitPolicyEffect(parameters.Placement(), int(parameters.Maximum()))
	if err != nil {
		return evaluatorOutput{}, err
	}
	inputReferences := []string{
		"placement:" + string(*input.placement),
		"selected_count:" + strconv.Itoa(*input.selectedCount),
		"maximum:" + strconv.FormatUint(uint64(parameters.Maximum()), 10),
	}
	if uint64(input.selectedCountValue()) >= uint64(parameters.Maximum()) {
		return evaluatorOutput{
			result:          domain.PolicyDecisionDeny,
			effects:         []domain.PolicyEffect{effect},
			reasonCode:      "capacity_exhausted",
			inputReferences: inputReferences,
		}, nil
	}
	return evaluatorOutput{
		result:          domain.PolicyDecisionAllow,
		effects:         []domain.PolicyEffect{effect},
		reasonCode:      "capacity_available",
		inputReferences: inputReferences,
	}, nil
}

func (input evaluationInput) selectedCountValue() int {
	if input.selectedCount == nil {
		return 0
	}
	return *input.selectedCount
}

func evaluateRequiredEvaluation(_ domain.PolicyInstance, input evaluationInput) (evaluatorOutput, error) {
	if input.evaluation == nil {
		return evaluatorOutput{
			result:        domain.PolicyDecisionIndeterminate,
			reasonCode:    "accepted_evaluation_missing",
			missingInputs: []string{"accepted_evaluation"},
		}, nil
	}
	return evaluatorOutput{
		result:          domain.PolicyDecisionAllow,
		reasonCode:      "accepted_evaluation_present",
		inputReferences: []string{"evaluation_version:" + string(input.evaluation.ID())},
		evidenceIDs:     input.evaluation.EvidenceIDs(),
	}, nil
}

func evaluateLifecycleEligibility(instance domain.PolicyInstance, input evaluationInput) (evaluatorOutput, error) {
	parameters, ok := instance.LifecycleEligibilityParameters()
	if !ok {
		return evaluatorOutput{}, fmt.Errorf("lifecycle eligibility parameters missing")
	}
	if input.lifecycle == nil {
		return evaluatorOutput{
			result:        domain.PolicyDecisionIndeterminate,
			reasonCode:    "lifecycle_state_missing",
			missingInputs: []string{"lifecycle_state"},
		}, nil
	}
	state := *input.lifecycle
	if !state.Valid() {
		return evaluatorOutput{}, fmt.Errorf("invalid lifecycle state %q", state)
	}
	inputReferences := []string{"lifecycle_state:" + string(state)}
	switch state {
	case domain.LifecycleApproved, domain.LifecycleActive, domain.LifecyclePaused:
		return evaluatorOutput{result: domain.PolicyDecisionAllow, reasonCode: "lifecycle_eligible", inputReferences: inputReferences}, nil
	case domain.LifecycleProposed:
		if parameters.AllowProposed() {
			return evaluatorOutput{result: domain.PolicyDecisionAllow, reasonCode: "proposed_allowed_by_policy", inputReferences: inputReferences}, nil
		}
		return evaluatorOutput{result: domain.PolicyDecisionDeny, reasonCode: "proposed_not_enabled", inputReferences: inputReferences}, nil
	case domain.LifecycleCandidate:
		return evaluatorOutput{result: domain.PolicyDecisionDeny, reasonCode: "candidate_not_orientation_eligible", inputReferences: inputReferences}, nil
	case domain.LifecycleCompleted, domain.LifecycleCancelled, domain.LifecycleArchived:
		return evaluatorOutput{result: domain.PolicyDecisionDeny, reasonCode: "terminal_lifecycle_excluded", inputReferences: inputReferences}, nil
	default:
		return evaluatorOutput{}, fmt.Errorf("unsupported lifecycle state %q", state)
	}
}

func evaluateConfidenceGate(instance domain.PolicyInstance, input evaluationInput) (evaluatorOutput, error) {
	parameters, ok := instance.ConfidenceGateParameters()
	if !ok {
		return evaluatorOutput{}, fmt.Errorf("confidence gate parameters missing")
	}
	if input.evaluation == nil {
		return evaluatorOutput{
			result:        domain.PolicyDecisionIndeterminate,
			reasonCode:    "confidence_unavailable",
			missingInputs: []string{"accepted_evaluation", "confidence_band"},
		}, nil
	}
	band := input.evaluation.Confidence().Band()
	diagnostic, err := domain.NewDiagnosticPolicyEffect("confidence_band", string(band))
	if err != nil {
		return evaluatorOutput{}, err
	}
	output := evaluatorOutput{
		effects:         []domain.PolicyEffect{diagnostic},
		inputReferences: []string{"evaluation_version:" + string(input.evaluation.ID()), "confidence_band:" + string(band)},
		evidenceIDs:     input.evaluation.EvidenceIDs(),
	}
	if confidenceRank(band) < confidenceRank(parameters.MinimumBand()) {
		output.result = domain.PolicyDecisionRequireReview
		output.reasonCode = "confidence_below_threshold"
		return output, nil
	}
	output.result = domain.PolicyDecisionAllow
	output.reasonCode = "confidence_sufficient"
	return output, nil
}

func evaluateFreshnessRule(instance domain.PolicyInstance, input evaluationInput) (evaluatorOutput, error) {
	parameters, ok := instance.FreshnessRuleParameters()
	if !ok {
		return evaluatorOutput{}, fmt.Errorf("freshness rule parameters missing")
	}
	if input.evaluation == nil {
		return evaluatorOutput{
			result:        domain.PolicyDecisionIndeterminate,
			reasonCode:    "freshness_unavailable",
			missingInputs: []string{"accepted_evaluation", "evaluation_evidence_as_of", "planning_horizon"},
		}, nil
	}
	evidenceAsOf := input.evaluation.Freshness().EvidenceAsOf()
	if input.asOf.Before(evidenceAsOf) {
		return evaluatorOutput{}, fmt.Errorf("policy evaluation time precedes evaluation evidence time")
	}
	age := input.asOf.Sub(evidenceAsOf)
	diagnostic, err := domain.NewDiagnosticPolicyEffect("freshness_age_seconds", strconv.FormatInt(int64(age/time.Second), 10))
	if err != nil {
		return evaluatorOutput{}, err
	}
	output := evaluatorOutput{
		effects: []domain.PolicyEffect{diagnostic},
		inputReferences: []string{
			"evaluation_version:" + string(input.evaluation.ID()),
			"planning_horizon:" + input.evaluation.PlanningHorizon(),
			"evidence_as_of:" + evidenceAsOf.UTC().Format(time.RFC3339Nano),
			"policy_evaluation_time:" + input.asOf.UTC().Format(time.RFC3339Nano),
			"freshness_max_age_seconds:" + strconv.FormatInt(int64(parameters.MaximumAge()/time.Second), 10),
		},
		evidenceIDs: input.evaluation.EvidenceIDs(),
	}
	if age > parameters.MaximumAge() {
		output.result = domain.PolicyDecisionRequireReview
		output.reasonCode = "evaluation_stale"
		return output, nil
	}
	output.result = domain.PolicyDecisionAllow
	output.reasonCode = "evaluation_fresh"
	return output, nil
}

func evaluateScoreMultiplier(instance domain.PolicyInstance, input evaluationInput) (evaluatorOutput, error) {
	parameters, ok := instance.ScoreMultiplierParameters()
	if !ok {
		return evaluatorOutput{}, fmt.Errorf("score multiplier parameters missing")
	}
	if input.evaluation == nil {
		return evaluatorOutput{
			result:        domain.PolicyDecisionIndeterminate,
			reasonCode:    "base_score_unavailable",
			missingInputs: []string{"accepted_evaluation", "base_score"},
		}, nil
	}
	effect, err := domain.NewScoreMultiplierPolicyEffect(int(parameters.BasisPoints()))
	if err != nil {
		return evaluatorOutput{}, err
	}
	return evaluatorOutput{
		result:     domain.PolicyDecisionAdjust,
		effects:    []domain.PolicyEffect{effect},
		reasonCode: "score_multiplier_applied",
		inputReferences: []string{
			"evaluation_version:" + string(input.evaluation.ID()),
			"base_score:" + input.evaluation.BaseScore().String(),
			"score_multiplier_basis_points:" + strconv.FormatUint(uint64(parameters.BasisPoints()), 10),
		},
		evidenceIDs: input.evaluation.EvidenceIDs(),
	}, nil
}

func confidenceRank(band domain.ConfidenceBand) int {
	switch band {
	case domain.ConfidenceWeakEvidence:
		return 1
	case domain.ConfidenceVisibleUncertainty:
		return 2
	case domain.ConfidenceWellSupported:
		return 3
	case domain.ConfidenceExceptionalEvidence:
		return 4
	default:
		return 0
	}
}

func multiplyExact(current ExactMultiplier, factorBasisPoints uint32) (ExactMultiplier, error) {
	if current.numerator == 0 || current.denominator == 0 {
		return ExactMultiplier{}, fmt.Errorf("invalid current exact multiplier")
	}
	if factorBasisPoints == 0 {
		return ExactMultiplier{}, fmt.Errorf("score multiplier factor must be positive")
	}
	a := current.numerator
	b := uint64(factorBasisPoints)
	c := current.denominator
	d := uint64(10000)
	g := gcd(a, d)
	a /= g
	d /= g
	g = gcd(b, c)
	b /= g
	c /= g
	hi, numerator := bits.Mul64(a, b)
	if hi != 0 {
		return ExactMultiplier{}, fmt.Errorf("score multiplier numerator overflow")
	}
	hi, denominator := bits.Mul64(c, d)
	if hi != 0 {
		return ExactMultiplier{}, fmt.Errorf("score multiplier denominator overflow")
	}
	g = gcd(numerator, denominator)
	return ExactMultiplier{numerator: numerator / g, denominator: denominator / g}, nil
}

func ratioLessThanBasisPoints(value ExactMultiplier, basisPoints uint32) bool {
	left := new(big.Int).Mul(new(big.Int).SetUint64(value.numerator), big.NewInt(10000))
	right := new(big.Int).Mul(new(big.Int).SetUint64(value.denominator), new(big.Int).SetUint64(uint64(basisPoints)))
	return left.Cmp(right) < 0
}

func ratioGreaterThanBasisPoints(value ExactMultiplier, basisPoints uint32) bool {
	left := new(big.Int).Mul(new(big.Int).SetUint64(value.numerator), big.NewInt(10000))
	right := new(big.Int).Mul(new(big.Int).SetUint64(value.denominator), new(big.Int).SetUint64(uint64(basisPoints)))
	return left.Cmp(right) > 0
}

func gcd(a, b uint64) uint64 {
	for b != 0 {
		a, b = b, a%b
	}
	if a == 0 {
		return 1
	}
	return a
}
