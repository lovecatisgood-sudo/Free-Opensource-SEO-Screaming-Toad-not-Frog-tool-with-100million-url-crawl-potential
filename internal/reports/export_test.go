package reports

import "testing"

func TestSpreadsheetSafeEscapesFormulaPrefixes(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"=cmd()", "+1", "-2", "@SUM(A1)", "  =hidden"} {
		if got := spreadsheetSafe(value); got[0] != '\'' {
			t.Errorf("spreadsheetSafe(%q) = %q", value, got)
		}
	}
	if got := spreadsheetSafe("ordinary"); got != "ordinary" {
		t.Fatalf("ordinary = %q", got)
	}
}
