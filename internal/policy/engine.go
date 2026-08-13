// Package policy owns deterministic built-in policy evaluation for Calathea.
//
// It depends only on domain contracts and the standard library. Persistence,
// network, provider, CLI, and external-governance concerns remain outside this package.
package policy

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/hackelia-micrantha/calathea-community/internal/domain"
)

const EvaluatorSemanticVersionV1 = "1"

const (
	PolicyLifecycleEligibility domain.PolicyID = "orientation.lifecycle.eligibility"
	PolicyRequiredEvaluation   domain.PolicyID = "orientation.evaluation.required"
	PolicyCapacityNow          domain.PolicyID = "orientation.capacity.now"
	PolicyCapacityNext         domain.PolicyID = "orientation.capacity.next"
	PolicyConfidence           domain.PolicyID = "orientation.evaluation.confidence"
	PolicyFreshness            domain.PolicyID = "orientation.evaluation.freshness"
)

// EvaluationFailure is an operational evaluator failure. It is deliberately
// distinct from a valid PolicyDecision with result indeterminate.
type EvaluationFailure struct {
	PolicyID   domain.PolicyID
	InstanceID domain.PolicyInstanceID
	Code       string
	Err        error
}

func (e *EvaluationFailure) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("policy %q instance %q evaluation failed: %s", e.PolicyID, e.InstanceID, e.Code)
	}
	return fmt.Sprintf("policy %q instance %q evaluation failed: %s: %v", e.PolicyID, e.InstanceID, e.Code, e.Err)
}

func (e *EvaluationFailure) Unwrap() error { return e.Err }

// Context contains operation inputs required by the built-in UC-01 evaluators.
// Nil pointers represent unavailable input and therefore produce an indeterminate
// policy decision when the policy is applicable.
type Context struct {
	Subject        domain.PolicySubject
	LifecycleState *domain.LifecycleState
	Evaluation     *domain.EvaluationVersion
	SelectedCount  *int
	AsOf           time.Time
	EvidenceIDs    []domain.EvidenceReferenceID
}

// Request identifies one exact configured policy instance within an immutable
// policy-set version and supplies operation-owned decision identity/time.
type Request struct {
	DecisionID  domain.PolicyDecisionID
	PolicySet   domain.PolicySetVersion
	InstanceID  domain.PolicyInstanceID
	OperationID domain.OperationID
	CreatedAt   time.Time
	Context     Context
}

type baselineSpec struct {
	policyID       domain.PolicyID
	evaluator      domain.PolicyEvaluatorType
	phase          domain.PolicyPhase
	effectClass    domain.PolicyEffectClass
	subjectType    domain.PolicySubjectType
	requiredInputs []domain.PolicyInputKind
}

var baselineSpecs = map[domain.PolicyID]baselineSpec{
	PolicyLifecycleEligibility: {
		policyID:       PolicyLifecycleEligibility,
		evaluator:      domain.PolicyEvaluatorLifecycleEligibility,
		phase:          domain.PolicyPhaseCandidateEligibility,
		effectClass:    domain.PolicyEffectHard,
		subjectType:    domain.PolicySubjectProject,
		requiredInputs: []domain.PolicyInputKind{domain.PolicyInputLifecycleState},
	},
	PolicyRequiredEvaluation: {
		policyID:       PolicyRequiredEvaluation,
		evaluator:      domain.PolicyEvaluatorRequiredEvaluation,
		phase:          domain.PolicyPhaseCandidateEligibility,
		effectClass:    domain.PolicyEffectHard,
		subjectType:    domain.PolicySubjectProject,
		requiredInputs: []domain.PolicyInputKind{domain.PolicyInputAcceptedEvaluation},
	},
	PolicyCapacityNow: {
		policyID:       PolicyCapacityNow,
		evaluator:      domain.PolicyEvaluatorCapacityLimit,
		phase:          domain.PolicyPhaseSetConstraints,
		effectClass:    domain.PolicyEffectHard,
		subjectType:    domain.PolicySubjectPlacementSet,
		requiredInputs: []domain.PolicyInputKind{domain.PolicyInputSelectedCount},
	},
	PolicyCapacityNext: {
		policyID:       PolicyCapacityNext,
		evaluator:      domain.PolicyEvaluatorCapacityLimit,
		phase:          domain.PolicyPhaseSetConstraints,
		effectClass:    domain.PolicyEffectHard,
		subjectType:    domain.PolicySubjectPlacementSet,
		requiredInputs: []domain.PolicyInputKind{domain.PolicyInputSelectedCount},
	},
	PolicyConfidence: {
		policyID:       PolicyConfidence,
		evaluator:      domain.PolicyEvaluatorConfidenceGate,
		phase:          domain.PolicyPhaseCandidateAdjustment,
		effectClass:    domain.PolicyEffectReviewRequired,
		subjectType:    domain.PolicySubjectProject,
		requiredInputs: []domain.PolicyInputKind{domain.PolicyInputAcceptedEvaluation},
	},
	PolicyFreshness: {
		policyID:       PolicyFreshness,
		evaluator:      domain.PolicyEvaluatorFreshnessRule,
		phase:          domain.PolicyPhaseCandidateAdjustment,
		effectClass:    domain.PolicyEffectReviewRequired,
		subjectType:    domain.PolicySubjectProject,
		requiredInputs: []domain.PolicyInputKind{domain.PolicyInputAcceptedEvaluation, domain.PolicyInputAsOfTime},
	},
}

