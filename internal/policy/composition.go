// Package policy composes immutable, already-evaluated policy decisions.
// This file implements the bounded UC-01 baseline; typed soft arithmetic and
// disposition-time exception integration require separately versioned contracts.
package policy

import (
	"fmt"
	"sort"

	"github.com/hackelia-micrantha/calathea-community/internal/domain"
)

// CompositionOutcome is a result for one exact subject and operation, not an
// orientation placement, an authorization to perform an effect, or a disposition.
type CompositionOutcome string

const (
	CompositionAllowed       CompositionOutcome = "allowed"
	CompositionDenied        CompositionOutcome = "denied"
	CompositionExcluded      CompositionOutcome = "excluded"
	CompositionReviewNeeded  CompositionOutcome = "review_required"
	CompositionFailed        CompositionOutcome = "fail_operation"
	CompositionNotApplicable CompositionOutcome = "not_applicable"
)

// CompositionStep preserves every original decision and its interpreted
// contribution in stable phase/priority/policy/instance order. The full
// PolicyDecision remains the source of truth for effects and input evidence.
type CompositionStep struct {
	DecisionID domain.PolicyDecisionID
	PolicyID domain.PolicyID
	InstanceID domain.PolicyInstanceID
	Phase domain.PolicyPhase
	Priority int
	Result domain.PolicyDecisionResult
	Interpretation string
	ReasonCode string
}

// Composition returns a deterministic subject-level summary without rewriting
// any decision. Accessors clone the trace to prevent caller mutation.
type Composition struct {
	policySetVersionID domain.PolicySetVersionID
	operationID domain.OperationID
	subject domain.PolicySubject
	outcome CompositionOutcome
	steps []CompositionStep
}

func (c Composition) PolicySetVersionID() domain.PolicySetVersionID { return c.policySetVersionID }
func (c Composition) OperationID() domain.OperationID { return c.operationID }
func (c Composition) Subject() domain.PolicySubject { return c.subject }
func (c Composition) Outcome() CompositionOutcome { return c.outcome }
func (c Composition) Steps() []CompositionStep {
	return append([]CompositionStep(nil), c.steps...)
}

// ComposeBaselineRequest supplies the complete decision set for one subject
// and operation. Every configured instance must have exactly one decision,
// including explicit not_applicable decisions for other subject types.
type ComposeBaselineRequest struct {
	PolicySet domain.PolicySetVersion
	OperationID domain.OperationID
	Subject domain.PolicySubject
	Decisions []domain.PolicyDecision
}

// CompositionFailure distinguishes invalid/untrusted composition input from
// evaluator missing-input results already retained as PolicyDecision records.
type CompositionFailure struct {
	Code string
	Detail string
}
func (e *CompositionFailure) Error() string { return e.Code + ": " + e.Detail }
func compositionFailure(code, detail string) error {
	return &CompositionFailure{Code: code, Detail: detail}
}

