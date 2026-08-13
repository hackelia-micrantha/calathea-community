package domain

import (
	"testing"
	"time"
)

func mustAxis(t *testing.T, value int, rationale string) AxisAssessment {
	t.Helper()
	axis, err := NewAxisAssessment(value, rationale)
	if err != nil {
		t.Fatalf("NewAxisAssessment() error = %v", err)
	}
	return axis
}

func mustAxes(t *testing.T) EvaluationAxes {
	t.Helper()
	axes, err := NewEvaluationAxes(
		mustAxis(t, 4, "material portfolio benefit"),
		mustAxis(t, 2, "bounded implementation"),
		mustAxis(t, 3, "addresses known delivery risk"),
		mustAxis(t, 2, "enables a follow-on automation"),
		mustAxis(t, 3, "delay has measurable planning-horizon cost"),
		[]string{"follow-on automation"},
	)
	if err != nil {
		t.Fatalf("NewEvaluationAxes() error = %v", err)
	}
	return axes
}

func TestAxisScoreRejectsOutOfRangeValues(t *testing.T) {
	for _, value := range []int{0, 6} {
		if _, err := NewAxisScore(value); err == nil {
			t.Fatalf("NewAxisScore(%d) succeeded", value)
		}
	}
}

func TestAxisAssessmentRequiresRationale(t *testing.T) {
	if _, err := NewAxisAssessment(3, ""); err == nil {
		t.Fatal("axis assessment without rationale succeeded")
	}
}

func TestOptionalityAboveOneRequiresNamedFuturePath(t *testing.T) {
	_, err := NewEvaluationAxes(
		mustAxis(t, 3, "impact"), mustAxis(t, 3, "effort"), mustAxis(t, 3, "risk"),
		mustAxis(t, 2, "optionality"), mustAxis(t, 3, "urgency"), nil,
	)
	if err == nil {
		t.Fatal("optionality above one without named future path succeeded")
	}
}

func TestBaseScoreIsExactAndDoesNotUseConfidence(t *testing.T) {
	axes := mustAxes(t)
	score, err := CalculateBaseScore(axes)
	if err != nil {
		t.Fatalf("CalculateBaseScore() error = %v", err)
	}
	if got, want := score.String(), "36"; got != want {
		t.Fatalf("BaseScore.String() = %q, want %q", got, want)
	}
	low, _ := NewConfidence(1000, "weak evidence")
	high, _ := NewConfidence(9500, "exceptional evidence")
	if low.BasisPoints() == high.BasisPoints() {
		t.Fatal("test confidence values unexpectedly equal")
	}
	again, err := CalculateBaseScore(axes)
	if err != nil || again != score {
		t.Fatalf("base score changed independently of axes: got %#v err=%v, want %#v", again, err, score)
	}
}

func TestConfidenceBandsUseExactThresholds(t *testing.T) {
	cases := []struct {
		basisPoints int
		want        ConfidenceBand
	}{{3999, ConfidenceWeakEvidence}, {4000, ConfidenceVisibleUncertainty}, {6999, ConfidenceVisibleUncertainty}, {7000, ConfidenceWellSupported}, {8999, ConfidenceWellSupported}, {9000, ConfidenceExceptionalEvidence}}
	for _, tc := range cases {
		confidence, err := NewConfidence(tc.basisPoints, "threshold test")
		if err != nil {
			t.Fatalf("NewConfidence(%d) error = %v", tc.basisPoints, err)
		}
		if got := confidence.Band(); got != tc.want {
			t.Fatalf("Confidence(%d).Band() = %q, want %q", tc.basisPoints, got, tc.want)
		}
	}
}

func validVersionInput(t *testing.T) EvaluationVersionInput {
	t.Helper()
	projectID := ProjectID("project-1")
	evaluation, _ := NewEvaluation(EvaluationID("evaluation-1"), projectID)
	projectVersion := testProjectVersion(t, ProjectVersionID("project-v1"), projectID)
	confidence, _ := NewConfidence(7000, "well-supported evidence")
	freshness, _ := NewFreshnessMetadata(testTime())
	return EvaluationVersionInput{
		ID:                     EvaluationVersionID("evaluation-v1"),
		Evaluation:             evaluation,
		ProjectVersion:         projectVersion,
		EvaluatedAt:            testTime(),
		PlanningHorizon:        "next 90 days",
		Freshness:              freshness,
		Axes:                   mustAxes(t),
		Confidence:             confidence,
		Derivation:             EvaluationDerivationAuthored,
		EvidenceIDs:            []EvidenceReferenceID{"evidence-1"},
		SemanticVersion:        EvaluationSemanticVersionV1,
		FormulaSemanticVersion: BaseScoreFormulaSemanticVersionV1,
		AcceptedBy:             testMaintainer(t),
		AcceptedAt:             testTime(),
	}
}

func TestEvaluationVersionRequiresMatchingProject(t *testing.T) {
	input := validVersionInput(t)
	input.ProjectVersion = testProjectVersion(t, ProjectVersionID("project-v2"), ProjectID("project-2"))
	if _, err := NewEvaluationVersion(input); err == nil {
		t.Fatal("evaluation version accepted a project version from another project")
	}
}

func TestEvaluationVersionAllowsAuthoredEvaluationWithoutEvidenceReference(t *testing.T) {
	input := validVersionInput(t)
	input.EvidenceIDs = nil
	version, err := NewEvaluationVersion(input)
	if err != nil {
		t.Fatalf("authored evaluation without evidence reference failed: %v", err)
	}
	if got, want := version.Derivation(), EvaluationDerivationAuthored; got != want {
		t.Fatalf("Derivation() = %q, want %q", got, want)
	}
	if len(version.EvidenceIDs()) != 0 {
		t.Fatalf("EvidenceIDs() = %v, want empty", version.EvidenceIDs())
	}
}

