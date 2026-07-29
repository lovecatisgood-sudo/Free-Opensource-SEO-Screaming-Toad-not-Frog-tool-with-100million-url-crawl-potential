package fetchpolicy

import (
	"context"
	"fmt"
)

type ValidatedTarget struct {
	NormalizedURL
	Resolution Resolution
}

type TargetValidator interface {
	Validate(context.Context, string) (ValidatedTarget, error)
}

type Guard struct {
	Resolver Resolver
	Scope    *Scope
}

func (g *Guard) Validate(ctx context.Context, raw string) (ValidatedTarget, error) {
	normalized, err := NormalizeURL(raw)
	if err != nil {
		return ValidatedTarget{}, fmt.Errorf("normalize target: %w", err)
	}
	if g.Scope == nil {
		return ValidatedTarget{}, fmt.Errorf("crawl scope is required")
	}
	if err := g.Scope.Evaluate(normalized); err != nil {
		return ValidatedTarget{}, err
	}
	resolution, err := ResolvePublic(ctx, g.Resolver, normalized.URL.Hostname())
	if err != nil {
		return ValidatedTarget{}, err
	}
	return ValidatedTarget{NormalizedURL: normalized, Resolution: resolution}, nil
}
