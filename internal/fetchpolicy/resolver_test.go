package fetchpolicy

import (
	"context"
	"net/netip"
	"testing"
)

type staticResolver []netip.Addr

func (r staticResolver) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) {
	return r, nil
}

func TestResolvePublicAcceptsOnlyPublicSet(t *testing.T) {
	t.Parallel()

	got, err := ResolvePublic(context.Background(), staticResolver{
		netip.MustParseAddr("8.8.8.8"),
		netip.MustParseAddr("2606:4700:4700::1111"),
	}, "Example.COM.")
	if err != nil {
		t.Fatalf("ResolvePublic: %v", err)
	}
	if got.Host != "example.com" || len(got.Addresses) != 2 {
		t.Fatalf("unexpected resolution: %+v", got)
	}
}

func TestResolvePublicRejectsMixedAnswers(t *testing.T) {
	t.Parallel()

	_, err := ResolvePublic(context.Background(), staticResolver{
		netip.MustParseAddr("8.8.8.8"),
		netip.MustParseAddr("127.0.0.1"),
	}, "example.com")
	if err == nil {
		t.Fatal("expected mixed answer to be rejected")
	}
}

func TestResolvePublicRejectsMetadataAddress(t *testing.T) {
	t.Parallel()

	if _, err := ResolvePublic(context.Background(), staticResolver{}, "169.254.169.254"); err == nil {
		t.Fatal("expected metadata address to be rejected")
	}
}