func TestEvaluationVersionPromotedDraftRequiresSourceEvidence(t *testing.T) {
	input := validVersionInput(t)
	input.Derivation = EvaluationDerivationPromotedDraft
	input.EvidenceIDs = nil
	if _, err := NewEvaluationVersion(input); err == nil {
		t.Fatal("promoted draft without source evidence succeeded")
	}

	input.EvidenceIDs = []EvidenceReferenceID{"draft-evidence-1"}
	version, err := NewEvaluationVersion(input)
	if err != nil {
		t.Fatalf("promoted draft with source evidence failed: %v", err)
	}
	if got, want := version.Derivation(), EvaluationDerivationPromotedDraft; got != want {
		t.Fatalf("Derivation() = %q, want %q", got, want)
	}
}

func TestEvaluationVersionRejectsUnknownDerivation(t *testing.T) {
	input := validVersionInput(t)
	input.Derivation = EvaluationDerivation("unknown")
	if _, err := NewEvaluationVersion(input); err == nil {
		t.Fatal("evaluation version with unknown derivation succeeded")
	}
}

func TestEvaluationVersionRequiresMaintainerAcceptanceAuthority(t *testing.T) {
	input := validVersionInput(t)
	input.AcceptedBy = Actor{}
	if _, err := NewEvaluationVersion(input); err == nil {
		t.Fatal("evaluation version without maintainer acceptance authority succeeded")
	}
}

func TestEvaluationVersionRejectsInvalidEvidenceReferenceAndUnsupportedVersions(t *testing.T) {
	input := validVersionInput(t)
	input.EvidenceIDs = []EvidenceReferenceID{""}
	if _, err := NewEvaluationVersion(input); err == nil {
		t.Fatal("evaluation version with empty evidence reference succeeded")
	}
	input = validVersionInput(t)
	input.SemanticVersion = "evaluation.unavailable"
	if _, err := NewEvaluationVersion(input); err == nil {
		t.Fatal("evaluation version with unsupported semantic version succeeded")
	}
	input = validVersionInput(t)
	input.FormulaSemanticVersion = "evaluation.base_score.unavailable"
	if _, err := NewEvaluationVersion(input); err == nil {
		t.Fatal("evaluation version with unsupported formula semantic version succeeded")
	}
}

func TestEvaluationVersionRejectsFutureEvidenceAndAcceptanceBeforeEvaluation(t *testing.T) {
	input := validVersionInput(t)
	futureFreshness, _ := NewFreshnessMetadata(input.EvaluatedAt.Add(time.Hour))
	input.Freshness = futureFreshness
	if _, err := NewEvaluationVersion(input); err == nil {
		t.Fatal("evaluation version accepted evidence from the future")
	}
	input = validVersionInput(t)
	input.AcceptedAt = input.EvaluatedAt.Add(-time.Second)
	if _, err := NewEvaluationVersion(input); err == nil {
		t.Fatal("evaluation version accepted before it was evaluated")
	}
}

func TestEvaluationVersionSupersessionRequiresSameEvaluationHistory(t *testing.T) {
	priorInput := validVersionInput(t)
	prior, err := NewEvaluationVersion(priorInput)
	if err != nil {
		t.Fatalf("create prior evaluation version: %v", err)
	}

	next := validVersionInput(t)
	next.ID = EvaluationVersionID("evaluation-v2")
	next.Supersedes = &prior
	version, err := NewEvaluationVersion(next)
	if err != nil {
		t.Fatalf("valid supersession failed: %v", err)
	}
	if got := version.Supersedes(); got == nil || *got != prior.ID() {
		t.Fatalf("Supersedes() = %v, want %q", got, prior.ID())
	}

	otherProject := ProjectID("project-2")
	otherEvaluation, _ := NewEvaluation(EvaluationID("evaluation-2"), otherProject)
	otherVersion := testProjectVersion(t, ProjectVersionID("project-2-v1"), otherProject)
	other := validVersionInput(t)
	other.ID = EvaluationVersionID("evaluation-other-v1")
	other.Evaluation = otherEvaluation
	other.ProjectVersion = otherVersion
	priorOther, err := NewEvaluationVersion(other)
	if err != nil {
		t.Fatalf("create other evaluation version: %v", err)
	}

	next = validVersionInput(t)
	next.ID = EvaluationVersionID("evaluation-v3")
	next.Supersedes = &priorOther
	if _, err := NewEvaluationVersion(next); err == nil {
		t.Fatal("evaluation version superseded a version from another evaluation history")
	}
}

func TestEvaluationVersionDefensivelyCopiesInputs(t *testing.T) {
	input := validVersionInput(t)
	evidence := input.EvidenceIDs
	version, err := NewEvaluationVersion(input)
	if err != nil {
		t.Fatalf("NewEvaluationVersion() error = %v", err)
	}
	evidence[0] = EvidenceReferenceID("changed")
	paths := version.Axes().FuturePaths()
	paths[0] = "changed"
	if got := version.EvidenceIDs()[0]; got != EvidenceReferenceID("evidence-1") {
		t.Fatalf("stored evidence changed to %q", got)
	}
	if got := version.Axes().FuturePaths()[0]; got != "follow-on automation" {
		t.Fatalf("stored future path changed to %q", got)
	}
}
