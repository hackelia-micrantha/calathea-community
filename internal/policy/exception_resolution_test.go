package policy

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/hackelia-micrantha/calathea-community/internal/domain"
)

func exceptionableReviewPolicySet(t *testing.T, exceptionablePolicyIDs ...domain.PolicyID) domain.PolicySetVersion {
	t.Helper()
	wanted := make(map[domain.PolicyID]bool, len(exceptionablePolicyIDs))
	for _, id := range exceptionablePolicyIDs {
		wanted[id] = true
	}
	instances := baselineInstances(t)
	for i, instance := range instances {
		if !wanted[instance.PolicyID()] {
			continue
		}
		instances[i] = mustPolicyInstance(t, domain.PolicyInstanceInput{
			ID:                   instance.ID(),
			PolicyID:             instance.PolicyID(),
			EvaluatorType:        instance.EvaluatorType(),
			EvaluatorVersion:     instance.EvaluatorVersion(),
			Phase:                instance.Phase(),
			EffectClass:          instance.EffectClass(),
			SubjectType:          instance.SubjectType(),
			RequiredInputs:       instance.RequiredInputs(),
			MissingInputBehavior: instance.MissingInputBehavior(),
			Priority:             instance.Priority(),
			Exceptionability:     domain.PolicyExceptionableWithReview,
			Parameters:           instance.Parameters(),
			Rationale:            instance.Rationale(),
		})
	}
	set, err := domain.NewPolicySetVersion(
		"policy-set-review-v1",
		"policy-set",
		policyTestTime(),
		instances...,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateForActivation(set); err != nil {
		t.Fatal(err)
	}
	return set
}

func reviewComposition(t *testing.T, set domain.PolicySetVersion, lifecycle domain.LifecycleState, confidence int, evidenceAsOf time.Time) (Composition, []domain.PolicyDecision) {
	t.Helper()
	return reviewCompositionForOperation(t, set, lifecycle, confidence, evidenceAsOf, "operation-review")
}

func reviewCompositionForOperation(t *testing.T, set domain.PolicySetVersion, lifecycle domain.LifecycleState, confidence int, evidenceAsOf time.Time, operationID domain.OperationID) (Composition, []domain.PolicyDecision) {
	t.Helper()
	subject := projectSubject(t)
	evaluation := testEvaluation(t, confidence, evidenceAsOf, "evidence-review")
	req := ComposeBaselineRequest{PolicySet: set, OperationID: operationID, Subject: subject}
	for _, instance := range set.Instances() {
		r := request(t, set, instance.ID(), subject)
		r.DecisionID = domain.PolicyDecisionID("decision-" + string(instance.ID()))
		r.OperationID = req.OperationID
		r.Context.LifecycleState = &lifecycle
		r.Context.Evaluation = &evaluation
		r.Context.AsOf = policyTestTime()
		decision, err := Evaluate(r)
		if err != nil {
			t.Fatalf("Evaluate(%s): %v", instance.ID(), err)
		}
		req.Decisions = append(req.Decisions, decision)
	}
	return mustCompose(t, req), req.Decisions
}

func decisionByPolicy(t *testing.T, decisions []domain.PolicyDecision, policyID domain.PolicyID) domain.PolicyDecision {
	t.Helper()
	for _, decision := range decisions {
		if decision.PolicyID() == policyID {
			return decision
		}
	}
	t.Fatalf("decision for policy %q not found", policyID)
	return domain.PolicyDecision{}
}

func reviewException(t *testing.T, set domain.PolicySetVersion, decision domain.PolicyDecision, maximumUses int) domain.PolicyException {
	t.Helper()
	return reviewExceptionWithDecisionBinding(t, set, decision, maximumUses, true)
}

func reviewExceptionWithDecisionBinding(t *testing.T, set domain.PolicySetVersion, decision domain.PolicyDecision, maximumUses int, bindDecision bool) domain.PolicyException {
	t.Helper()
	actor, err := domain.NewMaintainerActor("maintainer-review")
	if err != nil {
		t.Fatal(err)
	}
	related := decision.ID()
	var relatedDecisionID *domain.PolicyDecisionID
	if bindDecision {
		relatedDecisionID = &related
	}
	exception, err := domain.NewPolicyException(domain.PolicyExceptionInput{
		ID:                         "exception-review",
		PolicySetVersionID:         set.ID(),
		PolicyID:                   decision.PolicyID(),
		PolicyInstanceID:           decision.PolicyInstanceID(),
		EvaluatorVersion:           decision.EvaluatorVersion(),
		ConfigurationSchemaVersion: decision.ConfigurationSchemaVersion(),
		Workflow:                   decision.Workflow(),
		Phase:                      decision.Phase(),
		Subject:                    decision.Subject(),
		Deviation:                  domain.PolicyExceptionSatisfyReview,
		Actor:                      actor,
		Rationale:                  "maintainer reviewed the public-safe test evidence",
		EvidenceIDs:                []domain.EvidenceReferenceID{"exception-evidence"},
		CreatedAt:                  policyTestTime(),
		EffectiveAt:                policyTestTime(),
		ExpiresAt:                  policyTestTime().Add(24 * time.Hour),
		MaximumUses:                maximumUses,
		RelatedDecisionID:          relatedDecisionID,
	})
	if err != nil {
		t.Fatal(err)
	}
	return exception
}

func reviewActor(t *testing.T) domain.Actor {
	t.Helper()
	actor, err := domain.NewMaintainerActor("maintainer-review")
	if err != nil {
		t.Fatal(err)
	}
	return actor
}

func resolutionRequest(set domain.PolicySetVersion, composition Composition, decision domain.PolicyDecision, exception domain.PolicyException, actor domain.Actor) ResolveReviewWithExceptionRequest {
	return ResolveReviewWithExceptionRequest{
		Composition:   composition,
		PolicySet:     set,
		Decision:      decision,
		Exception:     exception,
		ApplicationID: "application-review",
		Actor:         actor,
		At:            policyTestTime().Add(time.Minute),
	}
}

func TestResolveReviewWithExceptionAllowsOnlyAfterFullValidation(t *testing.T) {
	set := exceptionableReviewPolicySet(t, PolicyConfidence)
	composition, decisions := reviewComposition(t, set, domain.LifecycleApproved, 1000, policyTestTime())
	decision := decisionByPolicy(t, decisions, PolicyConfidence)
	if composition.Outcome() != CompositionReviewNeeded {
		t.Fatalf("original outcome=%q", composition.Outcome())
	}
	exception := reviewException(t, set, decision, 1)

	resolution, err := ResolveReviewWithException(resolutionRequest(set, composition, decision, exception, reviewActor(t)))
	if err != nil {
		t.Fatal(err)
	}
	if resolution.EffectiveOutcome() != CompositionAllowed {
		t.Fatalf("effective outcome=%q, want allowed", resolution.EffectiveOutcome())
	}
	if resolution.ReasonCode() != ReviewSatisfiedByExceptionReason ||
		resolution.TargetDecisionID() != decision.ID() ||
		resolution.ExceptionID() != exception.ID() ||
		resolution.Application().PolicyDecisionID() != decision.ID() {
		t.Fatalf("resolution attribution incomplete: %#v", resolution)
	}
	if resolution.OriginalComposition().Outcome() != CompositionReviewNeeded {
		t.Fatal("resolution rewrote original review outcome")
	}
}

func TestResolveReviewExceptionCannotEraseHardDenial(t *testing.T) {
	set := exceptionableReviewPolicySet(t, PolicyConfidence)
	composition, decisions := reviewComposition(t, set, domain.LifecycleCandidate, 1000, policyTestTime())
	decision := decisionByPolicy(t, decisions, PolicyConfidence)
	resolution, err := ResolveReviewWithException(resolutionRequest(set, composition, decision, reviewException(t, set, decision, 1), reviewActor(t)))
	if err != nil {
		t.Fatal(err)
	}
	if resolution.EffectiveOutcome() != CompositionDenied || resolution.OriginalComposition().Outcome() != CompositionDenied {
		t.Fatalf("review exception relaxed hard denial: original=%q effective=%q", resolution.OriginalComposition().Outcome(), resolution.EffectiveOutcome())
	}
}

func TestResolveOneOfTwoReviewsKeepsOtherReviewVisible(t *testing.T) {
	set := exceptionableReviewPolicySet(t, PolicyConfidence, PolicyFreshness)
	composition, decisions := reviewComposition(t, set, domain.LifecycleApproved, 1000, policyTestTime().AddDate(0, 0, -120))
	confidence := decisionByPolicy(t, decisions, PolicyConfidence)
	freshness := decisionByPolicy(t, decisions, PolicyFreshness)
	if confidence.Result() != domain.PolicyDecisionRequireReview || freshness.Result() != domain.PolicyDecisionRequireReview {
		t.Fatalf("fixture did not create two reviews")
	}
	resolution, err := ResolveReviewWithException(resolutionRequest(set, composition, confidence, reviewException(t, set, confidence, 1), reviewActor(t)))
	if err != nil {
		t.Fatal(err)
	}
	if resolution.EffectiveOutcome() != CompositionReviewNeeded {
		t.Fatalf("resolving one review hid the other: %q", resolution.EffectiveOutcome())
	}
}

func TestResolveReviewRejectsDefaultNonExceptionablePolicy(t *testing.T) {
	set := baselinePolicySet(t)
	composition, decisions := reviewComposition(t, set, domain.LifecycleApproved, 1000, policyTestTime())
	decision := decisionByPolicy(t, decisions, PolicyConfidence)
	exception := reviewException(t, set, decision, 1)
	_, err := ResolveReviewWithException(resolutionRequest(set, composition, decision, exception, reviewActor(t)))
	var useErr *domain.PolicyExceptionUseError
	if !errors.As(err, &useErr) || useErr.Code != "non_exceptionable" {
		t.Fatalf("error=%v, want non_exceptionable", err)
	}
}

func TestResolveReviewRejectsNonReviewTargetBeforeAuthorityUse(t *testing.T) {
	set := exceptionableReviewPolicySet(t, PolicyConfidence)
	composition, decisions := reviewComposition(t, set, domain.LifecycleApproved, 1000, policyTestTime())
	lifecycle := decisionByPolicy(t, decisions, PolicyLifecycleEligibility)
	exception := reviewException(t, set, decisionByPolicy(t, decisions, PolicyConfidence), 1)
	_, err := ResolveReviewWithException(resolutionRequest(set, composition, lifecycle, exception, reviewActor(t)))
	var resolutionErr *ReviewExceptionResolutionError
	if !errors.As(err, &resolutionErr) || resolutionErr.Code != "not_review_required" {
		t.Fatalf("error=%v, want not_review_required", err)
	}
}

func TestResolveReviewHistoricalRetrySurvivesLaterRevocation(t *testing.T) {
	set := exceptionableReviewPolicySet(t, PolicyConfidence)
	composition, decisions := reviewComposition(t, set, domain.LifecycleApproved, 1000, policyTestTime())
	decision := decisionByPolicy(t, decisions, PolicyConfidence)
	exception := reviewException(t, set, decision, 2)
	actor := reviewActor(t)
	firstReq := resolutionRequest(set, composition, decision, exception, actor)
	first, err := ResolveReviewWithException(firstReq)
	if err != nil {
		t.Fatal(err)
	}
	revocation, err := domain.NewPolicyExceptionRevocation(
		"revocation-review",
		exception.ID(),
		actor,
		"later review invalidated future uses",
		first.Application().AppliedAt().Add(time.Minute),
	)
	if err != nil {
		t.Fatal(err)
	}
	retryReq := firstReq
	retryReq.ApplicationID = "different-retry-id"
	retryReq.At = revocation.RevokedAt().Add(time.Minute)
	retryReq.PriorApplications = []domain.PolicyExceptionApplication{first.Application()}
	retryReq.Revocations = []domain.PolicyExceptionRevocation{revocation}
	retry, err := ResolveReviewWithException(retryReq)
	if err != nil {
		t.Fatal(err)
	}
	if retry.Application().ID() != first.Application().ID() ||
		retry.EffectiveOutcome() != first.EffectiveOutcome() {
		t.Fatalf("historical retry changed after revocation")
	}

}

func TestResolveReviewLaterRevocationBlocksNewOperation(t *testing.T) {
	set := exceptionableReviewPolicySet(t, PolicyConfidence)
	firstComposition, firstDecisions := reviewCompositionForOperation(
		t, set, domain.LifecycleApproved, 1000, policyTestTime(), "operation-review-1",
	)
	firstDecision := decisionByPolicy(t, firstDecisions, PolicyConfidence)
	exception := reviewExceptionWithDecisionBinding(t, set, firstDecision, 2, false)
	actor := reviewActor(t)
	firstReq := resolutionRequest(set, firstComposition, firstDecision, exception, actor)
	firstReq.ApplicationID = "application-review-1"
	first, err := ResolveReviewWithException(firstReq)
	if err != nil {
		t.Fatal(err)
	}
	revocation, err := domain.NewPolicyExceptionRevocation(
		"revocation-review",
		exception.ID(),
		actor,
		"later review invalidated future uses",
		first.Application().AppliedAt().Add(time.Minute),
	)
	if err != nil {
		t.Fatal(err)
	}

	secondComposition, secondDecisions := reviewCompositionForOperation(
		t, set, domain.LifecycleApproved, 1000, policyTestTime(), "operation-review-2",
	)
	secondDecision := decisionByPolicy(t, secondDecisions, PolicyConfidence)
	secondReq := resolutionRequest(set, secondComposition, secondDecision, exception, actor)
	secondReq.ApplicationID = "application-review-2"
	secondReq.At = revocation.RevokedAt().Add(time.Minute)
	secondReq.PriorApplications = []domain.PolicyExceptionApplication{first.Application()}
	secondReq.Revocations = []domain.PolicyExceptionRevocation{revocation}
	_, err = ResolveReviewWithException(secondReq)
	var useErr *domain.PolicyExceptionUseError
	if !errors.As(err, &useErr) || useErr.Code != "revoked" {
		t.Fatalf("error=%v, want revoked", err)
	}
}

func TestResolveReviewRejectsWrongCompositionTarget(t *testing.T) {
	set := exceptionableReviewPolicySet(t, PolicyConfidence)
	composition, decisions := reviewComposition(t, set, domain.LifecycleApproved, 1000, policyTestTime())
	decision := decisionByPolicy(t, decisions, PolicyConfidence)
	exception := reviewException(t, set, decision, 1)
	req := resolutionRequest(set, composition, decision, exception, reviewActor(t))
	req.Composition.operationID = "foreign-operation"
	_, err := ResolveReviewWithException(req)
	var resolutionErr *ReviewExceptionResolutionError
	if !errors.As(err, &resolutionErr) || resolutionErr.Code != "decision_mismatch" {
		t.Fatalf("error=%v, want decision_mismatch", err)
	}
}

func TestResolveReviewHistoryOrderIsNonSemantic(t *testing.T) {
	set := exceptionableReviewPolicySet(t, PolicyConfidence)
	composition, decisions := reviewComposition(t, set, domain.LifecycleApproved, 1000, policyTestTime())
	decision := decisionByPolicy(t, decisions, PolicyConfidence)
	exception := reviewException(t, set, decision, 3)
	actor := reviewActor(t)

	firstReq := resolutionRequest(set, composition, decision, exception, actor)
	first, err := ResolveReviewWithException(firstReq)
	if err != nil {
		t.Fatal(err)
	}

	// Retry the same exact operation through two caller orderings. The canonical
	// #30 validator sorts retained history before interpreting it.
	rev1, _ := domain.NewPolicyExceptionRevocation("rev-a", exception.ID(), actor, "future a", first.Application().AppliedAt().Add(2*time.Hour))
	rev2, _ := domain.NewPolicyExceptionRevocation("rev-b", exception.ID(), actor, "future b", first.Application().AppliedAt().Add(3*time.Hour))
	a := firstReq
	a.ApplicationID = "retry-a"
	a.At = first.Application().AppliedAt().Add(time.Hour)
	a.PriorApplications = []domain.PolicyExceptionApplication{first.Application()}
	a.Revocations = []domain.PolicyExceptionRevocation{rev1, rev2}
	b := a
	b.Revocations = []domain.PolicyExceptionRevocation{rev2, rev1}
	left, err := ResolveReviewWithException(a)
	if err != nil {
		t.Fatal(err)
	}
	right, err := ResolveReviewWithException(b)
	if err != nil {
		t.Fatal(err)
	}
	if left.Application().ID() != right.Application().ID() ||
		left.EffectiveOutcome() != right.EffectiveOutcome() ||
		!reflect.DeepEqual(left.OriginalComposition().Steps(), right.OriginalComposition().Steps()) {
		t.Fatal("caller history order changed semantic resolution")
	}
}

func TestResolveReviewCopiesOriginalCompositionTrace(t *testing.T) {
	set := exceptionableReviewPolicySet(t, PolicyConfidence)
	composition, decisions := reviewComposition(t, set, domain.LifecycleApproved, 1000, policyTestTime())
	decision := decisionByPolicy(t, decisions, PolicyConfidence)
	resolution, err := ResolveReviewWithException(resolutionRequest(set, composition, decision, reviewException(t, set, decision, 1), reviewActor(t)))
	if err != nil {
		t.Fatal(err)
	}
	copyValue := resolution.OriginalComposition()
	copyValue.steps[0].Interpretation = "tampered-in-test"
	if resolution.OriginalComposition().steps[0].Interpretation == "tampered-in-test" {
		t.Fatal("resolution returned mutable original composition trace")
	}
}
