package fetchpolicy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"
)

type PinnedDialer struct {
	Resolver Resolver
	Dialer   net.Dialer
	// Dial, when set, receives only an approved numeric address. It exists for
	// deterministic fixture networking; production leaves it nil.
	Dial func(context.Context, string, string) (net.Conn, error)
}

func (d *PinnedDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, fmt.Errorf("unsupported network %q", network)
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("invalid dial address: %w", err)
	}
	if _, err := strconv.ParseUint(port, 10, 16); err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}
	resolution, err := ResolvePublic(ctx, d.Resolver, host)
	if err != nil {
		return nil, err
	}
	var lastErr error
	for _, approved := range resolution.Addresses {
		approvedAddress := net.JoinHostPort(approved.String(), port)
		var conn net.Conn
		if d.Dial != nil {
			conn, err = d.Dial(ctx, network, approvedAddress)
		} else {
			conn, err = d.Dialer.DialContext(ctx, network, approvedAddress)
		}
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("connect to approved addresses: %w", lastErr)
}

func NewHTTPTransport(resolver Resolver) *http.Transport {
	return NewHTTPTransportWithDial(resolver, nil)
}

func NewHTTPTransportWithDial(resolver Resolver, dial func(context.Context, string, string) (net.Conn, error)) *http.Transport {
	dialer := &PinnedDialer{
		Resolver: resolver,
		Dial:     dial,
		Dialer: net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		},
	}
	return &http.Transport{
		Proxy:                 nil,
		DisableCompression:    true,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   4,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 20 * time.Second,
		ExpectContinueTimeout: time.Second,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
	}
}
