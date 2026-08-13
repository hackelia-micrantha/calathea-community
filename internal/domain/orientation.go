package domain

import "time"

// Placement is an orientation recommendation, never a lifecycle state.
type Placement string

const (
	PlacementNow   Placement = "now"
	PlacementNext  Placement = "next"
	PlacementLater Placement = "later"
	PlacementKill  Placement = "kill"
)

func (p Placement) Valid() bool {
	switch p {
	case PlacementNow, PlacementNext, PlacementLater, PlacementKill:
		return true
	default:
		return false
	}
}

// PlacementRecommendation is one immutable project recommendation within a run.
type PlacementRecommendation struct {
	projectID ProjectID
	placement Placement
}

func NewPlacementRecommendation(projectID ProjectID, placement Placement) (PlacementRecommendation, error) {
	if err := requireIdentifier("project id", string(projectID)); err != nil {
		return PlacementRecommendation{}, err
	}
	if !placement.Valid() {
		return PlacementRecommendation{}, errInvalidPlacement(placement)
	}
	return PlacementRecommendation{projectID: projectID, placement: placement}, nil
}

func (r PlacementRecommendation) ProjectID() ProjectID { return r.projectID }
func (r PlacementRecommendation) Placement() Placement { return r.placement }

// OrientationRun is the immutable deterministic recommendation record defined by
// RFC 0000/RFC 0002.
//
// #25 establishes the identity/reference foundation needed by later work: the
// exact policy-set version, accepted evaluation-version references, per-policy
// decision references, operation identity, recommendations, and trace are retained
// as part of the record shape. #28 adds the remaining required run metadata
// (planning horizon, considered subjects, semantic/schema versions, diagnostics,
// imported/observed input references and replay identities) and the only public
// constructor. Keeping construction unavailable here prevents an incomplete run
// from being mistaken for a valid durable OrientationRun.
type OrientationRun struct {
	id                   OrientationRunID
	portfolioID          PortfolioID
	policySetVersionID   PolicySetVersionID
	evaluationVersionIDs []EvaluationVersionID
	policyDecisionIDs    []PolicyDecisionID
	operationID          OperationID
	recommendations      []PlacementRecommendation
	trace                OperationTrace
	createdAt            time.Time
}

func (r OrientationRun) ID() OrientationRunID                   { return r.id }
func (r OrientationRun) PortfolioID() PortfolioID               { return r.portfolioID }
func (r OrientationRun) PolicySetVersionID() PolicySetVersionID { return r.policySetVersionID }
func (r OrientationRun) EvaluationVersionIDs() []EvaluationVersionID {
	return cloneEvaluationVersionIDs(r.evaluationVersionIDs)
}
func (r OrientationRun) PolicyDecisionIDs() []PolicyDecisionID {
	return clonePolicyDecisionIDs(r.policyDecisionIDs)
}
func (r OrientationRun) OperationID() OperationID { return r.operationID }
func (r OrientationRun) Recommendations() []PlacementRecommendation {
	return cloneRecommendations(r.recommendations)
}
func (r OrientationRun) Trace() OperationTrace {
	return OperationTrace{operationID: r.trace.operationID, entries: cloneTraceEntries(r.trace.entries)}
}
func (r OrientationRun) CreatedAt() time.Time { return r.createdAt }

func cloneEvaluationVersionIDs(values []EvaluationVersionID) []EvaluationVersionID {
	if len(values) == 0 {
		return nil
	}
	return append([]EvaluationVersionID(nil), values...)
}

func clonePolicyDecisionIDs(values []PolicyDecisionID) []PolicyDecisionID {
	if len(values) == 0 {
		return nil
	}
	return append([]PolicyDecisionID(nil), values...)
}

func cloneRecommendations(values []PlacementRecommendation) []PlacementRecommendation {
	if len(values) == 0 {
		return nil
	}
	return append([]PlacementRecommendation(nil), values...)
}
