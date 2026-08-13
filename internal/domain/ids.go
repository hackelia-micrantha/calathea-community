package domain

import (
	"fmt"
	"strings"
)

// Distinct identifier types prevent accidental cross-entity substitution while
// leaving identity generation to the application boundary.
type (
	PortfolioID                  string
	ProjectID                    string
	ProjectVersionID             string
	EvaluationID                 string
	EvaluationVersionID          string
	PolicySetID                  string
	PolicySetVersionID           string
	PolicySelectionID            string
	PolicyID                     string
	PolicyDecisionID             string
	PolicyExceptionID            string
	PolicyExceptionApplicationID string
	OrientationRunID             string
	OrientationDispositionID     string
	LifecycleDecisionID          string
	EvidenceReferenceID          string
	OperationID                  string
	ActorID                      string
)

func requireIdentifier(kind, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s must not be empty", kind)
	}
	return nil
}

func requireText(kind, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s must not be empty", kind)
	}
	return nil
}
