package domain

import (
	"errors"
	"testing"
	"time"
)

func exceptionFixture(t *testing.T) (PolicyExceptionUseRequest, PolicyExceptionInput) {
	t.Helper()
	at := testTime()
	actor, err := NewMaintainerActor(ActorID("maintainer"))
	if err != nil {
		t.Fatal(err)
	}
	subject, err := NewProjectPolicySubject(ProjectID("project-1"))
	if err != nil {
		t.Fatal(err)
	}
	instance := mustPolicyInstance(t, PolicyInstanceInput{
		ID:                   PolicyInstanceID("confidence"),
		PolicyID:             PolicyID("orientation.evaluation.confidence"),
		EvaluatorType:        PolicyEvaluatorConfidenceGate,
		EvaluatorVersion:     "1",
		Phase:                PolicyPhaseCandidateAdjustment,
		EffectClass:          PolicyEffectReviewRequired,
		SubjectType:          PolicySubjectProject,
		RequiredInputs:       []PolicyInputKind{PolicyInputAcceptedEvaluation},
		MissingInputBehavior: PolicyMissingInputRequireReview,
		Exceptionability:     PolicyExceptionableWithReview,
		Parameters:           mustConfidenceParameters(t),
		Rationale:            "require evidence review",
	})
	policySet, err := NewPolicySetVersion(PolicySetVersionID("set-v1"), PolicySetID("set"), at.Add(-3*time.Hour), instance)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := NewPolicyInputReference(PolicyInputAcceptedEvaluation, "evaluation-v1")
	if err != nil {
		t.Fatal(err)
	}
	effect, err := NewRequireReviewPolicyEffect("confidence_below_review_threshold")
	if err != nil {
		t.Fatal(err)
	}
	decision, err := NewPolicyDecision(PolicyDecisionInput{
		ID:                   PolicyDecisionID("decision-1"),
		PolicySetVersionID:   policySet.ID(),
		PolicyID:             instance.PolicyID(),
		PolicyInstanceID:     instance.ID(),
		EvaluatorType:        instance.EvaluatorType(),
		EvaluatorVersion:     instance.EvaluatorVersion(),
		Phase:                instance.Phase(),
		EffectClass:          instance.EffectClass(),
		MissingInputBehavior: instance.MissingInputBehavior(),
		Subject:              subject,
		OperationID:          OperationID("operation-1"),
		Result:               PolicyDecisionRequireReview,
		ReasonCode:           "confidence_below_review_threshold",
		RequiredInputs:       []PolicyInputKind{PolicyInputAcceptedEvaluation},
		InputReferences:      []PolicyInputReference{ref},
		Effects:              []PolicyEffect{effect},
		Rationale:            "require evidence review",
		CreatedAt:            at,
	})
	if err != nil {
		t.Fatal(err)
	}
	in := PolicyExceptionInput{
		ID:                         PolicyExceptionID("exception-1"),
		PolicySetVersionID:         policySet.ID(),
		PolicyID:                   instance.PolicyID(),
		PolicyInstanceID:           instance.ID(),
		EvaluatorVersion:           instance.EvaluatorVersion(),
		ConfigurationSchemaVersion: instance.ConfigurationSchemaVersion(),
		Workflow:                   instance.Workflow(),
		Phase:                      instance.Phase(),
		Subject:                    subject,
		Deviation:                  PolicyExceptionSatisfyReview,
		Actor:                      actor,
		Rationale:                  "documented maintainer review exception",
		EvidenceIDs:                []EvidenceReferenceID{"evidence-1"},
		CreatedAt:                  at.Add(-2 * time.Hour),
		EffectiveAt:                at.Add(-time.Hour),
		ExpiresAt:                  at.Add(24 * time.Hour),
		MaximumUses:                1,
	}
	exception, err := NewPolicyException(in)
	if err != nil {
		t.Fatal(err)
	}
	return PolicyExceptionUseRequest{
		ID:          PolicyExceptionApplicationID("application-1"),
		Exception:   exception,
		PolicySet:   policySet,
		Decision:    decision,
		Actor:       actor,
		OperationID: decision.OperationID(),
		Subject:     subject,
		At:          at.Add(time.Hour),
	}, in
}

