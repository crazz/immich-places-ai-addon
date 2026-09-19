package providerhttp_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestPreserveDirectRoutingAndTLSVerification(t *testing.T) {
	proxyHits := 0
	proxy := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { proxyHits++ }))
	t.Cleanup(proxy.Close)
	t.Setenv("HTTP_PROXY", proxy.URL)
	t.Setenv("HTTPS_PROXY", proxy.URL)
	t.Setenv("http_proxy", proxy.URL)
	t.Setenv("https_proxy", proxy.URL)

	providerHits := 0
	cert, key := mustHostnameCert(t, "provider.example")
	tlsCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		providerHits++
		_, _ = w.Write([]byte(`{"ok":true}`))
	})}
	go server.Serve(tls.NewListener(listener, &tls.Config{Certificates: []tls.Certificate{tlsCert}}))
	t.Cleanup(func() { _ = server.Close() })

	port := strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
	baseURL := "https://provider.example:" + port + "/v1"
	policy, err := providers.ParseEgressPolicy(`[{"baseURL":"` + baseURL + `","addressClass":"local","allowedCIDRs":["127.0.0.0/8"]}]`)
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(cert) {
		t.Fatal("root cert")
	}
	client := providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
		RootCAs: roots,
	})
	_, err = client.Send(context.Background(), policy.Rules[0], []byte(`{"n":1}`))
	if err != nil {
		t.Fatalf("valid TLS dispatch failed: %v", err)
	}
	if proxyHits != 0 {
		t.Fatalf("environment proxy received provider traffic: %d", proxyHits)
	}
	if providerHits != 1 {
		t.Fatalf("providerHits=%d", providerHits)
	}

	badHits := 0
	bad := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { badHits++ }))
	t.Cleanup(bad.Close)
	badURL, _ := url.Parse(bad.URL)
	badBase := "https://provider.example:" + badURL.Port() + "/v1"
	badPolicy, err := providers.ParseEgressPolicy(`[{"baseURL":"` + badBase + `","addressClass":"local","allowedCIDRs":["127.0.0.0/8"]}]`)
	if err != nil {
		t.Fatal(err)
	}
	rootsBad := x509.NewCertPool()
	rootsBad.AddCert(bad.Certificate())
	badClient := providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
		RootCAs: rootsBad,
	})
	_, err = badClient.Send(context.Background(), badPolicy.Rules[0], []byte(`{"Authorization":"leak-me"}`))
	if err == nil || badHits != 0 || proxyHits != 0 {
		t.Fatalf("invalid TLS must fail closed: err=%v badHits=%d proxyHits=%d", err, badHits, proxyHits)
	}
}

func mustHostnameCert(t *testing.T, dnsName string) (certPEM, keyPEM []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: dnsName},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		DNSNames:     []string{dnsName},
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	return certPEM, keyPEM
}
