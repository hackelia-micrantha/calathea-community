package policy

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/hackelia-micrantha/calathea-community/internal/domain"
)

func baselineComposeFixture(t *testing.T, subject domain.PolicySubject, state *domain.LifecycleState, evaluation *domain.EvaluationVersion) ComposeBaselineRequest {
	t.Helper()
	set := baselinePolicySet(t)
	result := ComposeBaselineRequest{
		PolicySet:   set,
		OperationID: "operation-1",
		Subject:     subject,
	}
	for _, instance := range set.Instances() {
		request := request(t, set, instance.ID(), subject)
		request.DecisionID = domain.PolicyDecisionID("decision-" + string(instance.ID()))
		request.OperationID = result.OperationID
		request.Context.LifecycleState = state
		request.Context.Evaluation = evaluation
		request.Context.AsOf = policyTestTime()
		decision, err := Evaluate(request)
		if err != nil {
			t.Fatalf("Evaluate(%s): %v", instance.ID(), err)
		}
		result.Decisions = append(result.Decisions, decision)
	}
	return result
}

func mustCompose(t *testing.T, request ComposeBaselineRequest) Composition {
	t.Helper()
	out, err := ComposeBaseline(request)
	if err != nil {
		t.Fatalf("ComposeBaseline() = %v", err)
	}
	return out
}

func requireCompositionFailure(t *testing.T, err error, code string) {
	t.Helper()
	var typed *CompositionFailure
	if !errors.As(err, &typed) || typed.Code != code {
		t.Fatalf("error = %v, want composition code %q", err, code)
	}
}

func TestComposeBaselineStableOrderingAndDefensiveTrace(t *testing.T) {
	state := domain.LifecycleApproved
	evaluation := testEvaluation(t, 9000, policyTestTime())
	request := baselineComposeFixture(t, projectSubject(t), &state, &evaluation)
	first := mustCompose(t, request)
	if first.Outcome() != CompositionAllowed || len(first.Steps()) != len(request.PolicySet.Instances()) {
		t.Fatalf("outcome=%q, steps=%d", first.Outcome(), len(first.Steps()))
	}
	for i, step := range first.Steps() {
		instance := request.PolicySet.Instances()[i]
		if step.InstanceID != instance.ID() || step.PolicyID != instance.PolicyID() ||
			step.Phase != instance.Phase() || step.Priority != instance.Priority() {
			t.Fatalf("step %d is out of canonical policy order: %#v", i, step)
		}
	}
	shuffled := request
	shuffled.Decisions = append([]domain.PolicyDecision(nil), request.Decisions...)
	for i, j := 0, len(shuffled.Decisions)-1; i < j; i, j = i+1, j-1 {
		shuffled.Decisions[i], shuffled.Decisions[j] = shuffled.Decisions[j], shuffled.Decisions[i]
	}
	second := mustCompose(t, shuffled)
	if first.Outcome() != second.Outcome() || !reflect.DeepEqual(first.Steps(), second.Steps()) {
		t.Fatal("decision slice ordering changed composition")
	}
	steps := first.Steps()
	steps[0].Interpretation = "tampered"
	if first.Steps()[0].Interpretation == "tampered" {
		t.Fatal("caller mutated immutable composition trace")
	}
}

func TestComposeBaselineDenialDoesNotEraseIndependentReview(t *testing.T) {
	state := domain.LifecycleCandidate
	evaluation := testEvaluation(t, 1000, policyTestTime())
	result := mustCompose(t, baselineComposeFixture(t, projectSubject(t), &state, &evaluation))
	if result.Outcome() != CompositionDenied {
		t.Fatalf("outcome = %q, want denied", result.Outcome())
	}
	var deny, review bool
	for _, step := range result.Steps() {
		if step.Interpretation == "hard_deny" {
			deny = true
		}
		if step.Interpretation == "review_required" {
			review = true
		}
	}
	if !deny || !review {
		t.Fatalf("missing hard deny or independent review trace: %#v", result.Steps())
	}
}

func TestComposeBaselineReviewWithoutHardDenial(t *testing.T) {
	state := domain.LifecycleApproved
	evaluation := testEvaluation(t, 1000, policyTestTime())
	result := mustCompose(t, baselineComposeFixture(t, projectSubject(t), &state, &evaluation))
	if result.Outcome() != CompositionReviewNeeded {
		t.Fatalf("outcome = %q, want review_required", result.Outcome())
	}
}

