package domain

import "fmt"

const (
	EvaluationSemanticVersionV1       = "evaluation.v1"
	BaseScoreFormulaSemanticVersionV1 = "evaluation.base_score.v1"
)

// AxisScore is an ordinal evaluation-axis value in the closed interval [1, 5].
type AxisScore uint8

func NewAxisScore(value int) (AxisScore, error) {
	if value < 1 || value > 5 {
		return 0, fmt.Errorf("evaluation axis score %d is outside 1-5", value)
	}
	return AxisScore(value), nil
}

func (s AxisScore) Valid() bool { return s >= 1 && s <= 5 }
func (s AxisScore) Int() int    { return int(s) }

// AxisAssessment couples an ordinal axis value to its required rationale.
type AxisAssessment struct {
	value     AxisScore
	rationale string
}

func NewAxisAssessment(value int, rationale string) (AxisAssessment, error) {
	score, err := NewAxisScore(value)
	if err != nil {
		return AxisAssessment{}, err
	}
	if err := requireText("axis rationale", rationale); err != nil {
		return AxisAssessment{}, err
	}
	return AxisAssessment{value: score, rationale: rationale}, nil
}

func (a AxisAssessment) Value() AxisScore   { return a.value }
func (a AxisAssessment) Rationale() string { return a.rationale }

// EvaluationAxes is the complete RFC 0001 assessment surface.
type EvaluationAxes struct {
	impact        AxisAssessment
	effort        AxisAssessment
	riskReduction AxisAssessment
	optionality   AxisAssessment
	urgency       AxisAssessment
	futurePaths   []string
}

func NewEvaluationAxes(impact, effort, riskReduction, optionality, urgency AxisAssessment, futurePaths []string) (EvaluationAxes, error) {
	axes := EvaluationAxes{
		impact:        impact,
		effort:        effort,
		riskReduction: riskReduction,
		optionality:   optionality,
		urgency:       urgency,
		futurePaths:   cloneStrings(futurePaths),
	}
	if err := axes.validate(); err != nil {
		return EvaluationAxes{}, err
	}
	return axes, nil
}

func (a EvaluationAxes) Impact() AxisAssessment        { return a.impact }
func (a EvaluationAxes) Effort() AxisAssessment        { return a.effort }
func (a EvaluationAxes) RiskReduction() AxisAssessment { return a.riskReduction }
func (a EvaluationAxes) Optionality() AxisAssessment   { return a.optionality }
func (a EvaluationAxes) Urgency() AxisAssessment       { return a.urgency }
func (a EvaluationAxes) FuturePaths() []string         { return cloneStrings(a.futurePaths) }

func (a EvaluationAxes) validate() error {
	assessments := []struct {
		name       string
		assessment AxisAssessment
	}{
		{name: "impact", assessment: a.impact},
		{name: "effort", assessment: a.effort},
		{name: "risk reduction", assessment: a.riskReduction},
		{name: "optionality", assessment: a.optionality},
		{name: "urgency", assessment: a.urgency},
	}
	for _, item := range assessments {
		if !item.assessment.value.Valid() {
			return fmt.Errorf("%s score is outside 1-5", item.name)
		}
		if err := requireText(item.name+" rationale", item.assessment.rationale); err != nil {
			return err
		}
	}
	seen := make(map[string]struct{}, len(a.futurePaths))
	for _, path := range a.futurePaths {
		if err := requireText("optionality future path", path); err != nil {
			return err
		}
		if _, exists := seen[path]; exists {
			return fmt.Errorf("optionality future path %q is duplicated", path)
		}
		seen[path] = struct{}{}
	}
	if a.optionality.value > 1 && len(a.futurePaths) == 0 {
		return fmt.Errorf("optionality above 1 requires at least one named future path")
	}
	return nil
}

// BaseScore is the exact rational result of the v1 evaluation formula.
type BaseScore struct {
	numerator   uint64
	denominator uint64
}

func CalculateBaseScore(axes EvaluationAxes) (BaseScore, error) {
	if err := axes.validate(); err != nil {
		return BaseScore{}, err
	}
	numerator := uint64(axes.impact.value) *
		uint64(axes.riskReduction.value) *
		uint64(axes.optionality.value) *
		uint64(axes.urgency.value)
	denominator := uint64(axes.effort.value)
	divisor := greatestCommonDivisor(numerator, denominator)
	return BaseScore{numerator: numerator / divisor, denominator: denominator / divisor}, nil
}

func (s BaseScore) Numerator() uint64   { return s.numerator }
func (s BaseScore) Denominator() uint64 { return s.denominator }

func (s BaseScore) String() string {
	if s.denominator == 1 {
		return fmt.Sprintf("%d", s.numerator)
	}
	return fmt.Sprintf("%d/%d", s.numerator, s.denominator)
}

func greatestCommonDivisor(a, b uint64) uint64 {
	for b != 0 {
		a, b = b, a%b
	}
	if a == 0 {
		return 1
	}
	return a
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	return append([]string(nil), values...)
}