// ValidateForActivation proves that a PolicySetVersion contains exactly one
// correctly wired instance of every UC-01 baseline policy. It does not perform
// multi-policy composition; that is a separate policy-engine slice.
func ValidateForActivation(version domain.PolicySetVersion) error {
	if strings.TrimSpace(string(version.ID())) == "" {
		return errors.New("policy set version id must not be empty")
	}
	instances := version.Instances()
	if len(instances) == 0 {
		return errors.New("policy set version contains no policy instances")
	}

	seenPolicy := make(map[domain.PolicyID]domain.PolicyInstanceID, len(instances))
	for _, instance := range instances {
		if instance.EvaluatorVersion() != EvaluatorSemanticVersionV1 {
			return fmt.Errorf("policy %q uses unsupported evaluator version %q", instance.PolicyID(), instance.EvaluatorVersion())
		}
		if prior, exists := seenPolicy[instance.PolicyID()]; exists {
			return fmt.Errorf("policy %q appears in both instances %q and %q", instance.PolicyID(), prior, instance.ID())
		}
		seenPolicy[instance.PolicyID()] = instance.ID()

		spec, required := baselineSpecs[instance.PolicyID()]
		if !required {
			return fmt.Errorf("policy %q is not supported by the UC-01 baseline", instance.PolicyID())
		}
		if err := validateInstanceAgainstSpec(instance, spec); err != nil {
			return err
		}
	}

	missing := make([]string, 0)
	for policyID := range baselineSpecs {
		if _, exists := seenPolicy[policyID]; !exists {
			missing = append(missing, string(policyID))
		}
	}
	if len(missing) != 0 {
		sort.Strings(missing)
		return fmt.Errorf("policy set version is missing required baseline policies: %s", strings.Join(missing, ", "))
	}
	return nil
}

func validateInstanceAgainstSpec(instance domain.PolicyInstance, spec baselineSpec) error {
	if instance.EvaluatorType() != spec.evaluator {
		return fmt.Errorf("policy %q uses evaluator %q, want %q", instance.PolicyID(), instance.EvaluatorType(), spec.evaluator)
	}
	if instance.Phase() != spec.phase {
		return fmt.Errorf("policy %q uses phase %q, want %q", instance.PolicyID(), instance.Phase(), spec.phase)
	}
	if instance.EffectClass() != spec.effectClass {
		return fmt.Errorf("policy %q uses effect class %q, want %q", instance.PolicyID(), instance.EffectClass(), spec.effectClass)
	}
	if instance.SubjectType() != spec.subjectType {
		return fmt.Errorf("policy %q uses subject type %q, want %q", instance.PolicyID(), instance.SubjectType(), spec.subjectType)
	}
	if !sameInputKinds(instance.RequiredInputs(), spec.requiredInputs) {
		return fmt.Errorf("policy %q required inputs do not match evaluator contract", instance.PolicyID())
	}
	if instance.PolicyID() == PolicyCapacityNow {
		placement, ok := instance.Parameters().Placement()
		if !ok || placement != domain.PlacementNow {
			return fmt.Errorf("policy %q must constrain now placement", instance.PolicyID())
		}
	}
	if instance.PolicyID() == PolicyCapacityNext {
		placement, ok := instance.Parameters().Placement()
		if !ok || placement != domain.PlacementNext {
			return fmt.Errorf("policy %q must constrain next placement", instance.PolicyID())
		}
	}
	return nil
}

func sameInputKinds(left, right []domain.PolicyInputKind) bool {
	if len(left) != len(right) {
		return false
	}
	leftCopy := append([]domain.PolicyInputKind(nil), left...)
	rightCopy := append([]domain.PolicyInputKind(nil), right...)
	sort.Slice(leftCopy, func(i, j int) bool { return leftCopy[i] < leftCopy[j] })
	sort.Slice(rightCopy, func(i, j int) bool { return rightCopy[i] < rightCopy[j] })
	for i := range leftCopy {
		if leftCopy[i] != rightCopy[i] {
			return false
		}
	}
	return true
}

