package policy

import (
	"reflect"
	"testing"

	"github.com/hackelia-micrantha/calathea-community/internal/domain"
)

func scoreMultiplierInstance(t *testing.T, id domain.PolicyInstanceID, policyID domain.PolicyID, factorBasisPoints, priority int) domain.PolicyInstance {
	t.Helper()
	params, err := domain.NewScoreMultiplierParameters(factorBasisPoints)
	if err != nil {
		t.Fatal(err)
	}
	return mustPolicyInstance(t, domain.PolicyInstanceInput{
		ID:                   id,
		PolicyID:             policyID,
		EvaluatorType:        domain.PolicyEvaluatorScoreMultiplier,
		EvaluatorVersion:     EvaluatorSemanticVersionV1,
		Phase:                domain.PolicyPhaseCandidateAdjustment,
		EffectClass:          domain.PolicyEffectSoft,
		SubjectType:          domain.PolicySubjectProject,
		RequiredInputs:       nil,
		MissingInputBehavior: domain.PolicyMissingInputFailOperation,
		Priority:             priority,
		Exceptionability:     domain.PolicyNotExceptionable,
		Parameters:           params,
		Rationale:            "explicit public test multiplier",
	})
}

func softPolicySet(t *testing.T, cumulativeMin, cumulativeMax int, factors ...int) domain.PolicySetVersion {
	t.Helper()
	contract, err := domain.NewScoreMultiplierContract(
		domain.PolicySoftCombinatorMultiplyV1,
		5000,
		20000,
		cumulativeMin,
		cumulativeMax,
	)
	if err != nil {
		t.Fatal(err)
	}
	instances := baselineInstances(t)
	for i, factor := range factors {
		id := domain.PolicyInstanceID("score-multiplier-" + string(rune('a'+i)))
		policyID := domain.PolicyID("orientation.score." + string(rune('a'+i)))
		instances = append(instances, scoreMultiplierInstance(t, id, policyID, factor, 45+i))
	}
	version, err := domain.NewPolicySetVersionWithScoreMultiplierContract(
		"policy-set-soft-v1",
		"policy-set",
		policyTestTime(),
		contract,
		instances...,
	)
	if err != nil {
		t.Fatal(err)
	}
	return version
}

func composeProjectForSet(t *testing.T, set domain.PolicySetVersion, lifecycle domain.LifecycleState, confidence int) (ComposeBaselineRequest, Composition) {
	t.Helper()
	subject := projectSubject(t)
	evaluation := testEvaluation(t, confidence, policyTestTime())
	req := ComposeBaselineRequest{PolicySet: set, OperationID: "operation-soft", Subject: subject}
	for _, instance := range set.Instances() {
		r := request(t, set, instance.ID(), subject)
		r.DecisionID = domain.PolicyDecisionID("decision-" + string(instance.ID()))
		r.OperationID = req.OperationID
		r.Context.LifecycleState = &lifecycle
		r.Context.Evaluation = &evaluation
		r.Context.AsOf = policyTestTime()
		decision, err := Evaluate(r)
		if err != nil {
			t.Fatalf("Evaluate(%s): %v", instance.ID(), err)
		}
		req.Decisions = append(req.Decisions, decision)
	}
	return req, mustCompose(t, req)
}

func TestBaselineCompositionHasExactNeutralMultiplier(t *testing.T) {
	state := domain.LifecycleApproved
	evaluation := testEvaluation(t, 9000, policyTestTime())
	out := mustCompose(t, baselineComposeFixture(t, projectSubject(t), &state, &evaluation))
	if out.ScoreMultiplier().String() != "1" {
		t.Fatalf("baseline multiplier = %q, want 1", out.ScoreMultiplier().String())
	}
}

func TestScoreMultiplierRequiresExplicitPolicySetContract(t *testing.T) {
	instances := append(baselineInstances(t), scoreMultiplierInstance(t, "score-a", "orientation.score.a", 12500, 45))
	set, err := domain.NewPolicySetVersion("set-without-soft-contract", "policy-set", policyTestTime(), instances...)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateForActivation(set); err == nil {
		t.Fatal("soft multiplier policy activated without explicit contract")
	}
}

