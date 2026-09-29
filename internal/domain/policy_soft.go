package domain

import "fmt"

// PolicySoftCombinatorVersion identifies exact composition semantics for a
// supported typed soft effect. New semantics require a new version.
type PolicySoftCombinatorVersion string

const PolicySoftCombinatorMultiplyV1 PolicySoftCombinatorVersion = "score_multiplier.multiply.v1"

func (v PolicySoftCombinatorVersion) Valid() bool {
	return v == PolicySoftCombinatorMultiplyV1
}

// ScoreMultiplierContract is immutable policy-set metadata for optional score
// multiplier policies. Basis points are exact configuration values; 10000 is
// neutral 1.0. Bounds are explicit rather than hidden kernel calibration.
type ScoreMultiplierContract struct {
	combinator               PolicySoftCombinatorVersion
	perEffectMinBasisPoints  int
	perEffectMaxBasisPoints  int
	cumulativeMinBasisPoints int
	cumulativeMaxBasisPoints int
}

func NewScoreMultiplierContract(combinator PolicySoftCombinatorVersion, perEffectMinBasisPoints, perEffectMaxBasisPoints, cumulativeMinBasisPoints, cumulativeMaxBasisPoints int) (ScoreMultiplierContract, error) {
	c := ScoreMultiplierContract{
		combinator:               combinator,
		perEffectMinBasisPoints:  perEffectMinBasisPoints,
		perEffectMaxBasisPoints:  perEffectMaxBasisPoints,
		cumulativeMinBasisPoints: cumulativeMinBasisPoints,
		cumulativeMaxBasisPoints: cumulativeMaxBasisPoints,
	}
	if err := c.validate(); err != nil {
		return ScoreMultiplierContract{}, err
	}
	return c, nil
}

func (c ScoreMultiplierContract) validate() error {
	if !c.combinator.Valid() {
		return fmt.Errorf("unsupported score multiplier combinator %q", c.combinator)
	}
	if c.perEffectMinBasisPoints <= 0 || c.perEffectMaxBasisPoints <= 0 ||
		c.perEffectMinBasisPoints > c.perEffectMaxBasisPoints {
		return fmt.Errorf("invalid per-effect score multiplier bounds")
	}
	if c.cumulativeMinBasisPoints <= 0 || c.cumulativeMaxBasisPoints <= 0 ||
		c.cumulativeMinBasisPoints > 10000 || c.cumulativeMaxBasisPoints < 10000 ||
		c.cumulativeMinBasisPoints > c.cumulativeMaxBasisPoints {
		return fmt.Errorf("cumulative score multiplier bounds must be positive and include neutral 10000")
	}
	return nil
}

func (c ScoreMultiplierContract) Combinator() PolicySoftCombinatorVersion { return c.combinator }
func (c ScoreMultiplierContract) PerEffectMinBasisPoints() int            { return c.perEffectMinBasisPoints }
func (c ScoreMultiplierContract) PerEffectMaxBasisPoints() int            { return c.perEffectMaxBasisPoints }
func (c ScoreMultiplierContract) CumulativeMinBasisPoints() int           { return c.cumulativeMinBasisPoints }
func (c ScoreMultiplierContract) CumulativeMaxBasisPoints() int           { return c.cumulativeMaxBasisPoints }

func cloneScoreMultiplierContract(value *ScoreMultiplierContract) *ScoreMultiplierContract {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}

func (v PolicySetVersion) ScoreMultiplierContract() (ScoreMultiplierContract, bool) {
	if v.scoreMultiplierContract == nil {
		return ScoreMultiplierContract{}, false
	}
	return *cloneScoreMultiplierContract(v.scoreMultiplierContract), true
}
