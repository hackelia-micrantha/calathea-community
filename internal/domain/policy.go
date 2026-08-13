package domain

import "time"

// PolicyException is the immutable maintainer-authorized deviation record defined
// by RFC 0007.
//
// #25 establishes its identity/reference foundation. #27 adds the complete
// exception contract (permitted deviation, selector/workflow scope, evaluator
// compatibility, evidence/provenance, effective/expiry bounds, use limits,
// supersession/revocation and related-decision references) and the only public
// constructor. Keeping construction unavailable here prevents an incomplete
// exception from being treated as valid authority.
type PolicyException struct {
	id                 PolicyExceptionID
	policySetVersionID PolicySetVersionID
	policyID           PolicyID
	projectID          ProjectID
	actor              Actor
	rationale          string
	createdAt          time.Time
}

func (e PolicyException) ID() PolicyExceptionID                  { return e.id }
func (e PolicyException) PolicySetVersionID() PolicySetVersionID { return e.policySetVersionID }
func (e PolicyException) PolicyID() PolicyID                     { return e.policyID }
func (e PolicyException) ProjectID() ProjectID                   { return e.projectID }
func (e PolicyException) Actor() Actor                           { return e.actor }
func (e PolicyException) Rationale() string                      { return e.rationale }
func (e PolicyException) CreatedAt() time.Time                   { return e.createdAt }

// PolicyExceptionApplication records that an already-authorized exception was
// validly used in one operation.
//
// #27 adds the complete applicability checks and the only public constructor so
// scope, expiry, revocation, and use limits are evaluated before a durable
// application record can exist.
type PolicyExceptionApplication struct {
	id          PolicyExceptionApplicationID
	exceptionID PolicyExceptionID
	projectID   ProjectID
	operationID OperationID
	appliedAt   time.Time
}

func (a PolicyExceptionApplication) ID() PolicyExceptionApplicationID { return a.id }
func (a PolicyExceptionApplication) ExceptionID() PolicyExceptionID   { return a.exceptionID }
func (a PolicyExceptionApplication) ProjectID() ProjectID             { return a.projectID }
func (a PolicyExceptionApplication) OperationID() OperationID         { return a.operationID }
func (a PolicyExceptionApplication) AppliedAt() time.Time             { return a.appliedAt }