// Evaluate runs one exact configured policy instance. Valid missing input yields
// an indeterminate PolicyDecision; unsupported evaluator/runtime conditions return
// EvaluationFailure instead. Missing-input behavior is retained on the decision
// and interpreted only by deterministic composition.
func Evaluate(request Request) (domain.PolicyDecision, error) {
	instance, err := findInstance(request.PolicySet, request.InstanceID)
	if err != nil {
		return domain.PolicyDecision{}, err
	}
	if instance.EvaluatorVersion() != EvaluatorSemanticVersionV1 {
		return domain.PolicyDecision{}, failure(instance, "unsupported_evaluator_version", nil)
	}
	if request.Context.Subject.Type() != instance.SubjectType() {
		return makeDecision(request, instance, domain.PolicyDecisionNotApplicable, "subject_type_not_applicable", nil, nil)
	}

	switch instance.EvaluatorType() {
	case domain.PolicyEvaluatorLifecycleEligibility:
		return evaluateLifecycle(request, instance)
	case domain.PolicyEvaluatorRequiredEvaluation:
		return evaluateRequiredEvaluation(request, instance)
	case domain.PolicyEvaluatorCapacityLimit:
		return evaluateCapacity(request, instance)
	case domain.PolicyEvaluatorConfidenceGate:
		return evaluateConfidence(request, instance)
	case domain.PolicyEvaluatorFreshnessRule:
		return evaluateFreshness(request, instance)
	default:
		return domain.PolicyDecision{}, failure(instance, "unsupported_evaluator", nil)
	}
}

func findInstance(version domain.PolicySetVersion, instanceID domain.PolicyInstanceID) (domain.PolicyInstance, error) {
	for _, instance := range version.Instances() {
		if instance.ID() == instanceID {
			return instance, nil
		}
	}
	return domain.PolicyInstance{}, &EvaluationFailure{InstanceID: instanceID, Code: "policy_instance_not_found"}
}

func evaluateLifecycle(request Request, instance domain.PolicyInstance) (domain.PolicyDecision, error) {
	if request.Context.LifecycleState == nil {
		return makeDecision(request, instance, domain.PolicyDecisionIndeterminate, "missing_lifecycle_state", []domain.PolicyInputKind{domain.PolicyInputLifecycleState}, nil)
	}
	state := *request.Context.LifecycleState
	if !state.Valid() {
		return domain.PolicyDecision{}, failure(instance, "invalid_lifecycle_state", fmt.Errorf("state %q is invalid", state))
	}
	allowProposed, _ := instance.Parameters().AllowProposed()
	switch state {
	case domain.LifecycleApproved, domain.LifecycleActive, domain.LifecyclePaused:
		return makeDecision(request, instance, domain.PolicyDecisionAllow, "lifecycle_eligible", nil, nil)
	case domain.LifecycleProposed:
		if allowProposed {
			return makeDecision(request, instance, domain.PolicyDecisionAllow, "proposed_lifecycle_allowed", nil, nil)
		}
		return makeDecision(request, instance, domain.PolicyDecisionDeny, "proposed_lifecycle_not_allowed", nil, nil)
	case domain.LifecycleCandidate, domain.LifecycleCompleted, domain.LifecycleCancelled, domain.LifecycleArchived:
		return makeDecision(request, instance, domain.PolicyDecisionDeny, "lifecycle_ineligible", nil, nil)
	default:
		return domain.PolicyDecision{}, failure(instance, "invalid_lifecycle_state", fmt.Errorf("state %q is unsupported", state))
	}
}

func evaluateRequiredEvaluation(request Request, instance domain.PolicyInstance) (domain.PolicyDecision, error) {
	if request.Context.Evaluation == nil {
		return makeDecision(request, instance, domain.PolicyDecisionIndeterminate, "missing_accepted_evaluation", []domain.PolicyInputKind{domain.PolicyInputAcceptedEvaluation}, nil)
	}
	return makeDecision(request, instance, domain.PolicyDecisionAllow, "accepted_evaluation_present", nil, nil)
}

func evaluateCapacity(request Request, instance domain.PolicyInstance) (domain.PolicyDecision, error) {
	placement, _ := instance.Parameters().Placement()
	if request.Context.Subject.ID() != string(placement) {
		return makeDecision(request, instance, domain.PolicyDecisionNotApplicable, "placement_set_not_applicable", nil, nil)
	}
	if request.Context.SelectedCount == nil {
		return makeDecision(request, instance, domain.PolicyDecisionIndeterminate, "missing_selected_count", []domain.PolicyInputKind{domain.PolicyInputSelectedCount}, nil)
	}
	if *request.Context.SelectedCount < 0 {
		return domain.PolicyDecision{}, failure(instance, "invalid_selected_count", fmt.Errorf("selected count %d must not be negative", *request.Context.SelectedCount))
	}
	maximum, _ := instance.Parameters().Maximum()
	effect, err := domain.NewCapacityPolicyEffect(placement, maximum)
	if err != nil {
		return domain.PolicyDecision{}, failure(instance, "invalid_capacity_effect", err)
	}
	if *request.Context.SelectedCount >= maximum {
		return makeDecision(request, instance, domain.PolicyDecisionDeny, "capacity_exhausted", nil, []domain.PolicyEffect{effect})
	}
	return makeDecision(request, instance, domain.PolicyDecisionAllow, "capacity_available", nil, []domain.PolicyEffect{effect})
}

