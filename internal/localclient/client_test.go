package localclient

import "testing"

func TestClientRejectsNonLoopbackOrigins(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"https://127.0.0.1:7331", "http://example.com:7331", "http://localhost:7331", "http://127.0.0.1:7331/path"} {
		if _, err := New(value); err == nil {
			t.Errorf("accepted %q", value)
		}
	}
	if _, err := New("http://127.0.0.1:7331"); err != nil {
		t.Fatal(err)
	}
}
