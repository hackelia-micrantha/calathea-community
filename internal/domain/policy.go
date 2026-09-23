package domain

import (
	"fmt"
	"sort"
	"time"
)

// PolicyExceptionDeviation is a closed description of the permission granted
// by an exception. It does not mutate an existing policy decision.
type PolicyExceptionDeviation string

const (
	PolicyExceptionAllowDenial   PolicyExceptionDeviation = "allow_denial"
	PolicyExceptionSatisfyReview PolicyExceptionDeviation = "satisfy_review"
)

func (d PolicyExceptionDeviation) Valid() bool {
	return d == PolicyExceptionAllowDenial || d == PolicyExceptionSatisfyReview
}

// PolicyExceptionInput is the complete v0 maintainer-authorized exception
// contract. No field is inferred from currently active policy configuration.
type PolicyExceptionInput struct {
	ID                         PolicyExceptionID
	PolicySetVersionID         PolicySetVersionID
	PolicyID                   PolicyID
	PolicyInstanceID           PolicyInstanceID
	EvaluatorVersion           string
	ConfigurationSchemaVersion string
	Workflow                   PolicyWorkflow
	Phase                      PolicyPhase
	Subject                    PolicySubject
	Deviation                  PolicyExceptionDeviation
	Actor                      Actor
	Rationale                  string
	EvidenceIDs                []EvidenceReferenceID
	CreatedAt                  time.Time
	EffectiveAt                time.Time
	ExpiresAt                  time.Time
	MaximumUses                int
	Supersedes                 *PolicyExceptionID
	RelatedDecisionID          *PolicyDecisionID
}

// PolicyException is an immutable authority record, not a mutable use counter
// or a replacement PolicyDecision. Creation alone does not authorize a use.
type PolicyException struct {
	id                         PolicyExceptionID
	policySetVersionID         PolicySetVersionID
	policyID                   PolicyID
	policyInstanceID           PolicyInstanceID
	evaluatorVersion           string
	configurationSchemaVersion string
	workflow                   PolicyWorkflow
	phase                      PolicyPhase
	subject                    PolicySubject
	deviation                  PolicyExceptionDeviation
	actor                      Actor
	rationale                  string
	evidenceIDs                []EvidenceReferenceID
	createdAt                  time.Time
	effectiveAt                time.Time
	expiresAt                  time.Time
	maximumUses                int
	supersedes                 *PolicyExceptionID
	relatedDecisionID          *PolicyDecisionID
}

