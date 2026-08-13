package domain

import "time"

// Portfolio is the stable identity of one oriented project collection.
type Portfolio struct {
	id   PortfolioID
	name string
}

func NewPortfolio(id PortfolioID, name string) (Portfolio, error) {
	if err := requireIdentifier("portfolio id", string(id)); err != nil {
		return Portfolio{}, err
	}
	if err := requireText("portfolio name", name); err != nil {
		return Portfolio{}, err
	}
	return Portfolio{id: id, name: name}, nil
}

func (p Portfolio) ID() PortfolioID { return p.id }
func (p Portfolio) Name() string    { return p.name }

// Project is a stable identity. Mutable-looking project content lives in ProjectVersion.
type Project struct {
	id          ProjectID
	portfolioID PortfolioID
}

func NewProject(id ProjectID, portfolioID PortfolioID) (Project, error) {
	if err := requireIdentifier("project id", string(id)); err != nil {
		return Project{}, err
	}
	if err := requireIdentifier("portfolio id", string(portfolioID)); err != nil {
		return Project{}, err
	}
	return Project{id: id, portfolioID: portfolioID}, nil
}

func (p Project) ID() ProjectID            { return p.id }
func (p Project) PortfolioID() PortfolioID { return p.portfolioID }

// ProjectVersion contains authored project content required by the UC-01 foundation.
type ProjectVersion struct {
	id         ProjectVersionID
	projectID  ProjectID
	title      string
	createdAt  time.Time
	supersedes *ProjectVersionID
}

func NewProjectVersion(id ProjectVersionID, projectID ProjectID, title string, createdAt time.Time, supersedes *ProjectVersionID) (ProjectVersion, error) {
	if err := requireIdentifier("project version id", string(id)); err != nil {
		return ProjectVersion{}, err
	}
	if err := requireIdentifier("project id", string(projectID)); err != nil {
		return ProjectVersion{}, err
	}
	if err := requireText("project title", title); err != nil {
		return ProjectVersion{}, err
	}
	if createdAt.IsZero() {
		return ProjectVersion{}, errZeroTime("project version created at")
	}
	if supersedes != nil {
		if err := requireIdentifier("superseded project version id", string(*supersedes)); err != nil {
			return ProjectVersion{}, err
		}
		if *supersedes == id {
			return ProjectVersion{}, errSelfReference("project version supersedes")
		}
	}
	return ProjectVersion{id: id, projectID: projectID, title: title, createdAt: createdAt, supersedes: cloneProjectVersionID(supersedes)}, nil
}

func (v ProjectVersion) ID() ProjectVersionID          { return v.id }
func (v ProjectVersion) ProjectID() ProjectID          { return v.projectID }
func (v ProjectVersion) Title() string                 { return v.title }
func (v ProjectVersion) CreatedAt() time.Time          { return v.createdAt }
func (v ProjectVersion) Supersedes() *ProjectVersionID { return cloneProjectVersionID(v.supersedes) }

// Evaluation is the stable identity of a project's evaluation history. Evaluation
// axes and scoring are intentionally introduced by #26.
type Evaluation struct {
	id        EvaluationID
	projectID ProjectID
}

func NewEvaluation(id EvaluationID, projectID ProjectID) (Evaluation, error) {
	if err := requireIdentifier("evaluation id", string(id)); err != nil {
		return Evaluation{}, err
	}
	if err := requireIdentifier("project id", string(projectID)); err != nil {
		return Evaluation{}, err
	}
	return Evaluation{id: id, projectID: projectID}, nil
}

func (e Evaluation) ID() EvaluationID     { return e.id }
func (e Evaluation) ProjectID() ProjectID { return e.projectID }

// PolicySet is the stable identity of one policy configuration family.
type PolicySet struct {
	id PolicySetID
}

func NewPolicySet(id PolicySetID) (PolicySet, error) {
	if err := requireIdentifier("policy set id", string(id)); err != nil {
		return PolicySet{}, err
	}
	return PolicySet{id: id}, nil
}

func (p PolicySet) ID() PolicySetID { return p.id }

