package providers

import (
	"context"
	"errors"
	"net/netip"
)

const MaxRequestBytes = 15 << 20

var (
	ErrUnavailable  = errors.New("provider profile is unavailable")
	ErrDisabled     = errors.New("provider profile is disabled")
	ErrPolicyDenied = errors.New("provider destination is not approved")
	ErrCredential   = errors.New("provider credential is unavailable")
	ErrRedirect     = errors.New("provider redirect rejected")
)

func IsPolicyError(err error) bool {
	return errors.Is(err, ErrPolicyDenied)
}

type DispatchRequest struct {
	OwnerID   string
	ProfileID string
	Revision  int
	Body      []byte
}

type DispatchResult struct {
	StatusCode int
	Body       []byte
}

type ProfileVersion struct {
	OwnerID   string
	ProfileID string
	Revision  int
	Name      string
	BaseURL   string
	Model     string
	Enabled   bool
	HasSecret bool
}

type DispatchAdmission struct {
	Authorization string
}

type PinnedDestination struct {
	Rule      EgressRule
	Canonical CanonicalURL
	Addr      netip.Addr
}

type ProfileStore interface {
	LoadDispatchVersion(ctx context.Context, ownerID, profileID string, revision int) (ProfileVersion, error)
	AdmitDispatch(ctx context.Context, ownerID, profileID string, revision int) (DispatchAdmission, error)
}

type Transport interface {
	Pin(ctx context.Context, rule EgressRule) (PinnedDestination, error)
	Transmit(ctx context.Context, dest PinnedDestination, body []byte, authorization string) (DispatchResult, error)
}

type Dispatcher struct {
	Enabled      bool
	Policy       EgressPolicy
	Store        ProfileStore
	Transport    Transport
	AfterResolve func(context.Context)
}

func (d *Dispatcher) Dispatch(ctx context.Context, req DispatchRequest) (DispatchResult, error) {
	if !d.Enabled {
		return DispatchResult{}, ErrDisabled
	}
	if req.OwnerID == "" || req.ProfileID == "" || req.Revision < 1 {
		return DispatchResult{}, ErrUnavailable
	}
	if len(req.Body) > MaxRequestBytes {
		return DispatchResult{}, NewTransportFailure(FailureLimit, NewRequestID(), 0, nil)
	}
	version, err := d.Store.LoadDispatchVersion(ctx, req.OwnerID, req.ProfileID, req.Revision)
	if err != nil {
		return DispatchResult{}, err
	}
	if !version.Enabled {
		return DispatchResult{}, ErrDisabled
	}
	rule, err := MatchEgressDestination(d.Policy, version.BaseURL)
	if err != nil {
		return DispatchResult{}, ErrPolicyDenied
	}
	dest, err := d.Transport.Pin(ctx, rule)
	if err != nil {
		return DispatchResult{}, err
	}
	if d.AfterResolve != nil {
		d.AfterResolve(ctx)
	}
	if !d.Enabled {
		return DispatchResult{}, ErrDisabled
	}
	admission, err := d.Store.AdmitDispatch(ctx, req.OwnerID, req.ProfileID, req.Revision)
	if err != nil {
		return DispatchResult{}, err
	}
	return d.Transport.Transmit(ctx, dest, req.Body, admission.Authorization)
}