func NewPolicyException(in PolicyExceptionInput) (PolicyException, error) {
	for _, f := range []struct{ kind, value string }{
		{"exception id", string(in.ID)},
		{"policy set version id", string(in.PolicySetVersionID)},
		{"policy id", string(in.PolicyID)},
		{"policy instance id", string(in.PolicyInstanceID)},
		{"evaluator version", in.EvaluatorVersion},
		{"configuration schema version", in.ConfigurationSchemaVersion},
	} {
		if err := requireIdentifier(f.kind, f.value); err != nil {
			return PolicyException{}, err
		}
	}
	if !in.Workflow.Valid() || !in.Phase.Valid() || !in.Subject.valid() || !in.Deviation.Valid() {
		return PolicyException{}, fmt.Errorf("invalid exception workflow, phase, subject, or deviation")
	}
	if !in.Actor.IsMaintainer() {
		return PolicyException{}, errMaintainerAuthority("policy exception")
	}
	if err := requireText("policy exception rationale", in.Rationale); err != nil {
		return PolicyException{}, err
	}
	if len(in.EvidenceIDs) == 0 {
		return PolicyException{}, fmt.Errorf("policy exception requires supporting evidence")
	}
	if err := validateEvidenceIDs(in.EvidenceIDs); err != nil {
		return PolicyException{}, err
	}
	if in.CreatedAt.IsZero() || in.EffectiveAt.IsZero() || in.ExpiresAt.IsZero() ||
		in.CreatedAt.After(in.EffectiveAt) || !in.EffectiveAt.Before(in.ExpiresAt) {
		return PolicyException{}, fmt.Errorf("invalid exception creation/effective/expiry times")
	}
	if in.MaximumUses <= 0 {
		return PolicyException{}, fmt.Errorf("exception maximum uses must be positive")
	}
	if in.Supersedes != nil {
		if err := requireIdentifier("superseded exception id", string(*in.Supersedes)); err != nil {
			return PolicyException{}, err
		}
		if *in.Supersedes == in.ID {
			return PolicyException{}, errSelfReference("policy exception supersedes")
		}
	}
	if in.RelatedDecisionID != nil {
		if err := requireIdentifier("related policy decision id", string(*in.RelatedDecisionID)); err != nil {
			return PolicyException{}, err
		}
	}
	return PolicyException{
		id: in.ID, policySetVersionID: in.PolicySetVersionID, policyID: in.PolicyID,
		policyInstanceID: in.PolicyInstanceID, evaluatorVersion: in.EvaluatorVersion,
		configurationSchemaVersion: in.ConfigurationSchemaVersion, workflow: in.Workflow,
		phase: in.Phase, subject: in.Subject, deviation: in.Deviation, actor: in.Actor,
		rationale: in.Rationale, evidenceIDs: cloneEvidenceIDs(in.EvidenceIDs),
		createdAt: in.CreatedAt, effectiveAt: in.EffectiveAt, expiresAt: in.ExpiresAt,
		maximumUses: in.MaximumUses, supersedes: clonePolicyExceptionID(in.Supersedes),
		relatedDecisionID: cloneExceptionDecisionID(in.RelatedDecisionID),
	}, nil
}

func (e PolicyException) ID() PolicyExceptionID                  { return e.id }
func (e PolicyException) PolicySetVersionID() PolicySetVersionID { return e.policySetVersionID }
func (e PolicyException) PolicyID() PolicyID                     { return e.policyID }
func (e PolicyException) PolicyInstanceID() PolicyInstanceID     { return e.policyInstanceID }
func (e PolicyException) EvaluatorVersion() string               { return e.evaluatorVersion }
func (e PolicyException) ConfigurationSchemaVersion() string     { return e.configurationSchemaVersion }
func (e PolicyException) Workflow() PolicyWorkflow               { return e.workflow }
func (e PolicyException) Phase() PolicyPhase                     { return e.phase }
func (e PolicyException) Subject() PolicySubject                 { return e.subject }
func (e PolicyException) ProjectID() ProjectID {
	if e.subject.Type() != PolicySubjectProject {
		return ""
	}
	return ProjectID(e.subject.ID())
}
func (e PolicyException) Deviation() PolicyExceptionDeviation { return e.deviation }
func (e PolicyException) Actor() Actor                        { return e.actor }
func (e PolicyException) Rationale() string                   { return e.rationale }
func (e PolicyException) EvidenceIDs() []EvidenceReferenceID  { return cloneEvidenceIDs(e.evidenceIDs) }
func (e PolicyException) CreatedAt() time.Time                { return e.createdAt }
func (e PolicyException) EffectiveAt() time.Time              { return e.effectiveAt }
func (e PolicyException) ExpiresAt() time.Time                { return e.expiresAt }
func (e PolicyException) MaximumUses() int                    { return e.maximumUses }
func (e PolicyException) Supersedes() *PolicyExceptionID      { return clonePolicyExceptionID(e.supersedes) }
func (e PolicyException) RelatedDecisionID() *PolicyDecisionID {
	return cloneExceptionDecisionID(e.relatedDecisionID)
}

func cloneExceptionDecisionID(id *PolicyDecisionID) *PolicyDecisionID {
	if id == nil {
		return nil
	}
	copy := *id
	return &copy
}

