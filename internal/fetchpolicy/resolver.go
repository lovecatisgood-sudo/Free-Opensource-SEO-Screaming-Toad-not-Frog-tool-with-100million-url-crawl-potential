package fetchpolicy

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strings"
)

type Resolver interface {
	LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error)
}

type SystemResolver struct{ Resolver *net.Resolver }

func (r SystemResolver) LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error) {
	resolver := r.Resolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	return resolver.LookupNetIP(ctx, network, host)
}

type Resolution struct {
	Host      string
	Addresses []netip.Addr
}

func ResolvePublic(ctx context.Context, resolver Resolver, host string) (Resolution, error) {
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	if host == "" {
		return Resolution{}, fmt.Errorf("host is empty")
	}
	if literal, err := netip.ParseAddr(host); err == nil {
		decision := ClassifyIP(literal)
		if !decision.Allowed {
			return Resolution{}, fmt.Errorf("address blocked: %s", decision.Reason)
		}
		return Resolution{Host: host, Addresses: []netip.Addr{literal.Unmap()}}, nil
	}
	addresses, err := resolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return Resolution{}, fmt.Errorf("resolve host: %w", err)
	}
	if len(addresses) == 0 {
		return Resolution{}, fmt.Errorf("host has no addresses")
	}
	approved := make([]netip.Addr, 0, len(addresses))
	seen := make(map[netip.Addr]struct{}, len(addresses))
	for _, address := range addresses {
		address = address.Unmap()
		decision := ClassifyIP(address)
		if !decision.Allowed {
			// Mixed public/private answers are rejected as a unit.
			return Resolution{}, fmt.Errorf("host resolved to blocked address class: %s", decision.Reason)
		}
		if _, exists := seen[address]; !exists {
			approved = append(approved, address)
			seen[address] = struct{}{}
		}
	}
	return Resolution{Host: host, Addresses: approved}, nil
}
