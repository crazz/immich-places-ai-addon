package providerhttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"immich-places-backend/internal/ai/providers"
)

const (
	userAgent        = "immich-places-ai-provider/1.0"
	maxResponseBytes = 1 << 20
	maxHeaderBytes   = 32 << 10
	defaultTimeout   = 120 * time.Second
)

type LookupIPFunc func(ctx context.Context, host string) ([]net.IP, error)

type Options struct {
	LookupIP LookupIPFunc
	RootCAs  *x509.CertPool
	// MaxResponseBytes tightens the response ceiling for one client; zero keeps the general default.
	MaxResponseBytes int
}

type Client struct {
	lookup      LookupIPFunc
	rootCAs     *x509.CertPool
	maxResponse int
}

func New(opts Options) *Client {
	lookup := opts.LookupIP
	if lookup == nil {
		lookup = defaultLookupIP
	}
	maxResponse := opts.MaxResponseBytes
	if maxResponse <= 0 {
		maxResponse = maxResponseBytes
	}
	return &Client{lookup: lookup, rootCAs: opts.RootCAs, maxResponse: maxResponse}
}

func defaultLookupIP(ctx context.Context, host string) ([]net.IP, error) {
	return net.DefaultResolver.LookupIP(ctx, "ip", host)
}

func (c *Client) Send(ctx context.Context, rule providers.EgressRule, body []byte) (providers.DispatchResult, error) {
	if len(body) > providers.MaxRequestBytes {
		return providers.DispatchResult{}, limitFailure(providers.NewRequestID())
	}
	dest, err := c.Pin(ctx, rule)
	if err != nil {
		return providers.DispatchResult{}, err
	}
	return c.Transmit(ctx, dest, body, "")
}

func (c *Client) Pin(ctx context.Context, rule providers.EgressRule) (providers.PinnedDestination, error) {
	requestID := providers.NewRequestID()
	canonical, err := providers.CanonicalBaseURL(rule.BaseURL)
	if err != nil {
		return providers.PinnedDestination{}, providers.ErrPolicyDenied
	}
	ips, err := c.lookup(ctx, canonical.Host)
	if err != nil {
		return providers.PinnedDestination{}, mapRequestError(requestID, err)
	}
	addrs := make([]netip.Addr, 0, len(ips))
	for _, ip := range ips {
		addr, ok := netip.AddrFromSlice(ip)
		if !ok {
			return providers.PinnedDestination{}, providers.NewTransportFailure(providers.FailureConnection, requestID, 0, fmt.Errorf("destination resolution failed"))
		}
		addrs = append(addrs, addr)
	}
	approved, err := providers.ValidateResolvedAddresses(rule, addrs)
	if err != nil {
		return providers.PinnedDestination{}, providers.NewTransportFailure(providers.FailureConnection, requestID, 0, err)
	}
	return providers.PinnedDestination{Rule: rule, Canonical: canonical, Addr: approved[0]}, nil
}

func (c *Client) Transmit(ctx context.Context, dest providers.PinnedDestination, body []byte, authorization string) (providers.DispatchResult, error) {
	requestID := providers.NewRequestID()
	if len(body) > providers.MaxRequestBytes {
		return providers.DispatchResult{}, limitFailure(requestID)
	}
	return c.doPinnedRequest(ctx, dest.Canonical, dest.Addr, body, authorization, requestID)
}

func (c *Client) doPinnedRequest(ctx context.Context, canonical providers.CanonicalURL, addr netip.Addr, body []byte, authorization, requestID string) (providers.DispatchResult, error) {
	deadline := time.Now().Add(defaultTimeout)
	if existing, ok := ctx.Deadline(); !ok || existing.After(deadline) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, deadline)
		defer cancel()
	}
	operationPath := strings.TrimRight(canonical.Path, "/") + "/chat/completions"
	if canonical.Path == "/" {
		operationPath = "/chat/completions"
	}
	target := &url.URL{Scheme: canonical.Scheme, Host: net.JoinHostPort(addr.String(), canonical.Port), Path: operationPath}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), bytes.NewReader(body))
	if err != nil {
		return providers.DispatchResult{}, mapRequestError(requestID, err)
	}
	req.Host = canonical.Host
	if canonical.Port != "80" && canonical.Port != "443" {
		req.Host = net.JoinHostPort(canonical.Host, canonical.Port)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}

	transport := &http.Transport{
		Proxy: func(*http.Request) (*url.URL, error) { return nil, nil },
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			dialer := &net.Dialer{Timeout: 30 * time.Second}
			return dialer.DialContext(ctx, "tcp", net.JoinHostPort(addr.String(), canonical.Port))
		},
		DisableKeepAlives:      true,
		DisableCompression:     true,
		ForceAttemptHTTP2:      false,
		MaxResponseHeaderBytes: int64(maxHeaderBytes),
		ResponseHeaderTimeout:  60 * time.Second,
		TLSClientConfig: &tls.Config{
			ServerName: canonical.Host,
			MinVersion: tls.VersionTLS12,
			RootCAs:    c.rootCAs,
		},
	}
	client := &http.Client{
		Transport: transport,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return providers.ErrRedirect
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		if isResponseHeaderLimitError(err) {
			return providers.DispatchResult{}, limitFailure(requestID)
		}
		return providers.DispatchResult{}, mapRequestError(requestID, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		return providers.DispatchResult{}, providers.ErrRedirect
	}
	if responseHeaderSize(resp.Header) > maxHeaderBytes {
		return providers.DispatchResult{}, limitFailure(requestID)
	}
	encoding := resp.Header.Get("Content-Encoding")
	if encoding != "" && !strings.EqualFold(encoding, "identity") {
		return providers.DispatchResult{}, providers.NewTransportFailure(providers.FailureEncoding, requestID, 0, nil)
	}
	limited := io.LimitReader(resp.Body, int64(c.maxResponse)+1)
	payload, err := io.ReadAll(limited)
	if err != nil {
		return providers.DispatchResult{}, mapRequestError(requestID, err)
	}
	if len(payload) > c.maxResponse {
		return providers.DispatchResult{}, limitFailure(requestID)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return providers.DispatchResult{}, upstreamFailure(requestID, resp.StatusCode, payload)
	}
	return providers.DispatchResult{StatusCode: resp.StatusCode, Body: payload}, nil
}

func responseHeaderSize(header http.Header) int {
	size := 0
	for name, values := range header {
		for _, value := range values {
			size += len(name) + len(value) + 4
		}
	}
	return size
}