// A revocation is a separate immutable maintainer decision. It does not
// retroactively invalidate an application that was valid when made.
type PolicyExceptionRevocation struct {
	id          PolicyExceptionRevocationID
	exceptionID PolicyExceptionID
	actor       Actor
	rationale   string
	revokedAt   time.Time
}

func NewPolicyExceptionRevocation(id PolicyExceptionRevocationID, exceptionID PolicyExceptionID, actor Actor, rationale string, revokedAt time.Time) (PolicyExceptionRevocation, error) {
	if err := requireIdentifier("exception revocation id", string(id)); err != nil {
		return PolicyExceptionRevocation{}, err
	}
	if err := requireIdentifier("exception id", string(exceptionID)); err != nil {
		return PolicyExceptionRevocation{}, err
	}
	if !actor.IsMaintainer() {
		return PolicyExceptionRevocation{}, errMaintainerAuthority("exception revocation")
	}
	if err := requireText("revocation rationale", rationale); err != nil {
		return PolicyExceptionRevocation{}, err
	}
	if revokedAt.IsZero() {
		return PolicyExceptionRevocation{}, errZeroTime("exception revocation time")
	}
	return PolicyExceptionRevocation{id: id, exceptionID: exceptionID, actor: actor, rationale: rationale, revokedAt: revokedAt}, nil
}

func (r PolicyExceptionRevocation) ID() PolicyExceptionRevocationID { return r.id }
func (r PolicyExceptionRevocation) ExceptionID() PolicyExceptionID  { return r.exceptionID }
func (r PolicyExceptionRevocation) Actor() Actor                    { return r.actor }
func (r PolicyExceptionRevocation) Rationale() string               { return r.rationale }
func (r PolicyExceptionRevocation) RevokedAt() time.Time            { return r.revokedAt }

// Application is produced only by ValidateAndApplyPolicyException. A retry with
// the same operation, subject, target decision and actor returns its old record.
type PolicyExceptionApplication struct {
	id               PolicyExceptionApplicationID
	exceptionID      PolicyExceptionID
	subject          PolicySubject
	operationID      OperationID
	policyDecisionID PolicyDecisionID
	actor            Actor
	appliedAt        time.Time
}

func (a PolicyExceptionApplication) ID() PolicyExceptionApplicationID { return a.id }
func (a PolicyExceptionApplication) ExceptionID() PolicyExceptionID   { return a.exceptionID }
func (a PolicyExceptionApplication) Subject() PolicySubject           { return a.subject }
func (a PolicyExceptionApplication) ProjectID() ProjectID {
	if a.subject.Type() != PolicySubjectProject {
		return ""
	}
	return ProjectID(a.subject.ID())
}
func (a PolicyExceptionApplication) OperationID() OperationID           { return a.operationID }
func (a PolicyExceptionApplication) PolicyDecisionID() PolicyDecisionID { return a.policyDecisionID }
func (a PolicyExceptionApplication) Actor() Actor                       { return a.actor }
func (a PolicyExceptionApplication) AppliedAt() time.Time               { return a.appliedAt }

// The application boundary must atomically persist the returned application
// alongside the operation identity and use the complete retained history.
// This pure function cannot reserve usage or serialize concurrent writers.
type PolicyExceptionUseRequest struct {
	ID                PolicyExceptionApplicationID
	Exception         PolicyException
	PolicySet         PolicySetVersion
	Decision          PolicyDecision
	Actor             Actor
	OperationID       OperationID
	Subject           PolicySubject
	At                time.Time
	PriorApplications []PolicyExceptionApplication
	Revocations       []PolicyExceptionRevocation
}

// PolicyExceptionUseError has a stable code so a caller can distinguish
// invalid history from ordinary ineligibility. Never return partial authority.
type PolicyExceptionUseError struct {
	Code   string
	Detail string
}

func (e *PolicyExceptionUseError) Error() string { return e.Code + ": " + e.Detail }
func exceptionUseError(code, detail string) error {
	return &PolicyExceptionUseError{Code: code, Detail: detail}
}