func TestComposeBaselinePlacementSetPreservesNotApplicableDecisions(t *testing.T) {
	subject := placementSubject(t, domain.PlacementNow)
	set := baselinePolicySet(t)
	maximum := 3
	result := ComposeBaselineRequest{PolicySet: set, OperationID: "operation-1", Subject: subject}
	for _, instance := range set.Instances() {
		req := request(t, set, instance.ID(), subject)
		req.DecisionID = domain.PolicyDecisionID("decision-" + string(instance.ID()))
		req.OperationID = result.OperationID
		req.Context.SelectedCount = &maximum
		decision, err := Evaluate(req)
		if err != nil {
			t.Fatal(err)
		}
		result.Decisions = append(result.Decisions, decision)
	}
	out := mustCompose(t, result)
	if out.Outcome() != CompositionAllowed {
		t.Fatalf("outcome=%q, want allowed", out.Outcome())
	}
	var allowed, notApplicable int
	for _, step := range out.Steps() {
		if step.Interpretation == "hard_allow" {
			allowed++
		}
		if step.Interpretation == "not_applicable" {
			notApplicable++
		}
	}
	if allowed != 1 || notApplicable != 5 {
		t.Fatalf("placement set trace had %d allow and %d not applicable", allowed, notApplicable)
	}
}

func TestComposeBaselineMissingInputBehaviors(t *testing.T) {
	cases := []struct {
		name     string
		behavior domain.PolicyMissingInputBehavior
		outcome  CompositionOutcome
		code     string
	}{
		{"exclude", domain.PolicyMissingInputExcludeSubject, CompositionExcluded, "missing_input_exclude_subject"},
		{"deny", domain.PolicyMissingInputDeny, CompositionDenied, "missing_input_deny"},
		{"review", domain.PolicyMissingInputRequireReview, CompositionReviewNeeded, "missing_input_require_review"},
		{"fail", domain.PolicyMissingInputFailOperation, CompositionFailed, "missing_input_fail_operation"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			instances := baselineInstances(t)
			lifecycle := instances[0]
			reconfigured := mustPolicyInstance(t, domain.PolicyInstanceInput{
				ID: lifecycle.ID(), PolicyID: lifecycle.PolicyID(),
				EvaluatorType: lifecycle.EvaluatorType(), EvaluatorVersion: lifecycle.EvaluatorVersion(),
				Phase: lifecycle.Phase(), EffectClass: lifecycle.EffectClass(),
				SubjectType: lifecycle.SubjectType(), RequiredInputs: lifecycle.RequiredInputs(),
				MissingInputBehavior: tc.behavior, Priority: lifecycle.Priority(),
				Exceptionability: lifecycle.Exceptionability(), Parameters: lifecycle.Parameters(),
				Rationale: lifecycle.Rationale(),
			})
			instances[0] = reconfigured
			set, err := domain.NewPolicySetVersion("set-v1", "set", policyTestTime(), instances...)
			if err != nil {
				t.Fatal(err)
			}
			eval := testEvaluation(t, 9000, policyTestTime())
			req := ComposeBaselineRequest{PolicySet: set, OperationID: "op-1", Subject: projectSubject(t)}
			for _, instance := range set.Instances() {
				r := request(t, set, instance.ID(), req.Subject)
				r.DecisionID = domain.PolicyDecisionID("decision-" + string(instance.ID()))
				r.Context.Evaluation = &eval
				r.Context.AsOf = policyTestTime()
				d, err := Evaluate(r)
				if err != nil {
					t.Fatal(err)
				}
				req.Decisions = append(req.Decisions, d)
			}
			result := mustCompose(t, req)
			if result.Outcome() != tc.outcome {
				t.Fatalf("outcome=%q, want %q", result.Outcome(), tc.outcome)
			}
			if result.Steps()[0].Interpretation != tc.code {
				t.Fatalf("first step=%#v", result.Steps()[0])
			}
		})
	}
}