func evaluateConfidence(request Request, instance domain.PolicyInstance) (domain.PolicyDecision, error) {
	if request.Context.Evaluation == nil {
		return makeDecision(request, instance, domain.PolicyDecisionIndeterminate, "missing_accepted_evaluation", []domain.PolicyInputKind{domain.PolicyInputAcceptedEvaluation}, nil)
	}
	threshold, _ := instance.Parameters().ConfidenceReviewBelowBasisPoints()
	if int(request.Context.Evaluation.Confidence().BasisPoints()) < threshold {
		effect, err := domain.NewRequireReviewPolicyEffect("confidence_below_review_threshold")
		if err != nil {
			return domain.PolicyDecision{}, failure(instance, "invalid_review_effect", err)
		}
		return makeDecision(request, instance, domain.PolicyDecisionRequireReview, "confidence_below_review_threshold", nil, []domain.PolicyEffect{effect})
	}
	return makeDecision(request, instance, domain.PolicyDecisionAllow, "confidence_sufficient", nil, nil)
}

func evaluateFreshness(request Request, instance domain.PolicyInstance) (domain.PolicyDecision, error) {
	if request.Context.Evaluation == nil {
		return makeDecision(request, instance, domain.PolicyDecisionIndeterminate, "missing_accepted_evaluation", []domain.PolicyInputKind{domain.PolicyInputAcceptedEvaluation}, nil)
	}
	if request.Context.AsOf.IsZero() {
		return makeDecision(request, instance, domain.PolicyDecisionIndeterminate, "missing_as_of_time", []domain.PolicyInputKind{domain.PolicyInputAsOfTime}, nil)
	}
	evidenceAsOf := request.Context.Evaluation.Freshness().EvidenceAsOf()
	if request.Context.AsOf.Before(evidenceAsOf) {
		return domain.PolicyDecision{}, failure(instance, "invalid_freshness_context", fmt.Errorf("as-of time precedes evaluation evidence time"))
	}
	maxAgeDays, _ := instance.Parameters().FreshnessMaxAgeDays()
	freshThrough := evidenceAsOf.AddDate(0, 0, maxAgeDays)
	if request.Context.AsOf.After(freshThrough) {
		effect, err := domain.NewRequireReviewPolicyEffect("evaluation_stale")
		if err != nil {
			return domain.PolicyDecision{}, failure(instance, "invalid_review_effect", err)
		}
		return makeDecision(request, instance, domain.PolicyDecisionRequireReview, "evaluation_stale", nil, []domain.PolicyEffect{effect})
	}
	return makeDecision(request, instance, domain.PolicyDecisionAllow, "evaluation_fresh", nil, nil)
}

func makeDecision(request Request, instance domain.PolicyInstance, result domain.PolicyDecisionResult, reasonCode string, missingInputs []domain.PolicyInputKind, effects []domain.PolicyEffect) (domain.PolicyDecision, error) {
	return domain.NewPolicyDecision(domain.PolicyDecisionInput{
		ID:                   request.DecisionID,
		PolicySetVersionID:   request.PolicySet.ID(),
		PolicyID:             instance.PolicyID(),
		PolicyInstanceID:     instance.ID(),
		EvaluatorType:        instance.EvaluatorType(),
		EvaluatorVersion:     instance.EvaluatorVersion(),
		Phase:                instance.Phase(),
		EffectClass:          instance.EffectClass(),
		MissingInputBehavior: instance.MissingInputBehavior(),
		Subject:              request.Context.Subject,
		OperationID:          request.OperationID,
		Result:               result,
		ReasonCode:           reasonCode,
		RequiredInputs:       instance.RequiredInputs(),
		MissingInputs:        missingInputs,
		EvidenceIDs:          request.Context.EvidenceIDs,
		Effects:              effects,
		Priority:             instance.Priority(),
		Rationale:            instance.Rationale(),
		CreatedAt:            request.CreatedAt,
	})
}

func failure(instance domain.PolicyInstance, code string, err error) error {
	return &EvaluationFailure{PolicyID: instance.PolicyID(), InstanceID: instance.ID(), Code: code, Err: err}
}
