package policy

import (
	"fmt"
	"math/big"
)

// ExactMultiplier is a positive exact rational policy multiplier. Its internal
// big.Rat is never exposed directly, so callers cannot mutate composition state.
type ExactMultiplier struct {
	value *big.Rat
}

func neutralMultiplier() ExactMultiplier {
	return ExactMultiplier{value: new(big.Rat).SetInt64(1)}
}

func multiplierFromBasisPoints(value int) (ExactMultiplier, error) {
	if value <= 0 {
		return ExactMultiplier{}, fmt.Errorf("score multiplier basis points %d must be positive", value)
	}
	return ExactMultiplier{value: new(big.Rat).SetFrac(big.NewInt(int64(value)), big.NewInt(10000))}, nil
}

func (m ExactMultiplier) valid() bool {
	return m.value != nil && m.value.Sign() > 0
}

func (m ExactMultiplier) multiply(other ExactMultiplier) (ExactMultiplier, error) {
	if !m.valid() || !other.valid() {
		return ExactMultiplier{}, fmt.Errorf("invalid exact score multiplier")
	}
	return ExactMultiplier{value: new(big.Rat).Mul(new(big.Rat).Set(m.value), other.value)}, nil
}

func (m ExactMultiplier) compareBasisPoints(bound int) (int, error) {
	if !m.valid() || bound <= 0 {
		return 0, fmt.Errorf("invalid multiplier or bound")
	}
	b := new(big.Rat).SetFrac(big.NewInt(int64(bound)), big.NewInt(10000))
	return m.value.Cmp(b), nil
}

func (m ExactMultiplier) Numerator() *big.Int {
	if !m.valid() {
		return nil
	}
	return new(big.Int).Set(m.value.Num())
}

func (m ExactMultiplier) Denominator() *big.Int {
	if !m.valid() {
		return nil
	}
	return new(big.Int).Set(m.value.Denom())
}

func (m ExactMultiplier) String() string {
	if !m.valid() {
		return ""
	}
	return m.value.RatString()
}
