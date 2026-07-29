package contracts

import "testing"

func TestDefaultCrawlLimitsAreValid(t *testing.T) {
	t.Parallel()
	if err := DefaultCrawlLimits().Validate(); err != nil {
		t.Fatalf("default limits: %v", err)
	}
}

func TestCrawlTerminalStates(t *testing.T) {
	t.Parallel()
	for _, status := range []CrawlStatus{CrawlCancelled, CrawlCompleted, CrawlFailed, CrawlLimited} {
		if !status.Terminal() {
			t.Errorf("%s should be terminal", status)
		}
	}
	if CrawlRunning.Terminal() {
		t.Fatal("running should not be terminal")
	}
}
