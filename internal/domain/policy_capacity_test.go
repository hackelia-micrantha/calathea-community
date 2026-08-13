package domain

import (
	"strconv"
	"testing"
)

func TestCapacityLimitRejectsUint32Overflow(t *testing.T) {
	if strconv.IntSize < 64 {
		t.Skip("host int cannot represent a value above uint32")
	}
	tooLarge := int(uint64(^uint32(0)) + 1)
	if _, err := NewCapacityLimitParameters(PlacementNow, tooLarge); err == nil {
		t.Fatalf("NewCapacityLimitParameters() accepted %d above uint32 range", tooLarge)
	}
}
