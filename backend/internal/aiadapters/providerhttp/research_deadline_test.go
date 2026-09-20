package providerhttp

import (
	"context"
	"immich-places-backend/internal/ai/providers"
	"net"
	"net/netip"
	"strconv"
	"testing"
	"time"
)

func TestResearchResponseWaitPreservesTheFiniteParentDeadline(t *testing.T) {
	deadline := time.Now().Add(9 * time.Minute)
	parent, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	ctx, finish, wait := responseContext(parent, 10*time.Minute)
	defer finish()
	got, ok := ctx.Deadline()
	if !ok || !got.Equal(deadline) || wait <= 8*time.Minute || wait > 9*time.Minute {
		t.Fatal("Research response wait shortened or extended", wait)
	}
	cancel()
	if ctx.Err() != context.Canceled {
		t.Fatal("parent cancellation lost")
	}
}

func TestResearchTransportKeepsAShortTLSHandshakeLimit(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	ready := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			ready <- conn
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	done := make(chan error, 1)
	start := time.Now()
	go func() {
		_, err := New(Options{}).doPinnedRequest(ctx, providers.CanonicalURL{Scheme: "https", Host: "example.org", Port: strconv.Itoa(listener.Addr().(*net.TCPAddr).Port), Path: "/v1"}, netip.MustParseAddr("127.0.0.1"), []byte(`{}`), "", "synthetic-tls")
		done <- err
	}()
	select {
	case conn := <-ready:
		defer conn.Close()
	case <-ctx.Done():
		t.Fatal("TLS connection not established")
	}
	err = <-done
	if err == nil || time.Since(start) > 11*time.Second {
		t.Fatal("TLS handshake consumed the long response deadline", time.Since(start), err)
	}
}
