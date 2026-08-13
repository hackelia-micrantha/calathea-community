package domain

import (
	"strings"
	"testing"
)

func FuzzNewPlacementRecommendation(f *testing.F) {
	seeds := [][2]string{
		{"project-1", string(PlacementNow)},
		{"project-1", string(PlacementNext)},
		{"project-1", string(PlacementLater)},
		{"project-1", string(PlacementKill)},
		{"", string(PlacementNow)},
		{"   ", string(PlacementNext)},
		{"project-1", "invalid"},
	}
	for _, seed := range seeds {
		f.Add(seed[0], seed[1])
	}

	f.Fuzz(func(t *testing.T, projectIDValue, placementValue string) {
		projectID := ProjectID(projectIDValue)
		placement := Placement(placementValue)
		recommendation, err := NewPlacementRecommendation(projectID, placement)

		wantValid := strings.TrimSpace(projectIDValue) != "" && placement.Valid()
		if !wantValid {
			if err == nil {
				t.Fatalf("NewPlacementRecommendation(%q, %q) accepted invalid input", projectIDValue, placementValue)
			}
			return
		}

		if err != nil {
			t.Fatalf("NewPlacementRecommendation(%q, %q) returned unexpected error: %v", projectIDValue, placementValue, err)
		}
		if recommendation.ProjectID() != projectID {
			t.Fatalf("ProjectID() = %q, want %q", recommendation.ProjectID(), projectID)
		}
		if recommendation.Placement() != placement {
			t.Fatalf("Placement() = %q, want %q", recommendation.Placement(), placement)
		}
	})
}