func TestScoreMultipliersComposeExactlyAndIndependentlyOfCallerOrder(t *testing.T) {
	set := softPolicySet(t, 2500, 40000, 12500, 8000)
	if err := ValidateForActivation(set); err != nil {
		t.Fatal(err)
	}
	req, first := composeProjectForSet(t, set, domain.LifecycleApproved, 9000)
	if got := first.ScoreMultiplier().String(); got != "1" {
		t.Fatalf("12500 * 8000 basis points = %q, want exact 1", got)
	}
	var softSteps []CompositionStep
	for _, step := range first.Steps() {
		if step.Interpretation == "soft_score_multiplier" {
			softSteps = append(softSteps, step)
		}
	}
	if len(softSteps) != 2 {
		t.Fatalf("soft steps = %d, want 2", len(softSteps))
	}
	if softSteps[0].ScoreMultiplierBefore != "1" ||
		softSteps[0].ScoreMultiplierFactor != "5/4" ||
		softSteps[0].ScoreMultiplierProposedAfter != "5/4" ||
		softSteps[1].ScoreMultiplierFactor != "4/5" ||
		softSteps[1].ScoreMultiplierProposedAfter != "1" {
		t.Fatalf("unexpected exact multiplier trace: %#v", softSteps)
	}
	shuffled := req
	shuffled.Decisions = append([]domain.PolicyDecision(nil), req.Decisions...)
	for i, j := 0, len(shuffled.Decisions)-1; i < j; i, j = i+1, j-1 {
		shuffled.Decisions[i], shuffled.Decisions[j] = shuffled.Decisions[j], shuffled.Decisions[i]
	}
	second := mustCompose(t, shuffled)
	if first.ScoreMultiplier().String() != second.ScoreMultiplier().String() ||
		!reflect.DeepEqual(first.Steps(), second.Steps()) {
		t.Fatal("caller decision order changed exact multiplier composition or trace")
	}
}

func TestScoreMultiplierCumulativeBoundsFailClosed(t *testing.T) {
	set := softPolicySet(t, 5000, 11000, 12000)
	if err := ValidateForActivation(set); err != nil {
		t.Fatal(err)
	}
	subject := projectSubject(t)
	evaluation := testEvaluation(t, 9000, policyTestTime())
	state := domain.LifecycleApproved
	req := ComposeBaselineRequest{PolicySet: set, OperationID: "op-bound", Subject: subject}
	for _, instance := range set.Instances() {
		r := request(t, set, instance.ID(), subject)
		r.DecisionID = domain.PolicyDecisionID("decision-" + string(instance.ID()))
		r.OperationID = req.OperationID
		r.Context.LifecycleState = &state
		r.Context.Evaluation = &evaluation
		r.Context.AsOf = policyTestTime()
		d, err := Evaluate(r)
		if err != nil {
			t.Fatal(err)
		}
		req.Decisions = append(req.Decisions, d)
	}
	_, err := ComposeBaseline(req)
	requireCompositionFailure(t, err, "soft_effect_bounds")
}

func TestHardDenialSuppressesEffectiveMultiplierButKeepsProposedTrace(t *testing.T) {
	set := softPolicySet(t, 5000, 20000, 12500)
	_, out := composeProjectForSet(t, set, domain.LifecycleCandidate, 9000)
	if out.Outcome() != CompositionDenied {
		t.Fatalf("outcome = %q, want denied", out.Outcome())
	}
	if got := out.ScoreMultiplier().String(); got != "1" {
		t.Fatalf("suppressed effective multiplier = %q, want neutral 1", got)
	}
	found := false
	for _, step := range out.Steps() {
		if step.Interpretation != "soft_score_multiplier" {
			continue
		}
		found = true
		if !step.SuppressedByHardPolicy ||
			step.ScoreMultiplierProposedAfter != "5/4" ||
			step.ScoreMultiplierEffectiveAfter != "1" {
			t.Fatalf("soft suppression trace = %#v", step)
		}
	}
	if !found {
		t.Fatal("missing retained soft multiplier trace")
	}
}

func TestConfidenceAndFreshnessNeverImplicitlyCreateMultiplier(t *testing.T) {
	set := baselinePolicySet(t)
	_, high := composeProjectForSet(t, set, domain.LifecycleApproved, 9000)
	_, low := composeProjectForSet(t, set, domain.LifecycleApproved, 1000)
	if high.ScoreMultiplier().String() != "1" || low.ScoreMultiplier().String() != "1" {
		t.Fatalf("confidence changed multiplier: high=%q low=%q", high.ScoreMultiplier().String(), low.ScoreMultiplier().String())
	}
	if low.Outcome() != CompositionReviewNeeded {
		t.Fatalf("low confidence outcome = %q, want review without score mutation", low.Outcome())
	}
}
