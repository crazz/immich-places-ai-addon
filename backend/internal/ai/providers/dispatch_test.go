package providers

import (
	"context"
	"net/netip"
	"sync/atomic"
	"testing"
)

type countingStore struct {
	loads  atomic.Int64
	admits atomic.Int64
}

func (s *countingStore) LoadDispatchVersion(context.Context, string, string, int) (ProfileVersion, error) {
	s.loads.Add(1)
	return ProfileVersion{OwnerID: "o", ProfileID: "p", Revision: 1, BaseURL: "http://127.0.0.1:9/v1", Enabled: true}, nil
}

func (s *countingStore) AdmitDispatch(context.Context, string, string, int) (DispatchAdmission, error) {
	s.admits.Add(1)
	return DispatchAdmission{Authorization: "Bearer must-not-release"}, nil
}

type countingTransport struct {
	pins      atomic.Int64
	transmits atomic.Int64
}

func (t *countingTransport) Pin(context.Context, EgressRule) (PinnedDestination, error) {
	t.pins.Add(1)
	return PinnedDestination{Addr: netip.MustParseAddr("127.0.0.1")}, nil
}

func (t *countingTransport) Transmit(context.Context, PinnedDestination, []byte, string) (DispatchResult, error) {
	t.transmits.Add(1)
	return DispatchResult{StatusCode: 200}, nil
}

func TestDispatchRejectsOversizedBeforePinAndAdmit(t *testing.T) {
	store := &countingStore{}
	transport := &countingTransport{}
	dispatcher := &Dispatcher{
		Enabled:   true,
		Policy:    EgressPolicy{Rules: []EgressRule{{BaseURL: "http://127.0.0.1:9/v1", AddressClass: AddressClassLocal, AllowedCIDRs: []string{"127.0.0.0/8"}}}},
		Store:     store,
		Transport: transport,
	}
	_, err := dispatcher.Dispatch(context.Background(), DispatchRequest{
		OwnerID:   "owner",
		ProfileID: "profile",
		Revision:  1,
		Body:      make([]byte, MaxRequestBytes+1),
	})
	failure, ok := AsTransportFailure(err)
	if !ok || failure.Category != FailureLimit || failure.RequestID == "" {
		t.Fatalf("expected typed limit failure, got ok=%v %+v err=%v", ok, failure, err)
	}
	if store.loads.Load() != 0 || store.admits.Load() != 0 {
		t.Fatalf("oversized body reached store: loads=%d admits=%d", store.loads.Load(), store.admits.Load())
	}
	if transport.pins.Load() != 0 || transport.transmits.Load() != 0 {
		t.Fatalf("oversized body reached transport: pins=%d transmits=%d", transport.pins.Load(), transport.transmits.Load())
	}
}
