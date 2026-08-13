package domain

import (
	"strings"
	"testing"
)

func FuzzNewPlacementRecommendation(f *testing.F) {
	seeds := [][2]string{
		{"project-1", "now"},
		{"project-1", "next"},
		{"project-1", "later"},
		{"project-1", "kill"},
		{"", "now"},
		{"   ", "next"},
		{"project-1", "invalid"},
	}
	for _, seed := range seeds {
		f.Add(seed[0], seed[1])
	}

	f.Fuzz(func(t *testing.T, projectIDValue, placementValue string) {
		projectID := ProjectID(projectIDValue)
		placement := Placement(placementValue)
		recommendation, err := NewPlacementRecommendation(projectID, placement)

		placementValid := false
		switch placementValue {
		case "now", "next", "later", "kill":
			placementValid = true
		}
		wantValid := strings.TrimSpace(projectIDValue) != "" && placementValid

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
