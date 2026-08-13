package domain

import (
	"fmt"
	"sort"
)

const (
	PolicyConfigurationSchemaVersionV1 = "policy.instance.v1"
	PolicyDecisionSchemaVersionV1      = "policy.decision.v1"
)

type PolicyWorkflow string

const PolicyWorkflowOrientation PolicyWorkflow = "orientation"

func (w PolicyWorkflow) Valid() bool {
	return w == PolicyWorkflowOrientation
}

type PolicySubjectType string

const PolicySubjectProject PolicySubjectType = "project"

func (s PolicySubjectType) Valid() bool {
	return s == PolicySubjectProject
}

// PolicySubjectSelector is deliberately constrained in v0. The empty project
// set means all projects; otherwise it is an exact project-identity allow-list.
// No expression language, content search, regex, or runtime code is permitted.
type PolicySubjectSelector struct {
	projectIDs []ProjectID
}

func NewProjectPolicySubjectSelector(projectIDs []ProjectID) (PolicySubjectSelector, error) {
	seen := make(map[ProjectID]struct{}, len(projectIDs))
	values := make([]ProjectID, 0, len(projectIDs))
	for _, projectID := range projectIDs {
		if err := requireIdentifier("policy selector project id", string(projectID)); err != nil {
			return PolicySubjectSelector{}, err
		}
		if _, exists := seen[projectID]; exists {
			return PolicySubjectSelector{}, fmt.Errorf("policy selector project id %q is duplicated", projectID)
		}
		seen[projectID] = struct{}{}
		values = append(values, projectID)
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	return PolicySubjectSelector{projectIDs: values}, nil
}

func (s PolicySubjectSelector) ProjectIDs() []ProjectID {
	if len(s.projectIDs) == 0 {
		return nil
	}
	return append([]ProjectID(nil), s.projectIDs...)
}

func (s PolicySubjectSelector) MatchesProject(projectID ProjectID) bool {
	if len(s.projectIDs) == 0 {
		return true
	}
	index := sort.Search(len(s.projectIDs), func(i int) bool { return s.projectIDs[i] >= projectID })
	return index < len(s.projectIDs) && s.projectIDs[index] == projectID
}

// ConfigurationSchemaVersion is explicit even though v0 supports one schema.
// Future evaluator/configuration versions must not silently reinterpret v1 data.
func (p PolicyInstance) ConfigurationSchemaVersion() string {
	return PolicyConfigurationSchemaVersionV1
}

// Workflow is intentionally narrow for v0 while retaining an explicit contract.
func (p PolicyInstance) Workflow() PolicyWorkflow {
	return PolicyWorkflowOrientation
}

func (p PolicyInstance) SubjectType() PolicySubjectType {
	return PolicySubjectProject
}

func (p PolicyInstance) SubjectSelector() PolicySubjectSelector {
	return PolicySubjectSelector{}
}

// RequiredInputs provides the explicit preflight/explanation declaration for
// each supported evaluator. Derived values name the canonical input they depend
// on rather than introducing ambient lookup behavior.
func (p PolicyInstance) RequiredInputs() []string {
	switch p.EvaluatorType() {
	case PolicyEvaluatorCapacityLimit:
		return []string{"placement", "selected_count"}
	case PolicyEvaluatorRequiredEvaluation:
		return []string{"accepted_evaluation"}
	case PolicyEvaluatorLifecycleEligibility:
		return []string{"lifecycle_state"}
	case PolicyEvaluatorConfidenceGate:
		return []string{"accepted_evaluation", "confidence_band"}
	case PolicyEvaluatorFreshnessRule:
		return []string{"accepted_evaluation", "evaluation_evidence_as_of", "planning_horizon", "policy_evaluation_time"}
	case PolicyEvaluatorScoreMultiplier:
		return []string{"accepted_evaluation", "base_score"}
	default:
		return nil
	}
}

// ConflictKey identifies configuration conflicts that can be rejected at
// policy-set activation without operation data. Empty means no exclusive key.
func (p PolicyInstance) ConflictKey() string {
	if parameters, ok := p.CapacityLimitParameters(); ok {
		return "capacity:" + string(parameters.Placement())
	}
	return ""
}
