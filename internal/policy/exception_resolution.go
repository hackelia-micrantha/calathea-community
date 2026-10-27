package policy

import (
	"fmt"
	"time"

	"github.com/hackelia-micrantha/calathea-community/internal/domain"
)

const ReviewSatisfiedByExceptionReason = "review_satisfied_by_exception"

// ReviewExceptionResolution records the deterministic effect of one fully
// validated exception application at a review boundary. It never rewrites the
// original PolicyDecision or Composition.
type ReviewExceptionResolution struct {
	original         Composition
	application      domain.PolicyExceptionApplication
	exceptionID      domain.PolicyExceptionID
	targetDecisionID domain.PolicyDecisionID
	effectiveOutcome CompositionOutcome
	reasonCode       string
}

func (r ReviewExceptionResolution) OriginalComposition() Composition {
	return cloneComposition(r.original)
}
func (r ReviewExceptionResolution) Application() domain.PolicyExceptionApplication {
	return r.application
}
func (r ReviewExceptionResolution) ExceptionID() domain.PolicyExceptionID {
	return r.exceptionID
}
func (r ReviewExceptionResolution) TargetDecisionID() domain.PolicyDecisionID {
	return r.targetDecisionID
}
func (r ReviewExceptionResolution) EffectiveOutcome() CompositionOutcome {
	return r.effectiveOutcome
}
func (r ReviewExceptionResolution) ReasonCode() string { return r.reasonCode }

// ResolveReviewWithExceptionRequest contains complete retained authority/history.
// A raw PolicyExceptionApplication is deliberately not accepted as permission.
type ResolveReviewWithExceptionRequest struct {
	Composition       Composition
	PolicySet         domain.PolicySetVersion
	Decision          domain.PolicyDecision
	Exception         domain.PolicyException
	ApplicationID     domain.PolicyExceptionApplicationID
	Actor             domain.Actor
	At                time.Time
	PriorApplications []domain.PolicyExceptionApplication
	Revocations       []domain.PolicyExceptionRevocation
}

// ReviewExceptionResolutionError reports composition-boundary incoherence.
// PolicyExceptionUseError from the canonical domain validator is returned
// unchanged so callers retain its stable authorization/history code.
type ReviewExceptionResolutionError struct {
	Code   string
	Detail string
}

func (e *ReviewExceptionResolutionError) Error() string { return e.Code + ": " + e.Detail }

func reviewResolutionError(code, detail string) error {
	return &ReviewExceptionResolutionError{Code: code, Detail: detail}
}

// ResolveReviewWithException validates the exact target review and delegates all
// exception scope/time/revocation/use/idempotency checks to the canonical #30
// domain validator. The returned application still requires atomic persistence
// by the later application boundary.
func ResolveReviewWithException(req ResolveReviewWithExceptionRequest) (ReviewExceptionResolution, error) {
	c := req.Composition
	d := req.Decision
	if c.PolicySetVersionID() == "" || c.PolicySetVersionID() != req.PolicySet.ID() ||
		c.OperationID() == "" || c.Subject().ID() == "" {
		return ReviewExceptionResolution{}, reviewResolutionError("composition_mismatch", "composition does not identify the supplied policy set/subject/operation")
	}
	if d.ID() == "" || d.PolicySetVersionID() != c.PolicySetVersionID() ||
		d.OperationID() != c.OperationID() || d.Subject() != c.Subject() {
		return ReviewExceptionResolution{}, reviewResolutionError("decision_mismatch", "target decision does not belong to the exact composition context")
	}
	if d.Result() != domain.PolicyDecisionRequireReview || d.EffectClass() != domain.PolicyEffectReviewRequired {
		return ReviewExceptionResolution{}, reviewResolutionError("not_review_required", "only an explicit require_review decision can be satisfied at this boundary")
	}

	found := false
	for _, step := range c.steps {
		if step.DecisionID != d.ID() {
			continue
		}
		if found {
			return ReviewExceptionResolution{}, reviewResolutionError("composition_mismatch", "target decision appears more than once in composition trace")
		}
		found = true
		if step.PolicyID != d.PolicyID() || step.InstanceID != d.PolicyInstanceID() ||
			step.Phase != d.Phase() || step.Priority != d.Priority() ||
			step.Result != d.Result() || step.ReasonCode != d.ReasonCode() ||
			step.Interpretation != "review_required" {
			return ReviewExceptionResolution{}, reviewResolutionError("composition_mismatch", "target decision differs from retained composition trace")
		}
	}
	if !found {
		return ReviewExceptionResolution{}, reviewResolutionError("decision_not_composed", fmt.Sprintf("decision %q is absent from composition trace", d.ID()))
	}

	application, err := domain.ValidateAndApplyPolicyException(domain.PolicyExceptionUseRequest{
		ID:                req.ApplicationID,
		Exception:         req.Exception,
		PolicySet:         req.PolicySet,
		Decision:          d,
		Actor:             req.Actor,
		OperationID:       c.OperationID(),
		Subject:           c.Subject(),
		At:                req.At,
		PriorApplications: req.PriorApplications,
		Revocations:       req.Revocations,
	})
	if err != nil {
		return ReviewExceptionResolution{}, err
	}

	return ReviewExceptionResolution{
		original:         cloneComposition(c),
		application:      application,
		exceptionID:      req.Exception.ID(),
		targetDecisionID: d.ID(),
		effectiveOutcome: outcomeWithResolvedReview(c, d.ID()),
		reasonCode:       ReviewSatisfiedByExceptionReason,
	}, nil
}

func cloneComposition(c Composition) Composition {
	copyValue := c
	copyValue.steps = append([]CompositionStep(nil), c.steps...)
	copyValue.scoreMultiplier = c.ScoreMultiplier()
	return copyValue
}

func outcomeWithResolvedReview(c Composition, target domain.PolicyDecisionID) CompositionOutcome {
	var hasHardAllow, hasDeny, hasExclude, hasReview, hasFailure bool
	for _, step := range c.steps {
		if step.DecisionID == target && step.Interpretation == "review_required" {
			continue
		}
		switch step.Interpretation {
		case "hard_allow":
			hasHardAllow = true
		case "hard_deny", "missing_input_deny":
			hasDeny = true
		case "missing_input_exclude_subject":
			hasExclude = true
		case "review_required", "missing_input_require_review":
			hasReview = true
		case "missing_input_fail_operation":
			hasFailure = true
		}
	}
	switch {
	case hasFailure:
		return CompositionFailed
	case hasDeny:
		return CompositionDenied
	case hasExclude:
		return CompositionExcluded
	case hasReview:
		return CompositionReviewNeeded
	case hasHardAllow:
		return CompositionAllowed
	default:
		return CompositionNotApplicable
	}
}
