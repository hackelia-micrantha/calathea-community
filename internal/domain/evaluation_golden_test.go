package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type evaluationGoldenFile struct {
	FixtureVersion            string                 `json:"fixture_version"`
	EvaluationSemanticVersion string                 `json:"evaluation_semantic_version"`
	FormulaSemanticVersion    string                 `json:"formula_semantic_version"`
	Cases                     []evaluationGoldenCase `json:"cases"`
}

type evaluationGoldenCase struct {
	Name                     string   `json:"name"`
	PlanningHorizon          string   `json:"planning_horizon"`
	Impact                   int      `json:"impact"`
	ImpactRationale          string   `json:"impact_rationale"`
	Effort                   int      `json:"effort"`
	EffortRationale          string   `json:"effort_rationale"`
	RiskReduction            int      `json:"risk_reduction"`
	RiskReductionRationale   string   `json:"risk_reduction_rationale"`
	Optionality              int      `json:"optionality"`
	OptionalityRationale     string   `json:"optionality_rationale"`
	Urgency                  int      `json:"urgency"`
	UrgencyRationale         string   `json:"urgency_rationale"`
	FuturePaths              []string `json:"future_paths"`
	ConfidenceBasisPoints    int      `json:"confidence_basis_points"`
	ConfidenceRationale      string   `json:"confidence_rationale"`
	EvaluatedAt              string   `json:"evaluated_at"`
	EvidenceAsOf             string   `json:"evidence_as_of"`
	ExpectedScoreNumerator   uint64   `json:"expected_score_numerator"`
	ExpectedScoreDenominator uint64   `json:"expected_score_denominator"`
}

type evaluationInvalidFile struct {
	FixtureVersion string                  `json:"fixture_version"`
	Cases          []evaluationInvalidCase `json:"cases"`
}

type evaluationInvalidCase struct {
	Name                  string `json:"name"`
	AxisValue             *int   `json:"axis_value"`
	Rationale             string `json:"rationale"`
	ConfidenceBasisPoints *int   `json:"confidence_basis_points"`
	ConfidenceRationale   string `json:"confidence_rationale"`
}

func decodeFixture(t *testing.T, data []byte, target any) {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
}

