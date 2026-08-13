package domain

import "testing"

func TestPolicySetVersionRequiresNeutralMultiplierWithinBounds(t *testing.T) {
	set, err := NewPolicySet("policy-set")
	if err != nil {
		t.Fatalf("NewPolicySet() error = %v", err)
	}
	instance := mustPolicyInstance(t, "required-evaluation", 10)

	for _, tc := range []struct {
		name string
		min  int
		max  int
	}{
		{name: "neutral below minimum", min: 10001, max: 12000},
		{name: "neutral above maximum", min: 8000, max: 9999},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewPolicySetVersion(PolicySetVersionInput{
				ID:                             PolicySetVersionID("policy-set-" + tc.name),
				PolicySet:                      set,
				CreatedAt:                      testTime(),
				Instances:                      []PolicyInstance{instance},
				MinimumScoreMultiplierBasisPts: tc.min,
				MaximumScoreMultiplierBasisPts: tc.max,
			})
			if err == nil {
				t.Fatalf("NewPolicySetVersion() accepted multiplier bounds %d-%d excluding neutral 10000", tc.min, tc.max)
			}
		})
	}
}