// ComposeBaseline interprets only retained evaluator results; it never reruns
// evaluator truth or grants permission from a priority/allow that conflicts
// with any hard denial. All independent results remain visible in its trace.
func ComposeBaseline(req ComposeBaselineRequest) (Composition, error) {
	fail := func(code, detail string) (Composition, error) {
		return Composition{}, compositionFailure(code, detail)
	}
	if err := ValidateForActivation(req.PolicySet); err != nil {
		return fail("invalid_policy_set", err.Error())
	}
	if req.OperationID == "" || !req.Subject.Type().Valid() || req.Subject.ID() == "" {
		return fail("invalid_context", "composition requires one operation and valid typed subject")
	}
	instances := req.PolicySet.Instances()
	if len(req.Decisions) != len(instances) {
		return fail("incomplete_decisions", fmt.Sprintf("expected %d decisions, got %d", len(instances), len(req.Decisions)))
	}
	byInstance := make(map[domain.PolicyInstanceID]domain.PolicyInstance, len(instances))
	for _, instance := range instances {
		byInstance[instance.ID()] = instance
	}
	// Sort input before validating it: a malformed or duplicate set must not
	// select a different first error merely because a caller shuffled a slice.
	decisions := append([]domain.PolicyDecision(nil), req.Decisions...)
	sort.Slice(decisions, func(i,j int) bool {
		if decisions[i].PolicyInstanceID() != decisions[j].PolicyInstanceID() {
			return decisions[i].PolicyInstanceID() < decisions[j].PolicyInstanceID()
		}
		return decisions[i].ID() < decisions[j].ID()
	})
	byDecisionInstance := make(map[domain.PolicyInstanceID]domain.PolicyDecision,len(decisions))
	byID := make(map[domain.PolicyDecisionID]bool,len(decisions))
	for _, decision := range decisions {
		instance, exists := byInstance[decision.PolicyInstanceID()]
		if !exists {
			return fail("foreign_decision", fmt.Sprintf("unknown policy instance %q",decision.PolicyInstanceID()))
		}
		if byID[decision.ID()] || decision.ID()=="" {
			return fail("duplicate_decision", fmt.Sprintf("duplicate/invalid decision identity %q",decision.ID()))
		}
		byID[decision.ID()] = true
		if _, duplicate := byDecisionInstance[decision.PolicyInstanceID()]; duplicate {
			return fail("duplicate_instance_decision", fmt.Sprintf("instance %q has multiple decisions",decision.PolicyInstanceID()))
		}
		if err := validateComposedDecision(req,instance,decision); err != nil {
			return fail("incoherent_decision",fmt.Sprintf("instance %q: %v",instance.ID(),err))
		}
		byDecisionInstance[instance.ID()] = decision
	}
	steps:=make([]CompositionStep,0,len(instances))
	var hasHardAllow,hasDeny,hasExclude,hasReview,hasFailure bool
	for _, instance := range instances {
		decision, exists:=byDecisionInstance[instance.ID()]
		if !exists {
			return fail("incomplete_decisions",fmt.Sprintf("missing decision for instance %q",instance.ID()))
		}
		interpretation,err:=interpretBaselineDecision(instance,decision)
		if err!=nil {
			return fail("unsupported_result",fmt.Sprintf("instance %q: %v",instance.ID(),err))
		}
		switch interpretation {
		case "hard_allow":
			hasHardAllow=true
		case "hard_deny","missing_input_deny":
			hasDeny=true
		case "missing_input_exclude_subject":
			hasExclude=true
		case "review_required","missing_input_require_review":
			hasReview=true
		case "missing_input_fail_operation":
			hasFailure=true
		}
		steps=append(steps,CompositionStep{
			DecisionID:decision.ID(), PolicyID:instance.PolicyID(),
			InstanceID:instance.ID(),Phase:instance.Phase(),
			Priority:instance.Priority(),Result:decision.Result(),
			Interpretation:interpretation,ReasonCode:decision.ReasonCode(),
		})
	}
	// Failure takes precedence over denial, denial over exclusion, exclusion
	// over review, review over an otherwise passing hard policy. These
	// diagnostic categories never let a later allow cancel an earlier block.
	outcome:=CompositionNotApplicable
	switch {
	case hasFailure: outcome=CompositionFailed
	case hasDeny: outcome=CompositionDenied
	case hasExclude: outcome=CompositionExcluded
	case hasReview: outcome=CompositionReviewNeeded
	case hasHardAllow: outcome=CompositionAllowed
	}
	return Composition{
		policySetVersionID:req.PolicySet.ID(),operationID:req.OperationID,
		subject:req.Subject,outcome:outcome,steps:steps,
	},nil
}

