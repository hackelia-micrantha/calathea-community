package policy

import (
	"testing"

	"github.com/hackelia-micrantha/calathea-community/internal/domain"
)

func TestValidateForActivationRejectsDiagnosticOnlyNonAdvisoryPolicy(t *testing.T) {
	instances := baselineInstances(t)
	for i, original := range instances {
		if original.PolicyID() != PolicyConfidence {
			continue
		}
		replacement, err := domain.NewPolicyInstance(domain.PolicyInstanceInput{
			ID:                   original.ID(),
			PolicyID:             original.PolicyID(),
			EvaluatorType:        original.EvaluatorType(),
			EvaluatorVersion:     original.EvaluatorVersion(),
			Phase:                original.Phase(),
			EffectClass:          original.EffectClass(),
			SubjectType:          original.SubjectType(),
			RequiredInputs:       original.RequiredInputs(),
			MissingInputBehavior: domain.PolicyMissingInputDiagnosticOnly,
			Priority:             original.Priority(),
			Exceptionability:     original.Exceptionability(),
			Parameters:           original.Parameters(),
			Rationale:            original.Rationale(),
		})
		if err != nil {
			t.Fatalf("NewPolicyInstance() error = %v", err)
		}
		instances[i] = replacement
		break
	}
	version, err := domain.NewPolicySetVersion(
		domain.PolicySetVersionID("policy-set-diagnostic-only-invalid"),
		domain.PolicySetID("policy-set"),
		policyTestTime(),
		instances...,
	)
	if err != nil {
		t.Fatalf("NewPolicySetVersion() error = %v", err)
	}
	if err := ValidateForActivation(version); err == nil {
		t.Fatal("non-advisory policy with diagnostic_only missing-input behavior activated")
	}
}
