package domain

import (
	"fmt"
	"time"
)

// FreshnessMetadata retains neutral timing inputs without prescribing a decay
// policy. Later orientation policy may classify freshness but cannot mutate the
// evaluation version.
type FreshnessMetadata struct {
	evidenceAsOf time.Time
}

func NewFreshnessMetadata(evidenceAsOf time.Time) (FreshnessMetadata, error) {
	if evidenceAsOf.IsZero() {
		return FreshnessMetadata{}, errZeroTime("evaluation evidence as-of time")
	}
	return FreshnessMetadata{evidenceAsOf: evidenceAsOf}, nil
}

func (m FreshnessMetadata) EvidenceAsOf() time.Time { return m.evidenceAsOf }

// EvaluationDerivation records whether an accepted evaluation was authored
// directly by the maintainer or promoted from a retained draft/source.
type EvaluationDerivation string

const (
	EvaluationDerivationAuthored      EvaluationDerivation = "authored"
	EvaluationDerivationPromotedDraft EvaluationDerivation = "promoted_draft"
)

func (d EvaluationDerivation) Valid() bool {
	return d == EvaluationDerivationAuthored || d == EvaluationDerivationPromotedDraft
}

// EvaluationVersionInput contains all inputs required to create one canonical
// immutable accepted evaluation version under RFC 0001.
type EvaluationVersionInput struct {
	ID                     EvaluationVersionID
	Evaluation             Evaluation
	ProjectVersion         ProjectVersion
	EvaluatedAt            time.Time
	PlanningHorizon        string
	Freshness              FreshnessMetadata
	Axes                   EvaluationAxes
	Confidence             Confidence
	Derivation             EvaluationDerivation
	EvidenceIDs            []EvidenceReferenceID
	SemanticVersion        string
	FormulaSemanticVersion string
	AcceptedBy             Actor
	AcceptedAt             time.Time
	Supersedes             *EvaluationVersion
}

// EvaluationVersion is the canonical immutable accepted assessment record.
type EvaluationVersion struct {
	id                     EvaluationVersionID
	evaluationID           EvaluationID
	projectVersionID       ProjectVersionID
	evaluatedAt            time.Time
	planningHorizon        string
	freshness              FreshnessMetadata
	axes                   EvaluationAxes
	confidence             Confidence
	derivation             EvaluationDerivation
	baseScore              BaseScore
	evidenceIDs            []EvidenceReferenceID
	semanticVersion        string
	formulaSemanticVersion string
	acceptedBy             Actor
	acceptedAt             time.Time
	supersedes             *EvaluationVersionID
}

