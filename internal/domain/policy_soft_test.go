package domain

import "testing"

func TestScoreMultiplierContractRequiresSupportedExactBounds(t *testing.T) {
	if _, err := NewScoreMultiplierContract("unknown", 5000, 20000, 5000, 20000); err == nil {
		t.Fatal("unsupported combinator accepted")
	}
	if _, err := NewScoreMultiplierContract(PolicySoftCombinatorMultiplyV1, 5000, 20000, 11000, 20000); err == nil {
		t.Fatal("cumulative bounds excluding neutral accepted")
	}
	contract, err := NewScoreMultiplierContract(PolicySoftCombinatorMultiplyV1, 5000, 20000, 5000, 20000)
	if err != nil {
		t.Fatal(err)
	}
	if contract.Combinator() != PolicySoftCombinatorMultiplyV1 ||
		contract.CumulativeMinBasisPoints() != 5000 ||
		contract.CumulativeMaxBasisPoints() != 20000 {
		t.Fatalf("unexpected contract: %#v", contract)
	}
}

func TestScoreMultiplierParameterAndEffectRejectNonPositiveFactor(t *testing.T) {
	if _, err := NewScoreMultiplierParameters(0); err == nil {
		t.Fatal("zero parameter factor accepted")
	}
	if _, err := NewScoreMultiplierPolicyEffect(-1); err == nil {
		t.Fatal("negative effect factor accepted")
	}
}
