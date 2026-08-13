package domain

import "time"

// LifecycleState is independent from orientation placement.
type LifecycleState string

const (
	LifecycleCandidate LifecycleState = "candidate"
	LifecycleProposed  LifecycleState = "proposed"
	LifecycleApproved  LifecycleState = "approved"
	LifecycleActive    LifecycleState = "active"
	LifecyclePaused    LifecycleState = "paused"
	LifecycleCompleted LifecycleState = "completed"
	LifecycleCancelled LifecycleState = "cancelled"
	LifecycleArchived  LifecycleState = "archived"
)

func (s LifecycleState) Valid() bool {
	switch s {
	case LifecycleCandidate, LifecycleProposed, LifecycleApproved, LifecycleActive,
		LifecyclePaused, LifecycleCompleted, LifecycleCancelled, LifecycleArchived:
		return true
	default:
		return false
	}
}

// IsDirectLifecycleTransitionAllowed checks whether an edge exists in RFC 0006.
// It does not prove evidence/readiness preconditions for conditional edges. Archived
// restoration is deliberately excluded because its target must match the archive
// decision's recorded prior effective state.
func IsDirectLifecycleTransitionAllowed(from, to LifecycleState) bool {
	switch from {
	case LifecycleCandidate:
		return to == LifecycleProposed || to == LifecycleCancelled
	case LifecycleProposed:
		return to == LifecycleApproved || to == LifecycleCandidate || to == LifecycleCancelled
	case LifecycleApproved:
		return to == LifecycleActive || to == LifecycleCancelled || to == LifecycleArchived
	case LifecycleActive:
		return to == LifecyclePaused || to == LifecycleCompleted || to == LifecycleCancelled
	case LifecyclePaused:
		return to == LifecycleActive || to == LifecycleCompleted || to == LifecycleCancelled
	case LifecycleCompleted:
		return to == LifecycleActive || to == LifecycleArchived
	case LifecycleCancelled:
		return to == LifecycleApproved || to == LifecycleActive || to == LifecycleArchived
	default:
		return false
	}
}

// LifecycleDecisionKind distinguishes bootstrap/registration history from later
// business transitions. Corrections remain separate records when implemented.
type LifecycleDecisionKind string

const (
	LifecycleDecisionRegistration LifecycleDecisionKind = "registration"
	LifecycleDecisionBootstrap    LifecycleDecisionKind = "bootstrap"
	LifecycleDecisionTransition   LifecycleDecisionKind = "transition"
)

// LifecycleDecision is the immutable authoritative record behind the current
// lifecycle projection.
type LifecycleDecision struct {
	id                   LifecycleDecisionID
	projectID            ProjectID
	projectVersionID     *ProjectVersionID
	kind                 LifecycleDecisionKind
	hasPriorState        bool
	priorState           LifecycleState
	requestedState       LifecycleState
	actor                Actor
	reasonCode           string
	rationale            string
	skippedIntakeContext string
	operationID          OperationID
	decidedAt            time.Time
}

// NewRegistrationLifecycleDecision registers a new project as candidate. Normal
// registration never silently promotes a project to approved.
func NewRegistrationLifecycleDecision(id LifecycleDecisionID, projectID ProjectID, actor Actor, operationID OperationID, decidedAt time.Time) (LifecycleDecision, error) {
	if err := validateLifecycleDecisionBase(id, projectID, actor, operationID, decidedAt); err != nil {
		return LifecycleDecision{}, err
	}
	return LifecycleDecision{
		id:             id,
		projectID:      projectID,
		kind:           LifecycleDecisionRegistration,
		requestedState: LifecycleCandidate,
		actor:          actor,
		reasonCode:     "registration",
		operationID:    operationID,
		decidedAt:      decidedAt,
	}, nil
}

// NewApprovedBootstrapLifecycleDecision is the explicit UC-01 migration
// convenience for an existing portfolio. It binds approval to the exact current
// ProjectVersion and records why normal candidate → proposed → approved intake was
// skipped.
func NewApprovedBootstrapLifecycleDecision(id LifecycleDecisionID, projectID ProjectID, projectVersion ProjectVersion, actor Actor, operationID OperationID, decidedAt time.Time, rationale, skippedIntakeContext string) (LifecycleDecision, error) {
	if err := validateLifecycleDecisionBase(id, projectID, actor, operationID, decidedAt); err != nil {
		return LifecycleDecision{}, err
	}
	if err := requireIdentifier("project version id", string(projectVersion.ID())); err != nil {
		return LifecycleDecision{}, err
	}
	if projectVersion.ProjectID() != projectID {
		return LifecycleDecision{}, errProjectVersionMismatch(projectID, projectVersion.ID())
	}
	if err := requireText("bootstrap rationale", rationale); err != nil {
		return LifecycleDecision{}, err
	}
	if err := requireText("skipped intake context", skippedIntakeContext); err != nil {
		return LifecycleDecision{}, err
	}
	projectVersionID := projectVersion.ID()
	return LifecycleDecision{
		id:                   id,
		projectID:            projectID,
		projectVersionID:     cloneProjectVersionID(&projectVersionID),
		kind:                 LifecycleDecisionBootstrap,
		requestedState:       LifecycleApproved,
		actor:                actor,
		reasonCode:           "existing_portfolio_bootstrap",
		rationale:            rationale,
		skippedIntakeContext: skippedIntakeContext,
		operationID:          operationID,
		decidedAt:            decidedAt,
	}, nil
}

func validateLifecycleDecisionBase(id LifecycleDecisionID, projectID ProjectID, actor Actor, operationID OperationID, decidedAt time.Time) error {
	if err := requireIdentifier("lifecycle decision id", string(id)); err != nil {
		return err
	}
	if err := requireIdentifier("project id", string(projectID)); err != nil {
		return err
	}
	if !actor.IsMaintainer() {
		return errMaintainerAuthority("lifecycle decision")
	}
	if err := requireIdentifier("operation id", string(operationID)); err != nil {
		return err
	}
	if decidedAt.IsZero() {
		return errZeroTime("lifecycle decision time")
	}
	return nil
}

func (d LifecycleDecision) ID() LifecycleDecisionID { return d.id }
func (d LifecycleDecision) ProjectID() ProjectID    { return d.projectID }
func (d LifecycleDecision) ProjectVersionID() *ProjectVersionID {
	return cloneProjectVersionID(d.projectVersionID)
}
func (d LifecycleDecision) Kind() LifecycleDecisionKind    { return d.kind }
func (d LifecycleDecision) HasPriorState() bool            { return d.hasPriorState }
func (d LifecycleDecision) PriorState() LifecycleState     { return d.priorState }
func (d LifecycleDecision) RequestedState() LifecycleState { return d.requestedState }
func (d LifecycleDecision) Actor() Actor                   { return d.actor }
func (d LifecycleDecision) ReasonCode() string             { return d.reasonCode }
func (d LifecycleDecision) Rationale() string              { return d.rationale }
func (d LifecycleDecision) SkippedIntakeContext() string   { return d.skippedIntakeContext }
func (d LifecycleDecision) OperationID() OperationID       { return d.operationID }
func (d LifecycleDecision) DecidedAt() time.Time           { return d.decidedAt }