func mustConfidenceParameters(t *testing.T) PolicyParameters {
	t.Helper()
	p, err := NewConfidenceGateParameters(5000)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func requireExceptionCode(t *testing.T, err error, code string) {
	t.Helper()
	var typed *PolicyExceptionUseError
	if !errors.As(err, &typed) || typed.Code != code {
		t.Fatalf("error = %v; want PolicyExceptionUseError %q", err, code)
	}
}

func TestExceptionConstructionEvidenceAuthorityAndTime(t *testing.T) {
	req, in := exceptionFixture(t)
	evidence := in.EvidenceIDs
	evidence[0] = "changed"
	if got := req.Exception.EvidenceIDs()[0]; got != "evidence-1" {
		t.Fatalf("evidence mutated through constructor input: %q", got)
	}
	copyEvidence := req.Exception.EvidenceIDs()
	copyEvidence[0] = "changed-again"
	if req.Exception.EvidenceIDs()[0] != "evidence-1" {
		t.Fatal("evidence mutated through accessor")
	}
	for _, tc := range []struct {
		name string
		edit func(*PolicyExceptionInput)
	}{
		{"missing evidence", func(i *PolicyExceptionInput) { i.EvidenceIDs = nil }},
		{"zero uses", func(i *PolicyExceptionInput) { i.MaximumUses = 0 }},
		{"before creation", func(i *PolicyExceptionInput) { i.EffectiveAt = i.CreatedAt.Add(-time.Second) }},
		{"no effective duration", func(i *PolicyExceptionInput) { i.ExpiresAt = i.EffectiveAt }},
		{"wrong authority", func(i *PolicyExceptionInput) { i.Actor = Actor{} }},
		{"self supersession", func(i *PolicyExceptionInput) { i.Supersedes = &i.ID }},
		{"invalid related decision", func(i *PolicyExceptionInput) { blank := PolicyDecisionID(""); i.RelatedDecisionID = &blank }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bad := in
			tc.edit(&bad)
			if _, err := NewPolicyException(bad); err == nil {
				t.Fatal("accepted invalid exception")
			}
		})
	}
}

func TestExceptionUseOneShotAndIdempotentRetry(t *testing.T) {
	req, _ := exceptionFixture(t)
	first, err := ValidateAndApplyPolicyException(req)
	if err != nil {
		t.Fatal(err)
	}
	req.PriorApplications = []PolicyExceptionApplication{first}
	req.ID = "unused-retry-id"
	retry, err := ValidateAndApplyPolicyException(req)
	if err != nil || retry.ID() != first.ID() || retry.AppliedAt() != first.AppliedAt() {
		t.Fatalf("retry = %#v, %v; want original immutable application", retry, err)
	}
	req.OperationID = "operation-2"
	req.ID = "application-2"
	req.Decision.operationID = req.OperationID
	req.Decision.id = "decision-2"
	if _, err := ValidateAndApplyPolicyException(req); err == nil {
		t.Fatal("one-shot exception accepted a second operation")
	} else {
		requireExceptionCode(t, err, "uses_exhausted")
	}
}

