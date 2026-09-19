package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestRevocationWhileResolutionIsPending(t *testing.T) {
	for _, tc := range []struct {
		name     string
		mutate   func(t *testing.T, db *Database, dispatcher *providers.Dispatcher, profileID string, revision int)
		want     error
		wantAuth string
		wantHits int
	}{
		{
			name: "installation-disablement",
			mutate: func(t *testing.T, _ *Database, dispatcher *providers.Dispatcher, _ string, _ int) {
				t.Helper()
				dispatcher.Enabled = false
			},
			want: providers.ErrDisabled,
		},
		{
			name: "profile-disablement",
			mutate: func(t *testing.T, db *Database, _ *providers.Dispatcher, profileID string, _ int) {
				t.Helper()
				if _, err := db.db.ExecContext(context.Background(), `UPDATE ai_provider_profiles SET enabled=0 WHERE userID=? AND id=?`, testUserID, profileID); err != nil {
					t.Fatal(err)
				}
			},
			want: providers.ErrDisabled,
		},
		{
			name: "account-deletion",
			mutate: func(t *testing.T, db *Database, _ *providers.Dispatcher, _ string, _ int) {
				t.Helper()
				if _, err := db.db.ExecContext(context.Background(), `DELETE FROM users WHERE ID=?`, testUserID); err != nil {
					t.Fatal(err)
				}
			},
			want: providers.ErrUnavailable,
		},
		{
			name: "credential-removal",
			mutate: func(t *testing.T, db *Database, _ *providers.Dispatcher, profileID string, revision int) {
				t.Helper()
				empty := ""
				var baseURL string
				if err := db.db.QueryRowContext(context.Background(), `SELECT baseURL FROM ai_provider_versions WHERE userID=? AND profileID=? AND revision=?`, testUserID, profileID, revision).Scan(&baseURL); err != nil {
					t.Fatal(err)
				}
				input := providers.Input{
					Config:  providers.Config{Name: "Race", BaseURL: baseURL, Model: "vision"},
					Enabled: true,
					Secret:  &empty,
				}
				if _, err := db.updateAIProvider(context.Background(), testUserID, profileID, revision, input); err != nil {
					t.Fatal(err)
				}
			},
			wantAuth: "",
			wantHits: 1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hits := 0
			var lastAuth string
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits++
				lastAuth = r.Header.Get("Authorization")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"ok":true}`))
			}))
			server.Listener = listener
			server.Start()
			t.Cleanup(server.Close)

			base, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			baseURL := "http://127.0.0.1:" + base.Port() + "/v1"
			policy, err := providers.ParseEgressPolicy(`[{"baseURL":"` + baseURL + `","addressClass":"local","allowedCIDRs":["127.0.0.0/8"]}]`)
			if err != nil {
				t.Fatal(err)
			}
			db := newTestDB(t)
			ctx := context.Background()
			profile, err := db.createAIProvider(ctx, testUserID, "race-"+tc.name, providers.Input{
				Config:  providers.Config{Name: "Race", BaseURL: baseURL, Model: "vision"},
				Enabled: true,
				Secret:  ptr("race-secret-must-not-leak"),
			})
			if err != nil {
				t.Fatal(err)
			}

			resolved := make(chan struct{})
			admit := make(chan struct{})
			transport := providerhttp.New(providerhttp.Options{
				LookupIP: func(context.Context, string) ([]net.IP, error) {
					return []net.IP{net.ParseIP("127.0.0.1")}, nil
				},
			})
			dispatcher := newAIProviderDispatcher(db, true, policy, transport)
			dispatcher.AfterResolve = func(context.Context) {
				close(resolved)
				<-admit
			}

			var result providers.DispatchResult
			var dispatchErr error
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				result, dispatchErr = dispatcher.Dispatch(ctx, providers.DispatchRequest{
					OwnerID: testUserID, ProfileID: profile.ID, Revision: profile.Revision, Body: []byte(`{}`),
				})
			}()

			<-resolved
			tc.mutate(t, db, dispatcher, profile.ID, profile.Revision)
			close(admit)
			wg.Wait()

			if tc.want != nil {
				if !errors.Is(dispatchErr, tc.want) || hits != 0 {
					t.Fatalf("hits=%d err=%v want=%v", hits, dispatchErr, tc.want)
				}
				return
			}
			if dispatchErr != nil || hits != tc.wantHits || lastAuth != tc.wantAuth || result.StatusCode != http.StatusOK {
				t.Fatalf("credential-free after removal: hits=%d auth=%q result=%+v err=%v", hits, lastAuth, result, dispatchErr)
			}
		})
	}
}