func TestComposeBaselineRejectsIncompleteDuplicateAndForeignDecisions(t *testing.T) {
	state := domain.LifecycleApproved
	evaluation := testEvaluation(t, 9000, policyTestTime())
	base := baselineComposeFixture(t, projectSubject(t), &state, &evaluation)
	cases := []struct {
		name string
		edit func(*ComposeBaselineRequest)
		code string
	}{
		{"missing", func(r *ComposeBaselineRequest) { r.Decisions = r.Decisions[1:] }, "incomplete_decisions"},
		{"duplicate instance", func(r *ComposeBaselineRequest) {
			replacement := decisionInputFrom(r.Decisions[0])
			replacement.ID = "second-decision-for-same-instance"
			d, err := domain.NewPolicyDecision(replacement)
			if err != nil {
				t.Fatal(err)
			}
			r.Decisions[1] = d
		}, "duplicate_instance_decision"},
		{"wrong operation", func(r *ComposeBaselineRequest) { r.OperationID = "other" }, "incoherent_decision"},
		{"wrong subject", func(r *ComposeBaselineRequest) { r.Subject = projectSubjectFor(t, "other-project") }, "incoherent_decision"},
		{"wrong policy set", func(r *ComposeBaselineRequest) {
			var err error
			r.PolicySet, err = domain.NewPolicySetVersion("other-version", r.PolicySet.PolicySetID(), r.PolicySet.CreatedAt(), r.PolicySet.Instances()...)
			if err != nil {
				t.Fatal(err)
			}
		}, "incoherent_decision"},
		{"duplicate decision id", func(r *ComposeBaselineRequest) {
			replacement := decisionInputFrom(r.Decisions[1])
			replacement.ID = r.Decisions[0].ID()
			d, err := domain.NewPolicyDecision(replacement)
			if err != nil {
				t.Fatal(err)
			}
			r.Decisions[1] = d
		}, "duplicate_decision"},
		{"wrong evaluator version", func(r *ComposeBaselineRequest) {
			replacement := decisionInputFrom(r.Decisions[0])
			replacement.EvaluatorVersion = "unsupported"
			d, err := domain.NewPolicyDecision(replacement)
			if err != nil {
				t.Fatal(err)
			}
			r.Decisions[0] = d
		}, "incoherent_decision"},
		{"wrong priority", func(r *ComposeBaselineRequest) {
			replacement := decisionInputFrom(r.Decisions[0])
			replacement.Priority++
			d, err := domain.NewPolicyDecision(replacement)
			if err != nil {
				t.Fatal(err)
			}
			r.Decisions[0] = d
		}, "incoherent_decision"},
		{"invalid effect", func(r *ComposeBaselineRequest) {
			replacement := decisionInputFrom(r.Decisions[0])
			effect, err := domain.NewDiagnosticPolicyEffect("untrusted")
			if err != nil {
				t.Fatal(err)
			}
			replacement.Effects = []domain.PolicyEffect{effect}
			d, err := domain.NewPolicyDecision(replacement)
			if err != nil {
				t.Fatal(err)
			}
			r.Decisions[0] = d
		}, "incoherent_decision"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := base
			req.Decisions = append([]domain.PolicyDecision(nil), base.Decisions...)
			tc.edit(&req)
			_, err := ComposeBaseline(req)
			requireCompositionFailure(t, err, tc.code)
		})
	}
}

func decisionInputFrom(d domain.PolicyDecision) domain.PolicyDecisionInput {
	return domain.PolicyDecisionInput{
		ID: d.ID(), PolicySetVersionID: d.PolicySetVersionID(), PolicyID: d.PolicyID(),
		PolicyInstanceID: d.PolicyInstanceID(), EvaluatorType: d.EvaluatorType(),
		EvaluatorVersion: d.EvaluatorVersion(), Phase: d.Phase(), EffectClass: d.EffectClass(),
		MissingInputBehavior: d.MissingInputBehavior(), Subject: d.Subject(),
		OperationID: d.OperationID(), Result: d.Result(), ReasonCode: d.ReasonCode(),
		RequiredInputs: d.RequiredInputs(), InputReferences: d.InputReferences(),
		MissingInputs: d.MissingInputs(), EvidenceIDs: d.EvidenceIDs(), Effects: d.Effects(),
		Priority: d.Priority(), Rationale: d.Rationale(), CreatedAt: d.CreatedAt(),
	}
}

func TestComposeBaselineRejectsFutureAndMissingTime(t *testing.T) {
	state := domain.LifecycleApproved
	evaluation := testEvaluation(t, 9000, policyTestTime())
	req := baselineComposeFixture(t, projectSubject(t), &state, &evaluation)
	changed := decisionInputFrom(req.Decisions[0])
	changed.CreatedAt = policyTestTime().Add(-time.Minute)
	d, err := domain.NewPolicyDecision(changed)
	if err != nil {
		t.Fatal(err)
	}
	req.Decisions[0] = d
	_, err = ComposeBaseline(req)
	requireCompositionFailure(t, err, "incoherent_decision")
}