func ValidateAndApplyPolicyException(req PolicyExceptionUseRequest) (PolicyExceptionApplication, error) {
	fail := func(code, detail string) (PolicyExceptionApplication, error) {
		return PolicyExceptionApplication{}, exceptionUseError(code, detail)
	}
	e := req.Exception
	if err := requireIdentifier("exception application id", string(req.ID)); err != nil {
		return fail("invalid_request", err.Error())
	}
	if err := requireIdentifier("operation id", string(req.OperationID)); err != nil {
		return fail("invalid_request", err.Error())
	}
	if req.At.IsZero() || !req.Actor.IsMaintainer() || !req.Subject.valid() {
		return fail("invalid_request", "time, maintainer actor and typed subject are required")
	}
	if e.id == "" || e.policySetVersionID != req.PolicySet.ID() {
		return fail("policy_mismatch", "unknown exception or different policy set version")
	}

	var instance *PolicyInstance
	for _, candidate := range req.PolicySet.Instances() {
		if candidate.ID() == e.policyInstanceID {
			c := candidate
			instance = &c
			break
		}
	}
	if instance == nil {
		return fail("policy_mismatch", "exception policy instance is absent")
	}
	p := *instance
	if p.PolicyID() != e.policyID || p.EvaluatorVersion() != e.evaluatorVersion ||
		p.ConfigurationSchemaVersion() != e.configurationSchemaVersion ||
		p.Workflow() != e.workflow || p.Phase() != e.phase || p.SubjectType() != e.subject.Type() ||
		e.subject != req.Subject {
		return fail("policy_mismatch", "exception policy identity, version, phase or scope differs")
	}
	// These evaluators protect system invariants regardless of a misconfigured
	// exceptionability flag on an otherwise valid policy instance.
	switch p.EvaluatorType() {
	case PolicyEvaluatorCapacityLimit, PolicyEvaluatorRequiredEvaluation, PolicyEvaluatorLifecycleEligibility:
		return fail("non_exceptionable", "system-invariant evaluator cannot be bypassed")
	}
	if p.Exceptionability() == PolicyNotExceptionable {
		return fail("non_exceptionable", "policy is not exceptionable")
	}
	if p.Exceptionability() == PolicyExceptionableWithConstraints {
		return fail("unsupported_constraints", "v0 has no structured constraint contract for this evaluator")
	}
	if p.Exceptionability() != PolicyExceptionableWithReview {
		return fail("non_exceptionable", "unsupported exceptionability mode")
	}
	d := req.Decision
	if d.ID() == "" || d.PolicySetVersionID() != e.policySetVersionID ||
		d.PolicyID() != e.policyID || d.PolicyInstanceID() != e.policyInstanceID ||
		d.EvaluatorType() != p.EvaluatorType() || d.EffectClass() != p.EffectClass() ||
		d.MissingInputBehavior() != p.MissingInputBehavior() ||
		d.EvaluatorVersion() != e.evaluatorVersion || d.ConfigurationSchemaVersion() != e.configurationSchemaVersion ||
		d.Workflow() != e.workflow || d.Phase() != e.phase ||
		d.Subject() != req.Subject || d.OperationID() != req.OperationID {
		return fail("decision_mismatch", "exception target is not the exact policy decision/operation")
	}
	if d.CreatedAt().After(req.At) || d.CreatedAt().IsZero() {
		return fail("decision_mismatch", "target decision is after application time")
	}
	if e.relatedDecisionID != nil && *e.relatedDecisionID != d.ID() {
		return fail("decision_mismatch", "exception is bound to a different decision")
	}
	if e.deviation == PolicyExceptionAllowDenial && (d.Result() != PolicyDecisionDeny || p.EffectClass() != PolicyEffectHard) ||
		e.deviation == PolicyExceptionSatisfyReview && (d.Result() != PolicyDecisionRequireReview || p.EffectClass() != PolicyEffectReviewRequired) ||
		!e.deviation.Valid() {
		return fail("deviation_mismatch", "exception does not permit the target decision result")
	}

	// Validate the entire supplied history in sorted order. Even a future
	// record must be well-formed; a duplicate operation is not two uses.
	revocations := append([]PolicyExceptionRevocation(nil), req.Revocations...)
	sort.Slice(revocations, func(i, j int) bool {
		if revocations[i].revokedAt.Equal(revocations[j].revokedAt) {
			return revocations[i].id < revocations[j].id
		}
		return revocations[i].revokedAt.Before(revocations[j].revokedAt)
	})
	revIDs := make(map[PolicyExceptionRevocationID]bool, len(revocations))
	for _, r := range revocations {
		if r.id == "" || r.exceptionID != e.id || !r.actor.IsMaintainer() || r.rationale == "" ||
			r.revokedAt.IsZero() || r.revokedAt.Before(e.createdAt) || revIDs[r.id] {
			return fail("invalid_history", "invalid, foreign, or duplicate revocation")
		}
		revIDs[r.id] = true
	}
	apps := append([]PolicyExceptionApplication(nil), req.PriorApplications...)
	sort.Slice(apps, func(i, j int) bool {
		if apps[i].appliedAt.Equal(apps[j].appliedAt) {
			return apps[i].id < apps[j].id
		}
		return apps[i].appliedAt.Before(apps[j].appliedAt)
	})
	appIDs := make(map[PolicyExceptionApplicationID]bool, len(apps))
	operations := make(map[OperationID]bool, len(apps))
	used := 0
	var retry *PolicyExceptionApplication
	for i := range apps {
		a := apps[i]
		if a.id == "" || appIDs[a.id] || a.exceptionID != e.id || !a.subject.valid() ||
			a.subject != e.subject || a.operationID == "" || a.policyDecisionID == "" ||
			!a.actor.IsMaintainer() || a.appliedAt.Before(e.effectiveAt) || !a.appliedAt.Before(e.expiresAt) ||
			operations[a.operationID] {
			return fail("invalid_history", "invalid, foreign, or duplicate application")
		}
		appIDs[a.id], operations[a.operationID] = true, true
		for _, r := range revocations {
			if !r.revokedAt.After(a.appliedAt) {
				return fail("invalid_history", "application after exception revocation")
			}
		}
		if !a.appliedAt.After(req.At) {
			used++
		}
		if a.operationID == req.OperationID {
			if a.subject != req.Subject || a.policyDecisionID != d.ID() || a.actor != req.Actor {
				return fail("idempotency_conflict", "operation was already applied to a different target or actor")
			}
			copy := a
			retry = &copy
		}
	}
	if len(apps) > e.maximumUses {
		return fail("invalid_history", "retained applications exceed the exception use limit")
	}
	if retry == nil && len(apps) > 0 && req.At.Before(apps[len(apps)-1].appliedAt) {
		return fail("historical_write", "new application predates an existing use")
	}
	if retry != nil {
		if req.At.Before(retry.appliedAt) {
			return fail("idempotency_conflict", "retry time precedes the recorded application")
		}
		return *retry, nil
	}
	if appIDs[req.ID] {
		return fail("idempotency_conflict", "application id belongs to another operation")
	}
	if req.At.Before(e.effectiveAt) || !req.At.Before(e.expiresAt) {
		return fail("not_effective", "exception is outside effective interval")
	}
	for _, r := range revocations {
		if !r.revokedAt.After(req.At) {
			return fail("revoked", "exception was revoked before application")
		}
	}
	if used >= e.maximumUses {
		return fail("uses_exhausted", "exception use limit reached")
	}
	return PolicyExceptionApplication{
		id: req.ID, exceptionID: e.id, subject: req.Subject, operationID: req.OperationID,
		policyDecisionID: d.ID(), actor: req.Actor, appliedAt: req.At,
	}, nil
}