func validateComposedDecision(req ComposeBaselineRequest,instance domain.PolicyInstance,d domain.PolicyDecision) error {
	if d.PolicySetVersionID()!=req.PolicySet.ID() ||
		d.OperationID()!=req.OperationID || d.Subject()!=req.Subject ||
		d.PolicyID()!=instance.PolicyID() || d.PolicyInstanceID()!=instance.ID() ||
		d.EvaluatorType()!=instance.EvaluatorType() ||
		d.EvaluatorVersion()!=instance.EvaluatorVersion() ||
		d.ConfigurationSchemaVersion()!=instance.ConfigurationSchemaVersion() ||
		d.Workflow()!=instance.Workflow() || d.Phase()!=instance.Phase() ||
		d.Priority()!=instance.Priority() ||
		d.EffectClass()!=instance.EffectClass() ||
		d.MissingInputBehavior()!=instance.MissingInputBehavior() ||
		d.Rationale()!=instance.Rationale() ||
		!sameInputKinds(d.RequiredInputs(),instance.RequiredInputs()) {
		return fmt.Errorf("decision identity/configuration/input declaration does not match the exact policy instance")
	}
	if d.CreatedAt().Before(req.PolicySet.CreatedAt()) || d.CreatedAt().IsZero() {
		return fmt.Errorf("decision predates its policy set version")
	}
	missing:=d.MissingInputs()
	references:=d.InputReferences()
	required:=d.RequiredInputs()
	if d.Result()==domain.PolicyDecisionNotApplicable {
		if len(missing)!=0 || len(references)!=0 || len(d.Effects())!=0 {
			return fmt.Errorf("not_applicable must have empty input/effect trace")
		}
		if instance.SubjectType()==req.Subject.Type() {
			// Capacity constraints are not applicable to other placements of
			// the same typed subject; other baseline evaluators are applicable
			// to every subject of their declared type.
			if instance.EvaluatorType()!=domain.PolicyEvaluatorCapacityLimit {
				return fmt.Errorf("matching subject may not omit baseline policy")
			}
			placement,ok:=instance.Parameters().Placement()
			if !ok || req.Subject.ID()==string(placement) {
				return fmt.Errorf("capacity constraint cannot omit its matching placement set")
			}
		}
		return nil
	}
	if instance.SubjectType()!=req.Subject.Type() {
		return fmt.Errorf("foreign subject type produced an applicable result")
	}
	if instance.EvaluatorType()==domain.PolicyEvaluatorCapacityLimit {
		placement,ok:=instance.Parameters().Placement()
		if !ok || req.Subject.ID()!=string(placement) {
			return fmt.Errorf("capacity constraint applies to a different placement set")
		}
	}
	if len(references)+len(missing)!=len(required) {
		return fmt.Errorf("incomplete required input trace")
	}
	if d.Result()==domain.PolicyDecisionIndeterminate {
		if len(missing)==0 || len(d.Effects())!=0 {
			return fmt.Errorf("indeterminate requires missing input and no effects")
		}
	} else if len(missing)!=0 {
		return fmt.Errorf("non-indeterminate decision reports missing input")
	}
	if instance.EffectClass()==domain.PolicyEffectSoft {
		return fmt.Errorf("no v0 soft evaluator/combinator contract is activated")
	}
	switch d.Result() {
	case domain.PolicyDecisionAllow,domain.PolicyDecisionDeny:
		if d.Result()==domain.PolicyDecisionDeny && instance.EffectClass()!=domain.PolicyEffectHard {
			return fmt.Errorf("non-hard policy cannot deny")
		}
		if instance.EvaluatorType()==domain.PolicyEvaluatorCapacityLimit {
			effects:=d.Effects()
			placement,havePlacement:=instance.Parameters().Placement()
			maximum,haveMaximum:=instance.Parameters().Maximum()
			if len(effects)!=1 || effects[0].Type()!=domain.PolicyEffectCapacityLimit ||
				!havePlacement || !haveMaximum {
				return fmt.Errorf("capacity decision requires its exact typed effect")
			}
			gotPlacement,okPlacement:=effects[0].Placement()
			gotMaximum,okMaximum:=effects[0].Maximum()
			if !okPlacement || !okMaximum || gotPlacement!=placement || gotMaximum!=maximum {
				return fmt.Errorf("capacity effect differs from policy instance parameters")
			}
		} else if len(d.Effects())!=0 {
			return fmt.Errorf("unexpected effect on baseline allow/deny")
		}
	case domain.PolicyDecisionRequireReview:
		effects:=d.Effects()
		if instance.EffectClass()!=domain.PolicyEffectReviewRequired || len(effects)!=1 ||
			effects[0].Type()!=domain.PolicyEffectRequireReview || effects[0].Code()!=d.ReasonCode() {
			return fmt.Errorf("review requirement must retain its matching typed effect")
		}
	case domain.PolicyDecisionIndeterminate:
		// Missing-input interpretation occurs below, not as an evaluator failure.
	default:
		return fmt.Errorf("unsupported baseline result %q",d.Result())
	}
	return nil
}

func interpretBaselineDecision(instance domain.PolicyInstance,d domain.PolicyDecision) (string,error) {
	switch d.Result() {
	case domain.PolicyDecisionNotApplicable:
		return "not_applicable",nil
	case domain.PolicyDecisionAllow:
		if instance.EffectClass()==domain.PolicyEffectHard {
			return "hard_allow",nil
		}
		if instance.EffectClass()==domain.PolicyEffectReviewRequired {
			return "review_gate_passed",nil
		}
		return "",fmt.Errorf("unsupported allow effect class %q",instance.EffectClass())
	case domain.PolicyDecisionDeny:
		return "hard_deny",nil
	case domain.PolicyDecisionRequireReview:
		return "review_required",nil
	case domain.PolicyDecisionIndeterminate:
		switch d.MissingInputBehavior() {
		case domain.PolicyMissingInputDeny:
			return "missing_input_deny",nil
		case domain.PolicyMissingInputExcludeSubject:
			return "missing_input_exclude_subject",nil
		case domain.PolicyMissingInputRequireReview:
			return "missing_input_require_review",nil
		case domain.PolicyMissingInputFailOperation:
			return "missing_input_fail_operation",nil
		case domain.PolicyMissingInputDiagnosticOnly:
			if instance.EffectClass()!=domain.PolicyEffectAdvisory {
				return "",fmt.Errorf("diagnostic_only requires advisory effect class")
			}
			return "missing_input_diagnostic_only",nil
		}
	}
	return "",fmt.Errorf("unsupported result/behavior combination")
}
