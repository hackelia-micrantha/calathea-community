package domain

import "time"

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
// instance to one project in an operation.
//
// #25 establishes the record/reference contract. #27 adds the public constructor,
// evaluator identity/version, typed bounded adjustment effects, missing-input
// semantics, and full applicability validation before these records are emitted.
type PolicyDecision struct {
	id                 PolicyDecisionID
	policySetVersionID PolicySetVersionID
	policyID           PolicyID
	projectID          ProjectID
	operationID        OperationID
	result             PolicyDecisionResult
	reasonCode         string
	evidenceIDs        []EvidenceReferenceID
	createdAt          time.Time
}

func (d PolicyDecision) ID() PolicyDecisionID                   { return d.id }
func (d PolicyDecision) PolicySetVersionID() PolicySetVersionID { return d.policySetVersionID }
func (d PolicyDecision) PolicyID() PolicyID                     { return d.policyID }
func (d PolicyDecision) ProjectID() ProjectID                   { return d.projectID }
func (d PolicyDecision) OperationID() OperationID               { return d.operationID }
func (d PolicyDecision) Result() PolicyDecisionResult           { return d.result }
func (d PolicyDecision) ReasonCode() string                     { return d.reasonCode }
func (d PolicyDecision) EvidenceIDs() []EvidenceReferenceID     { return cloneEvidenceIDs(d.evidenceIDs) }
func (d PolicyDecision) CreatedAt() time.Time                   { return d.createdAt }