func TestEvaluationGoldenScenarios(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "evaluation_golden.json"))
	if err != nil {
		t.Fatalf("read golden fixture: %v", err)
	}
	var fixture evaluationGoldenFile
	decodeFixture(t, data, &fixture)
	if fixture.FixtureVersion != "evaluation-golden.v1" {
		t.Fatalf("fixture version = %q, want evaluation-golden.v1", fixture.FixtureVersion)
	}
	if fixture.EvaluationSemanticVersion != EvaluationSemanticVersionV1 || fixture.FormulaSemanticVersion != BaseScoreFormulaSemanticVersionV1 {
		t.Fatal("golden fixture semantic versions do not match implementation")
	}
	if len(fixture.Cases) < 7 {
		t.Fatalf("golden fixture count = %d, want at least 7", len(fixture.Cases))
	}

	scoresByName := make(map[string]BaseScore, len(fixture.Cases))
	for index, tc := range fixture.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			axes, err := NewEvaluationAxes(
				mustAxis(t, tc.Impact, tc.ImpactRationale),
				mustAxis(t, tc.Effort, tc.EffortRationale),
				mustAxis(t, tc.RiskReduction, tc.RiskReductionRationale),
				mustAxis(t, tc.Optionality, tc.OptionalityRationale),
				mustAxis(t, tc.Urgency, tc.UrgencyRationale),
				tc.FuturePaths,
			)
			if err != nil {
				t.Fatalf("NewEvaluationAxes() error = %v", err)
			}
			confidence, err := NewConfidence(tc.ConfidenceBasisPoints, tc.ConfidenceRationale)
			if err != nil {
				t.Fatalf("NewConfidence() error = %v", err)
			}
			evaluatedAt, err := time.Parse(time.RFC3339, tc.EvaluatedAt)
			if err != nil {
				t.Fatalf("parse evaluated_at: %v", err)
			}
			evidenceAsOf, err := time.Parse(time.RFC3339, tc.EvidenceAsOf)
			if err != nil {
				t.Fatalf("parse evidence_as_of: %v", err)
			}
			freshness, err := NewFreshnessMetadata(evidenceAsOf)
			if err != nil {
				t.Fatalf("NewFreshnessMetadata() error = %v", err)
			}

			projectID := ProjectID(fmt.Sprintf("golden-project-%d", index+1))
			evaluation, err := NewEvaluation(EvaluationID(fmt.Sprintf("golden-evaluation-%d", index+1)), projectID)
			if err != nil {
				t.Fatalf("NewEvaluation() error = %v", err)
			}
			projectVersion := testProjectVersion(t, ProjectVersionID(fmt.Sprintf("golden-project-version-%d", index+1)), projectID)
			version, err := NewEvaluationVersion(EvaluationVersionInput{
				ID:                     EvaluationVersionID(fmt.Sprintf("golden-evaluation-version-%d", index+1)),
				Evaluation:             evaluation,
				ProjectVersion:         projectVersion,
				EvaluatedAt:            evaluatedAt,
				PlanningHorizon:        tc.PlanningHorizon,
				Freshness:              freshness,
				Axes:                   axes,
				Confidence:             confidence,
				Derivation:             EvaluationDerivationAuthored,
				EvidenceIDs:            []EvidenceReferenceID{EvidenceReferenceID(fmt.Sprintf("golden-evidence-%d", index+1))},
				SemanticVersion:        fixture.EvaluationSemanticVersion,
				FormulaSemanticVersion: fixture.FormulaSemanticVersion,
				AcceptedBy:             testMaintainer(t),
				AcceptedAt:             evaluatedAt,
			})
			if err != nil {
				t.Fatalf("NewEvaluationVersion() error = %v", err)
			}
			score := version.BaseScore()
			if score.Numerator() != tc.ExpectedScoreNumerator || score.Denominator() != tc.ExpectedScoreDenominator {
				t.Fatalf("score = %d/%d, want %d/%d", score.Numerator(), score.Denominator(), tc.ExpectedScoreNumerator, tc.ExpectedScoreDenominator)
			}
			if version.Confidence().BasisPoints() != uint16(tc.ConfidenceBasisPoints) {
				t.Fatalf("confidence basis points = %d, want %d", version.Confidence().BasisPoints(), tc.ConfidenceBasisPoints)
			}
			if version.PlanningHorizon() != tc.PlanningHorizon {
				t.Fatalf("planning horizon = %q, want %q", version.PlanningHorizon(), tc.PlanningHorizon)
			}
			if version.Derivation() != EvaluationDerivationAuthored {
				t.Fatalf("derivation = %q, want %q", version.Derivation(), EvaluationDerivationAuthored)
			}
			if tc.Name == "stale strategically important" && evaluatedAt.Sub(version.Freshness().EvidenceAsOf()) < 180*24*time.Hour {
				t.Fatalf("stale scenario age = %s, want at least 180 days", evaluatedAt.Sub(version.Freshness().EvidenceAsOf()))
			}
			scoresByName[tc.Name] = score
		})
	}
	if got, want := scoresByName["materially uncertain"], scoresByName["tie calibration"]; got != want {
		t.Fatalf("expected explicit tie fixture: materially uncertain=%v tie calibration=%v", got, want)
	}
}

func TestEvaluationInvalidFixtures(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "evaluation_invalid.json"))
	if err != nil {
		t.Fatalf("read invalid fixture: %v", err)
	}
	var fixture evaluationInvalidFile
	decodeFixture(t, data, &fixture)
	if fixture.FixtureVersion != "evaluation-invalid.v1" {
		t.Fatalf("fixture version = %q, want evaluation-invalid.v1", fixture.FixtureVersion)
	}
	if len(fixture.Cases) < 6 {
		t.Fatalf("invalid fixture count = %d, want at least 6", len(fixture.Cases))
	}
	for _, tc := range fixture.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			switch {
			case tc.AxisValue != nil:
				if _, err := NewAxisAssessment(*tc.AxisValue, tc.Rationale); err == nil {
					t.Fatal("invalid axis fixture unexpectedly succeeded")
				}
			case tc.ConfidenceBasisPoints != nil:
				if _, err := NewConfidence(*tc.ConfidenceBasisPoints, tc.ConfidenceRationale); err == nil {
					t.Fatal("invalid confidence fixture unexpectedly succeeded")
				}
			default:
				t.Fatal("invalid fixture has no testable input")
			}
		})
	}
}
