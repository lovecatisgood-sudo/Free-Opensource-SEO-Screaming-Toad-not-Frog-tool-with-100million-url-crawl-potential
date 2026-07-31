package googleprofiles

import "testing"

func TestBundledProfilesAreVersioned(t *testing.T) {
	bundle, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Metadata.Version != "2026-07-30" || bundle.Profiles[0].ID != "breadcrumb" {
		t.Fatalf("unexpected bundle: %+v", bundle)
	}
}