func NewEvaluationVersion(input EvaluationVersionInput) (EvaluationVersion, error) {
	if err := requireIdentifier("evaluation version id", string(input.ID)); err != nil {
		return EvaluationVersion{}, err
	}
	if err := requireIdentifier("evaluation id", string(input.Evaluation.ID())); err != nil {
		return EvaluationVersion{}, err
	}
	if err := requireIdentifier("project version id", string(input.ProjectVersion.ID())); err != nil {
		return EvaluationVersion{}, err
	}
	if input.Evaluation.ProjectID() != input.ProjectVersion.ProjectID() {
		return EvaluationVersion{}, fmt.Errorf("evaluation %q and project version %q refer to different projects", input.Evaluation.ID(), input.ProjectVersion.ID())
	}
	if input.EvaluatedAt.IsZero() {
		return EvaluationVersion{}, errZeroTime("evaluation time")
	}
	if err := requireText("planning horizon", input.PlanningHorizon); err != nil {
		return EvaluationVersion{}, err
	}
	if input.Freshness.evidenceAsOf.IsZero() {
		return EvaluationVersion{}, errZeroTime("evaluation evidence as-of time")
	}
	if input.Freshness.evidenceAsOf.After(input.EvaluatedAt) {
		return EvaluationVersion{}, fmt.Errorf("evaluation evidence as-of time must not be after evaluation time")
	}
	if err := input.Axes.validate(); err != nil {
		return EvaluationVersion{}, err
	}
	if !input.Confidence.valid() {
		return EvaluationVersion{}, fmt.Errorf("confidence is invalid or lacks rationale")
	}
	if !input.Derivation.Valid() {
		return EvaluationVersion{}, fmt.Errorf("invalid evaluation derivation %q", input.Derivation)
	}
	if input.Derivation == EvaluationDerivationPromotedDraft && len(input.EvidenceIDs) == 0 {
		return EvaluationVersion{}, fmt.Errorf("promoted evaluation draft requires at least one source evidence reference")
	}
	for _, evidenceID := range input.EvidenceIDs {
		if err := requireIdentifier("evaluation evidence reference id", string(evidenceID)); err != nil {
			return EvaluationVersion{}, err
		}
	}
	if input.SemanticVersion != EvaluationSemanticVersionV1 {
		return EvaluationVersion{}, fmt.Errorf("unsupported evaluation semantic version %q", input.SemanticVersion)
	}
	if input.FormulaSemanticVersion != BaseScoreFormulaSemanticVersionV1 {
		return EvaluationVersion{}, fmt.Errorf("unsupported base-score formula semantic version %q", input.FormulaSemanticVersion)
	}
	if !input.AcceptedBy.IsMaintainer() {
		return EvaluationVersion{}, errMaintainerAuthority("evaluation acceptance")
	}
	if input.AcceptedAt.IsZero() {
		return EvaluationVersion{}, errZeroTime("evaluation acceptance time")
	}
	if input.AcceptedAt.Before(input.EvaluatedAt) {
		return EvaluationVersion{}, fmt.Errorf("evaluation acceptance time must not precede evaluation time")
	}
	var supersedes *EvaluationVersionID
	if input.Supersedes != nil {
		if err := requireIdentifier("superseded evaluation version id", string(input.Supersedes.ID())); err != nil {
			return EvaluationVersion{}, err
		}
		if input.Supersedes.ID() == input.ID {
			return EvaluationVersion{}, errSelfReference("evaluation version supersedes")
		}
		if input.Supersedes.EvaluationID() != input.Evaluation.ID() {
			return EvaluationVersion{}, fmt.Errorf("superseded evaluation version %q belongs to evaluation %q, want %q", input.Supersedes.ID(), input.Supersedes.EvaluationID(), input.Evaluation.ID())
		}
		priorID := input.Supersedes.ID()
		supersedes = &priorID
	}
	baseScore, err := CalculateBaseScore(input.Axes)
	if err != nil {
		return EvaluationVersion{}, err
	}
	return EvaluationVersion{
		id:                     input.ID,
		evaluationID:           input.Evaluation.ID(),
		projectVersionID:       input.ProjectVersion.ID(),
		evaluatedAt:            input.EvaluatedAt,
		planningHorizon:        input.PlanningHorizon,
		freshness:              input.Freshness,
		axes:                   cloneEvaluationAxes(input.Axes),
		confidence:             input.Confidence,
		derivation:             input.Derivation,
		baseScore:              baseScore,
		evidenceIDs:            cloneEvidenceIDs(input.EvidenceIDs),
		semanticVersion:        input.SemanticVersion,
		formulaSemanticVersion: input.FormulaSemanticVersion,
		acceptedBy:             input.AcceptedBy,
		acceptedAt:             input.AcceptedAt,
		supersedes:             cloneEvaluationVersionID(supersedes),
	}, nil
}

func (v EvaluationVersion) ID() EvaluationVersionID            { return v.id }
func (v EvaluationVersion) EvaluationID() EvaluationID         { return v.evaluationID }
func (v EvaluationVersion) ProjectVersionID() ProjectVersionID { return v.projectVersionID }
func (v EvaluationVersion) EvaluatedAt() time.Time             { return v.evaluatedAt }
func (v EvaluationVersion) PlanningHorizon() string            { return v.planningHorizon }
func (v EvaluationVersion) Freshness() FreshnessMetadata       { return v.freshness }
func (v EvaluationVersion) Axes() EvaluationAxes               { return cloneEvaluationAxes(v.axes) }
func (v EvaluationVersion) Confidence() Confidence             { return v.confidence }
func (v EvaluationVersion) Derivation() EvaluationDerivation   { return v.derivation }
func (v EvaluationVersion) BaseScore() BaseScore               { return v.baseScore }
func (v EvaluationVersion) EvidenceIDs() []EvidenceReferenceID {
	return cloneEvidenceIDs(v.evidenceIDs)
}
func (v EvaluationVersion) SemanticVersion() string        { return v.semanticVersion }
func (v EvaluationVersion) FormulaSemanticVersion() string { return v.formulaSemanticVersion }
func (v EvaluationVersion) AcceptedBy() Actor              { return v.acceptedBy }
func (v EvaluationVersion) AcceptedAt() time.Time          { return v.acceptedAt }
func (v EvaluationVersion) Supersedes() *EvaluationVersionID {
	return cloneEvaluationVersionID(v.supersedes)
}

func cloneEvaluationVersionID(value *EvaluationVersionID) *EvaluationVersionID {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}

func cloneEvaluationAxes(value EvaluationAxes) EvaluationAxes {
	value.futurePaths = cloneStrings(value.futurePaths)
	return value
}