func TestExceptionMultiUseAndRevocationAsOf(t *testing.T) {
	req, in := exceptionFixture(t)
	in.MaximumUses = 2
	var err error
	req.Exception, err = NewPolicyException(in)
	if err != nil {
		t.Fatal(err)
	}
	first, err := ValidateAndApplyPolicyException(req)
	if err != nil {
		t.Fatal(err)
	}
	req.PriorApplications = []PolicyExceptionApplication{first}
	req.OperationID, req.ID = "operation-2", "application-2"
	req.Decision.operationID, req.Decision.id = req.OperationID, "decision-2"
	req.At = req.At.Add(time.Hour)
	second, err := ValidateAndApplyPolicyException(req)
	if err != nil {
		t.Fatal(err)
	}
	req.PriorApplications = []PolicyExceptionApplication{second, first}
	req.OperationID, req.ID = "operation-3", "application-3"
	req.Decision.operationID, req.Decision.id = req.OperationID, "decision-3"
	if _, err := ValidateAndApplyPolicyException(req); err == nil {
		t.Fatal("accepted third application beyond use limit")
	} else {
		requireExceptionCode(t, err, "uses_exhausted")
	}
	revocation, err := NewPolicyExceptionRevocation("revocation-1", in.ID, in.Actor, "superseded risk decision", second.AppliedAt().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	req.Revocations = []PolicyExceptionRevocation{revocation}
	req.At = revocation.RevokedAt().Add(time.Second)
	if _, err := ValidateAndApplyPolicyException(req); err == nil {
		t.Fatal("accepted new application after revocation")
	} else {
		requireExceptionCode(t, err, "revoked")
	}
	req.OperationID, req.ID = "operation-1", "retry-id"
	req.Decision.operationID, req.Decision.id = req.OperationID, first.PolicyDecisionID()
	req.At = req.At.Add(time.Hour)
	retry, err := ValidateAndApplyPolicyException(req)
	if err != nil || retry.ID() != first.ID() {
		t.Fatalf("later revocation invalidated historical retry: %v", err)
	}
}

func TestExceptionEffectiveIntervalAndTargetBinding(t *testing.T) {
	req, in := exceptionFixture(t)
	req.Decision.createdAt = in.CreatedAt
	req.At = in.EffectiveAt.Add(-time.Nanosecond)
	if _, err := ValidateAndApplyPolicyException(req); err == nil {
		t.Fatal("accepted exception before effective time")
	} else {
		requireExceptionCode(t, err, "not_effective")
	}
	req.At = in.ExpiresAt
	if _, err := ValidateAndApplyPolicyException(req); err == nil {
		t.Fatal("accepted exception at exclusive expiry")
	} else {
		requireExceptionCode(t, err, "not_effective")
	}
	req.At = in.EffectiveAt
	req.Decision.createdAt = in.EffectiveAt.Add(-time.Second)
	if _, err := ValidateAndApplyPolicyException(req); err != nil {
		t.Fatalf("rejected inclusive effective boundary: %v", err)
	}
	req.At = in.EffectiveAt.Add(time.Hour)
	req.Decision.createdAt = req.At.Add(-time.Second)
	req.Subject, _ = NewProjectPolicySubject("different-project")
	if _, err := ValidateAndApplyPolicyException(req); err == nil {
		t.Fatal("accepted wrong subject")
	} else {
		requireExceptionCode(t, err, "policy_mismatch")
	}
}

func TestExceptionBindingRejectsMismatches(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*PolicyExceptionUseRequest)
		code string
	}{
		{"wrong set", func(r *PolicyExceptionUseRequest) { r.Exception.policySetVersionID = "different" }, "policy_mismatch"},
		{"wrong evaluator", func(r *PolicyExceptionUseRequest) { r.Exception.evaluatorVersion = "2" }, "policy_mismatch"},
		{"wrong schema", func(r *PolicyExceptionUseRequest) { r.Exception.configurationSchemaVersion = "2" }, "policy_mismatch"},
		{"wrong workflow", func(r *PolicyExceptionUseRequest) { r.Exception.workflow = "other" }, "policy_mismatch"},
		{"wrong phase", func(r *PolicyExceptionUseRequest) { r.Exception.phase = PolicyPhaseSetConstraints }, "policy_mismatch"},
		{"different decision", func(r *PolicyExceptionUseRequest) {
			r.Decision.id = "other"
			id := PolicyDecisionID("decision-1")
			r.Exception.relatedDecisionID = &id
		}, "decision_mismatch"},
		{"wrong operation", func(r *PolicyExceptionUseRequest) { r.Decision.operationID = "other" }, "decision_mismatch"},
		{"wrong deviation", func(r *PolicyExceptionUseRequest) { r.Exception.deviation = PolicyExceptionAllowDenial }, "deviation_mismatch"},
		{"constraints unsupported", func(r *PolicyExceptionUseRequest) {
			p := r.PolicySet.instances[0]
			p.exceptionability = PolicyExceptionableWithConstraints
			r.PolicySet.instances[0] = p
		}, "unsupported_constraints"},
		{"nonexceptionable", func(r *PolicyExceptionUseRequest) {
			p := r.PolicySet.instances[0]
			p.exceptionability = PolicyNotExceptionable
			r.PolicySet.instances[0] = p
		}, "non_exceptionable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := exceptionFixture(t)
			tc.edit(&req)
			_, err := ValidateAndApplyPolicyException(req)
			requireExceptionCode(t, err, tc.code)
		})
	}
}

