package domain

const PolicyConfigurationSchemaVersionV1 = "1"

type PolicyWorkflow string

const PolicyWorkflowOrientation PolicyWorkflow = "orientation"

func (w PolicyWorkflow) Valid() bool {
	return w == PolicyWorkflowOrientation
}

type PolicyApplicability string

const (
	PolicyApplicable    PolicyApplicability = "applicable"
	PolicyNotApplicable PolicyApplicability = "not_applicable"
)

func (a PolicyApplicability) Valid() bool {
	return a == PolicyApplicable || a == PolicyNotApplicable
}

// ConfigurationSchemaVersion and Workflow are fixed by the v1 constructor
// surface. Future schema/workflow support requires a new validated domain path
// rather than silently changing the meaning of an existing instance.
func (p PolicyInstance) ConfigurationSchemaVersion() string {
	return PolicyConfigurationSchemaVersionV1
}

func (p PolicyInstance) Workflow() PolicyWorkflow {
	return PolicyWorkflowOrientation
}

func (d PolicyDecision) ConfigurationSchemaVersion() string {
	return PolicyConfigurationSchemaVersionV1
}

func (d PolicyDecision) Workflow() PolicyWorkflow {
	return PolicyWorkflowOrientation
}

func (d PolicyDecision) Applicability() PolicyApplicability {
	if d.Result() == PolicyDecisionNotApplicable {
		return PolicyNotApplicable
	}
	return PolicyApplicable
}
