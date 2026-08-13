package domain

import "time"

// Disposition records the maintainer's decision about an immutable orientation run.
type Disposition string

const (
	DispositionAccepted              Disposition = "accepted"
	DispositionAcceptedWithOverrides Disposition = "accepted_with_overrides"
	DispositionRejected              Disposition = "rejected"
	DispositionDeferred              Disposition = "deferred"
)

func (d Disposition) Valid() bool {
	switch d {
	case DispositionAccepted, DispositionAcceptedWithOverrides, DispositionRejected, DispositionDeferred:
		return true
	default:
		return false
	}
}

func (d Disposition) AcceptsOrientation() bool {
	return d == DispositionAccepted || d == DispositionAcceptedWithOverrides
}

// PlacementOverride is a maintainer-authored replacement of one recommendation.
type PlacementOverride struct {
	projectID   ProjectID
	placement   Placement
	rationale   string
	exceptionID *PolicyExceptionID
}

func NewPlacementOverride(projectID ProjectID, placement Placement, rationale string, exceptionID *PolicyExceptionID) (PlacementOverride, error) {
	override := PlacementOverride{
		projectID:   projectID,
		placement:   placement,
		rationale:   rationale,
		exceptionID: clonePolicyExceptionID(exceptionID),
	}
	if err := validatePlacementOverride(override); err != nil {
		return PlacementOverride{}, err
	}
	return override, nil
}

func validatePlacementOverride(override PlacementOverride) error {
	if err := requireIdentifier("project id", string(override.projectID)); err != nil {
		return err
	}
	if !override.placement.Valid() {
		return errInvalidPlacement(override.placement)
	}
	if err := requireText("override rationale", override.rationale); err != nil {
		return err
	}
	if override.exceptionID != nil {
		if err := requireIdentifier("policy exception id", string(*override.exceptionID)); err != nil {
			return err
		}
	}
	return nil
}

func (o PlacementOverride) ProjectID() ProjectID { return o.projectID }
func (o PlacementOverride) Placement() Placement { return o.placement }
func (o PlacementOverride) Rationale() string    { return o.rationale }
func (o PlacementOverride) ExceptionID() *PolicyExceptionID {
	return clonePolicyExceptionID(o.exceptionID)
}

// OrientationDisposition is a canonical immutable maintainer decision about a run.
type OrientationDisposition struct {
	id          OrientationDispositionID
	runID       OrientationRunID
	disposition Disposition
	actor       Actor
	operationID OperationID
	rationale   string
	overrides   []PlacementOverride
	supersedes  *OrientationDispositionID
	decidedAt   time.Time
}

func NewOrientationDisposition(id OrientationDispositionID, runID OrientationRunID, disposition Disposition, actor Actor, operationID OperationID, rationale string, overrides []PlacementOverride, supersedes *OrientationDispositionID, decidedAt time.Time) (OrientationDisposition, error) {
	if err := requireIdentifier("orientation disposition id", string(id)); err != nil {
		return OrientationDisposition{}, err
	}
	if err := requireIdentifier("orientation run id", string(runID)); err != nil {
		return OrientationDisposition{}, err
	}
	if !disposition.Valid() {
		return OrientationDisposition{}, errInvalidDisposition(disposition)
	}
	if !actor.IsMaintainer() {
		return OrientationDisposition{}, errMaintainerAuthority("orientation disposition")
	}
	if err := requireIdentifier("operation id", string(operationID)); err != nil {
		return OrientationDisposition{}, err
	}
	if decidedAt.IsZero() {
		return OrientationDisposition{}, errZeroTime("orientation disposition time")
	}
	if disposition == DispositionAcceptedWithOverrides && len(overrides) == 0 {
		return OrientationDisposition{}, errOverridesRequired()
	}
	if disposition != DispositionAcceptedWithOverrides && len(overrides) != 0 {
		return OrientationDisposition{}, errOverridesNotAllowed(disposition)
	}
	if disposition != DispositionAccepted {
		if err := requireText("orientation disposition rationale", rationale); err != nil {
			return OrientationDisposition{}, err
		}
	}
	if supersedes != nil {
		if !disposition.AcceptsOrientation() {
			return OrientationDisposition{}, errSupersessionNotAllowed(disposition)
		}
		if err := requireIdentifier("superseded orientation disposition id", string(*supersedes)); err != nil {
			return OrientationDisposition{}, err
		}
		if *supersedes == id {
			return OrientationDisposition{}, errSelfReference("orientation disposition supersedes")
		}
	}
	seen := make(map[ProjectID]struct{}, len(overrides))
	for _, override := range overrides {
		if err := validatePlacementOverride(override); err != nil {
			return OrientationDisposition{}, err
		}
		if _, exists := seen[override.projectID]; exists {
			return OrientationDisposition{}, errDuplicateProjectOverride(override.projectID)
		}
		seen[override.projectID] = struct{}{}
	}
	return OrientationDisposition{
		id:          id,
		runID:       runID,
		disposition: disposition,
		actor:       actor,
		operationID: operationID,
		rationale:   rationale,
		overrides:   cloneOverrides(overrides),
		supersedes:  cloneOrientationDispositionID(supersedes),
		decidedAt:   decidedAt,
	}, nil
}

func (d OrientationDisposition) ID() OrientationDispositionID   { return d.id }
func (d OrientationDisposition) RunID() OrientationRunID        { return d.runID }
func (d OrientationDisposition) Disposition() Disposition       { return d.disposition }
func (d OrientationDisposition) Actor() Actor                   { return d.actor }
func (d OrientationDisposition) OperationID() OperationID       { return d.operationID }
func (d OrientationDisposition) Rationale() string              { return d.rationale }
func (d OrientationDisposition) Overrides() []PlacementOverride { return cloneOverrides(d.overrides) }
func (d OrientationDisposition) Supersedes() *OrientationDispositionID {
	return cloneOrientationDispositionID(d.supersedes)
}
func (d OrientationDisposition) DecidedAt() time.Time { return d.decidedAt }

func cloneOrientationDispositionID(value *OrientationDispositionID) *OrientationDispositionID {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}

func cloneOverrides(values []PlacementOverride) []PlacementOverride {
	if len(values) == 0 {
		return nil
	}
	result := make([]PlacementOverride, len(values))
	for i, value := range values {
		result[i] = PlacementOverride{
			projectID:   value.projectID,
			placement:   value.placement,
			rationale:   value.rationale,
			exceptionID: clonePolicyExceptionID(value.exceptionID),
		}
	}
	return result
}

func clonePolicyExceptionID(value *PolicyExceptionID) *PolicyExceptionID {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}
