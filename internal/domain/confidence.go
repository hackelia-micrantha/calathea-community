package domain

import "fmt"

// Confidence stores an ordinal evidence-quality indicator as exact basis points
// over [0, 1]. 10000 basis points equals 1.0000. It is not a probability.
type Confidence struct {
	basisPoints uint16
	rationale   string
}

func NewConfidence(basisPoints int, rationale string) (Confidence, error) {
	if basisPoints < 0 || basisPoints > 10000 {
		return Confidence{}, fmt.Errorf("confidence basis points %d are outside 0-10000", basisPoints)
	}
	if err := requireText("confidence rationale", rationale); err != nil {
		return Confidence{}, err
	}
	return Confidence{basisPoints: uint16(basisPoints), rationale: rationale}, nil
}

func (c Confidence) BasisPoints() uint16 { return c.basisPoints }
func (c Confidence) Rationale() string   { return c.rationale }
func (c Confidence) DecimalString() string {
	return fmt.Sprintf("%d.%04d", c.basisPoints/10000, c.basisPoints%10000)
}

func (c Confidence) valid() bool {
	return c.basisPoints <= 10000 && requireText("confidence rationale", c.rationale) == nil
}

type ConfidenceBand string

const (
	ConfidenceWeakEvidence        ConfidenceBand = "weak_evidence"
	ConfidenceVisibleUncertainty  ConfidenceBand = "visible_uncertainty"
	ConfidenceWellSupported       ConfidenceBand = "well_supported"
	ConfidenceExceptionalEvidence ConfidenceBand = "exceptional_evidence"
)

func (b ConfidenceBand) Valid() bool {
	switch b {
	case ConfidenceWeakEvidence, ConfidenceVisibleUncertainty, ConfidenceWellSupported, ConfidenceExceptionalEvidence:
		return true
	default:
		return false
	}
}

func (c Confidence) Band() ConfidenceBand {
	switch {
	case c.basisPoints < 4000:
		return ConfidenceWeakEvidence
	case c.basisPoints < 7000:
		return ConfidenceVisibleUncertainty
	case c.basisPoints < 9000:
		return ConfidenceWellSupported
	default:
		return ConfidenceExceptionalEvidence
	}
}