// PolicySetVersion identifies one immutable fully resolved policy configuration.
// Policy instances are introduced by #27.
type PolicySetVersion struct {
	id          PolicySetVersionID
	policySetID PolicySetID
	createdAt   time.Time
}

func NewPolicySetVersion(id PolicySetVersionID, policySetID PolicySetID, createdAt time.Time) (PolicySetVersion, error) {
	if err := requireIdentifier("policy set version id", string(id)); err != nil {
		return PolicySetVersion{}, err
	}
	if err := requireIdentifier("policy set id", string(policySetID)); err != nil {
		return PolicySetVersion{}, err
	}
	if createdAt.IsZero() {
		return PolicySetVersion{}, errZeroTime("policy set version created at")
	}
	return PolicySetVersion{id: id, policySetID: policySetID, createdAt: createdAt}, nil
}

func (v PolicySetVersion) ID() PolicySetVersionID   { return v.id }
func (v PolicySetVersion) PolicySetID() PolicySetID { return v.policySetID }
func (v PolicySetVersion) CreatedAt() time.Time     { return v.createdAt }

// PolicySelectionRecord is the implementation-level record that makes the current
// effective PolicySetVersion rebuildable. The RFC glossary intentionally does not
// require this Go type name.
type PolicySelectionRecord struct {
	id                 PolicySelectionID
	portfolioID        PortfolioID
	policySetVersionID PolicySetVersionID
	actor              Actor
	operationID        OperationID
	selectedAt         time.Time
	rationale          string
	supersedes         *PolicySelectionID
}

func NewPolicySelectionRecord(id PolicySelectionID, portfolioID PortfolioID, policySetVersionID PolicySetVersionID, actor Actor, operationID OperationID, selectedAt time.Time, rationale string, supersedes *PolicySelectionID) (PolicySelectionRecord, error) {
	if err := requireIdentifier("policy selection id", string(id)); err != nil {
		return PolicySelectionRecord{}, err
	}
	if err := requireIdentifier("portfolio id", string(portfolioID)); err != nil {
		return PolicySelectionRecord{}, err
	}
	if err := requireIdentifier("policy set version id", string(policySetVersionID)); err != nil {
		return PolicySelectionRecord{}, err
	}
	if !actor.IsMaintainer() {
		return PolicySelectionRecord{}, errMaintainerAuthority("policy selection")
	}
	if err := requireIdentifier("operation id", string(operationID)); err != nil {
		return PolicySelectionRecord{}, err
	}
	if selectedAt.IsZero() {
		return PolicySelectionRecord{}, errZeroTime("policy selection time")
	}
	if err := requireText("policy selection rationale", rationale); err != nil {
		return PolicySelectionRecord{}, err
	}
	if supersedes != nil {
		if err := requireIdentifier("superseded policy selection id", string(*supersedes)); err != nil {
			return PolicySelectionRecord{}, err
		}
		if *supersedes == id {
			return PolicySelectionRecord{}, errSelfReference("policy selection supersedes")
		}
	}
	return PolicySelectionRecord{
		id:                 id,
		portfolioID:        portfolioID,
		policySetVersionID: policySetVersionID,
		actor:              actor,
		operationID:        operationID,
		selectedAt:         selectedAt,
		rationale:          rationale,
		supersedes:         clonePolicySelectionID(supersedes),
	}, nil
}

func (r PolicySelectionRecord) ID() PolicySelectionID                  { return r.id }
func (r PolicySelectionRecord) PortfolioID() PortfolioID               { return r.portfolioID }
func (r PolicySelectionRecord) PolicySetVersionID() PolicySetVersionID { return r.policySetVersionID }
func (r PolicySelectionRecord) Actor() Actor                           { return r.actor }
func (r PolicySelectionRecord) OperationID() OperationID               { return r.operationID }
func (r PolicySelectionRecord) SelectedAt() time.Time                  { return r.selectedAt }
func (r PolicySelectionRecord) Rationale() string                      { return r.rationale }
func (r PolicySelectionRecord) Supersedes() *PolicySelectionID {
	return clonePolicySelectionID(r.supersedes)
}

func cloneProjectVersionID(value *ProjectVersionID) *ProjectVersionID {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}

func clonePolicySelectionID(value *PolicySelectionID) *PolicySelectionID {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}
