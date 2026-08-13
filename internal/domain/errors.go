package domain

import "fmt"

func errZeroTime(field string) error {
	return fmt.Errorf("%s must not be zero", field)
}

func errSelfReference(field string) error {
	return fmt.Errorf("%s must not reference itself", field)
}

func errMaintainerAuthority(action string) error {
	return fmt.Errorf("%s requires maintainer authority", action)
}

func errProjectVersionMismatch(projectID ProjectID, versionID ProjectVersionID) error {
	return fmt.Errorf("project version %q does not belong to project %q", versionID, projectID)
}

func errInvalidPlacement(placement Placement) error {
	return fmt.Errorf("invalid placement %q", placement)
}

func errInvalidDisposition(disposition Disposition) error {
	return fmt.Errorf("invalid orientation disposition %q", disposition)
}

func errTraceOperationMismatch() error {
	return fmt.Errorf("orientation run operation id must match trace operation id")
}

func errDuplicateProjectRecommendation(projectID ProjectID) error {
	return fmt.Errorf("project %q has more than one placement recommendation", projectID)
}

func errDuplicateProjectOverride(projectID ProjectID) error {
	return fmt.Errorf("project %q has more than one placement override", projectID)
}

func errOverridesRequired() error {
	return fmt.Errorf("accepted_with_overrides disposition requires at least one override")
}

func errOverridesNotAllowed(disposition Disposition) error {
	return fmt.Errorf("%s disposition must not contain placement overrides", disposition)
}

func errSupersessionNotAllowed(disposition Disposition) error {
	return fmt.Errorf("%s disposition must not supersede an accepted disposition", disposition)
}