func TestExceptionHistoryDeterminismAndConflict(t *testing.T) {
	req, _ := exceptionFixture(t)
	first, err := ValidateAndApplyPolicyException(req)
	if err != nil {
		t.Fatal(err)
	}
	req.PriorApplications = []PolicyExceptionApplication{first, first}
	requireExceptionCode(t, func() error { _, err := ValidateAndApplyPolicyException(req); return err }(), "invalid_history")
	req.PriorApplications = []PolicyExceptionApplication{first}
	req.Actor, _ = NewMaintainerActor("other-actor")
	requireExceptionCode(t, func() error { _, err := ValidateAndApplyPolicyException(req); return err }(), "idempotency_conflict")
	req.Actor = first.Actor()
	req.PriorApplications[0].appliedAt = req.Exception.EffectiveAt().Add(-time.Second)
	requireExceptionCode(t, func() error { _, err := ValidateAndApplyPolicyException(req); return err }(), "invalid_history")
}

func TestExceptionRejectsOverusedOrBackdatedHistory(t *testing.T) {
	req, _ := exceptionFixture(t)
	first, err := ValidateAndApplyPolicyException(req)
	if err != nil {
		t.Fatal(err)
	}
	second := first
	second.id = "application-2"
	second.operationID = "operation-2"
	second.policyDecisionID = "decision-2"
	second.appliedAt = first.appliedAt.Add(time.Minute)
	req.PriorApplications = []PolicyExceptionApplication{second, first}
	requireExceptionCode(t, func() error {
		_, err := ValidateAndApplyPolicyException(req)
		return err
	}(), "invalid_history")

	req.PriorApplications = []PolicyExceptionApplication{first}
	req.OperationID, req.ID = "operation-3", "application-3"
	req.Decision.operationID, req.Decision.id = req.OperationID, "decision-3"
	req.At = first.AppliedAt().Add(-time.Minute)
	requireExceptionCode(t, func() error {
		_, err := ValidateAndApplyPolicyException(req)
		return err
	}(), "historical_write")
}

func TestExceptionRevocationAuthorityAndForeignHistory(t *testing.T) {
	req, _ := exceptionFixture(t)
	if _, err := NewPolicyExceptionRevocation("rev-1", req.Exception.ID(), Actor{}, "invalid", req.At); err == nil {
		t.Fatal("accepted unauthorized revocation")
	}
	revoke, err := NewPolicyExceptionRevocation("rev-1", "different", req.Actor, "wrong scope", req.At)
	if err != nil {
		t.Fatal(err)
	}
	req.Revocations = []PolicyExceptionRevocation{revoke}
	requireExceptionCode(t, func() error { _, err := ValidateAndApplyPolicyException(req); return err }(), "invalid_history")
}

func TestExceptionRejectsInvariantEvenWithMisconfiguredExceptionability(t *testing.T) {
	req, _ := exceptionFixture(t)
	instance := req.PolicySet.instances[0]
	instance.evaluatorType = PolicyEvaluatorCapacityLimit
	instance.effectClass = PolicyEffectHard
	instance.exceptionability = PolicyExceptionableWithReview
	req.PolicySet.instances[0] = instance
	req.Exception.policyID = instance.policyID
	req.Exception.deviation = PolicyExceptionAllowDenial
	req.Decision.result = PolicyDecisionDeny
	req.Decision.effectClass = PolicyEffectHard
	requireExceptionCode(t, func() error { _, err := ValidateAndApplyPolicyException(req); return err }(), "non_exceptionable")
}
