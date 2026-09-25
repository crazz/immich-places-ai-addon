package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"immich-places-backend/internal/aiadapters/providerhttp"
)

type contextKey string

const userContextKey contextKey = "user"

func getUserFromContext(r *http.Request) *UserRow {
	user, ok := r.Context().Value(userContextKey).(*UserRow)
	if !ok {
		return nil
	}
	return user
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("[Server] Failed to load config: %v", err)
	}

	db, err := newDatabase(cfg.DataDir, cfg.EncryptionKey)
	if err != nil {
		log.Fatalf("[Server] Failed to initialize database: %v", err)
	}
	defer db.close()
	if err := db.interruptRunningAIProviderCapabilities(context.Background()); err != nil {
		log.Fatalf("[Server] Failed to interrupt residual capability checks: %v", err)
	}

	immichFactory := newImmichClientFactory(cfg.ImmichURL, cfg.Debug)
	geocodeTimeout := time.Duration(cfg.GeocodeTimeoutSecs) * time.Second
	geocoder := newGeocodeProvider(cfg.GeocodeProvider, geocodeKeys{
		here:   cfg.HereAPIKey,
		google: cfg.GoogleAPIKey,
		legacy: cfg.GeocodeAPIKey,
	}, geocodeTimeout)
	log.Printf("[Geocode] Provider: %s (timeout: %v)", describeProvider(geocoder), geocodeTimeout)
	syncService := newSyncService(db, immichFactory, geocoder)
	suggestions := newSuggestionService(db, cfg.NeighborWindowHours)
	handlers := newHandlers(db, immichFactory, cfg.ImmichExternalURL, syncService, suggestions, cfg.defaultTimezoneLocation, geocoder)
	libraryHandlers := newLibraryHandlers(db, immichFactory, syncService)
	authHandlers := newAuthHandlers(db, immichFactory, syncService, cfg.RegistrationEnabled, !cfg.AllowInsecure)
	dawarichSync := newDawarichSyncService(db, cfg.DawarichURL)
	dawarichHandlers := newDawarichHandlers(db, cfg.DawarichURL, dawarichSync, cfg.defaultTimezoneLocation)

	authMux := http.NewServeMux()
	authMux.HandleFunc("POST /auth/register", authHandlers.handleRegister)
	authMux.HandleFunc("POST /auth/login", authHandlers.handleLogin)
	authMux.HandleFunc("POST /auth/logout", authHandlers.handleLogout)
	authMux.HandleFunc("GET /auth/status", authHandlers.handleAuthStatus)
	authMux.Handle("GET /auth/me", sessionMiddleware(db, http.HandlerFunc(authHandlers.handleMe)))
	authMux.Handle("PUT /auth/settings", sessionMiddleware(db, http.HandlerFunc(authHandlers.handleUpdateSettings)))

	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("GET /albums", handlers.handleGetAlbums)
	protectedMux.HandleFunc("GET /tags", handlers.handleGetTags)
	protectedMux.HandleFunc("GET /folders", handlers.handleGetFolders)
	protectedMux.HandleFunc("GET /folders/assets", handlers.handleGetFolderAssets)
	protectedMux.HandleFunc("GET /assets", handlers.handleGetAssets)
	protectedMux.HandleFunc("GET /assets/day-counts", handlers.handleGetAssetDayCounts)
	protectedMux.HandleFunc("GET /assets/missing-location-count", handlers.handleGetMissingLocationCount)
	protectedMux.HandleFunc("GET /assets/{assetID}/folder", handlers.handleGetAssetFolder)
	protectedMux.HandleFunc("GET /map-markers", handlers.handleGetMapMarkers)
	protectedMux.HandleFunc("GET /assets/{assetID}/thumbnail", handlers.handleGetThumbnail)
	protectedMux.HandleFunc("GET /assets/{assetID}/preview", handlers.handleGetPreview)
	protectedMux.HandleFunc("PUT /assets/{assetID}/location", handlers.handleUpdateLocation)
	protectedMux.HandleFunc("PUT /assets/{assetID}/hidden", handlers.handleUpdateHidden)
	protectedMux.HandleFunc("PUT /assets/bulk-hidden", handlers.handleBulkUpdateHidden)
	protectedMux.HandleFunc("GET /assets/{assetID}/suggestions", handlers.handleGetSuggestions)
	protectedMux.HandleFunc("GET /frequent-locations", handlers.handleGetFrequentLocations)
	protectedMux.HandleFunc("GET /assets/{assetID}/page-info", handlers.handleGetAssetPageInfo)
	protectedMux.HandleFunc("POST /sync", handlers.handleTriggerSync)
	protectedMux.HandleFunc("POST /sync/full", handlers.handleTriggerFullSync)
	protectedMux.HandleFunc("GET /sync/status", handlers.handleSyncStatus)
	protectedMux.HandleFunc("GET /libraries", libraryHandlers.handleGetLibraries)
	protectedMux.HandleFunc("PUT /libraries/{libraryID}", libraryHandlers.handleUpdateLibrary)
	protectedMux.HandleFunc("POST /libraries/refresh", libraryHandlers.handleRefreshLibraries)
	protectedMux.HandleFunc("GET /favorite-places", handlers.handleGetFavoritePlaces)
	protectedMux.HandleFunc("POST /favorite-places", handlers.handleAddFavoritePlace)
	protectedMux.HandleFunc("DELETE /favorite-places", handlers.handleRemoveFavoritePlace)
	protectedMux.HandleFunc("POST /gpx/preview", handlers.handleGPXPreview)
	protectedMux.HandleFunc("PUT /dawarich/settings", dawarichHandlers.handleDawarichSettings)
	protectedMux.HandleFunc("DELETE /dawarich/settings", dawarichHandlers.handleDeleteDawarichSettings)
	protectedMux.HandleFunc("GET /dawarich/tracks", dawarichHandlers.handleDawarichTracks)
	protectedMux.HandleFunc("POST /dawarich/preview", dawarichHandlers.handleDawarichPreview)
	protectedMux.HandleFunc("GET /dawarich/sync/status", dawarichHandlers.handleDawarichSyncStatus)
	protectedMux.HandleFunc("POST /dawarich/sync", dawarichHandlers.handleDawarichTriggerSync)
	protectedMux.HandleFunc("GET /geocode/search", handlers.handleGeocodeSearch)

	capabilityTransport := newAIProviderCapabilityTransport(providerhttp.Options{})
	providerDispatcher := newAIProviderDispatcher(db, cfg.AIEnabled, cfg.AIProviderEgressPolicy, capabilityTransport)

	mainMux := http.NewServeMux()
	mainMux.HandleFunc("GET /health", handlers.handleHealth)
	mainMux.Handle("/auth/", authMux)
	mainMux.Handle("/ai/", newAIProviderHandler(db, cfg, providerDispatcher))
	selectionHandler := newAISelectionHandler(db, cfg)
	mainMux.Handle("/ai/selection-preview", selectionHandler)
	mainMux.Handle("/ai/selections/", selectionHandler)
	productionRuntime, err := newAIProductionRuntime(db, cfg, selectionHandler.store, providerDispatcher)
	if err != nil {
		log.Fatalf("[AI jobs] Production job runtime initialization failed")
	}
	jobHandler := newAIJobHandler(productionRuntime.jobs, cfg.AIPublicOrigin)
	mainMux.Handle("/ai/jobs", jobHandler)
	mainMux.Handle("/ai/jobs/", jobHandler)
	resultImages, err := newAIImagePreparer(db, selectionHandler.store, cfg.ImmichURL)
	if err != nil {
		log.Printf("[AI results] Current image reads unavailable")
	}
	resultStore := &aiResultStore{jobs: productionRuntime.jobs.store, origin: cfg.AIPublicOrigin}
	writeRuntime := newAIWriteRuntime(resultStore, resultImages, syncService, cfg)
	resultHandler := newAIResultHandler(resultStore, resultImages)
	mainMux.Handle("/ai/write-operations", resultHandler)
	mainMux.Handle("/ai/write-operations/", resultHandler)
	mainMux.Handle("/ai/drafts/", resultHandler)
	mainMux.Handle("/ai/write-previews", resultHandler)
	mainMux.Handle("/ai/write-previews/", resultHandler)
	mainMux.Handle("/ai/results", resultHandler)
	mainMux.Handle("/ai/results/", resultHandler)
	mainMux.Handle("GET /ai/jobs/{jobID}/items/{itemID}/result", resultHandler)
	mainMux.Handle("GET /ai/jobs/{jobID}/items/{itemID}/thumbnail", resultHandler)

	mainMux.Handle("/", sessionMiddleware(db, protectedMux))

	handler := requestHardeningMiddleware(mainMux)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	productionDone := make(chan struct{})
	go func() {
		defer close(productionDone)
		if err := productionRuntime.run(ctx, func() { log.Printf("[AI jobs] Worker operation failed; bounded recovery will retry") }); err != nil {
			log.Printf("[AI jobs] Worker runtime stopped with an invalid configuration")
		}
	}()

	selectionCleanupTicker := time.NewTicker(time.Minute)
	defer selectionCleanupTicker.Stop()
	go selectionHandler.store.runCleanup(ctx, selectionCleanupTicker.C, func() { log.Printf("[AI selection] Expired snapshot cleanup failed; next pass will retry") })

	writeDone := make(chan struct{})
	go func() { defer close(writeDone); writeRuntime.run(ctx) }()
	syncService.shutdownCtx = ctx
	dawarichSync.shutdownCtx = ctx

	users, err := db.getUsersWithAPIKeys(ctx)
	if err != nil {
		log.Printf("[Server] Failed to load users for startup sync: %v", err)
	} else {
		syncService.runStartupSyncs(ctx, users)
		dawarichSync.runStartupSyncs(ctx, users)
	}

	syncService.startPeriodicSync(ctx, cfg.SyncIntervalMS)
	dawarichSync.startPeriodicSync(ctx, cfg.DawarichSyncIntervalMS)

	addr := fmt.Sprintf(":%d", cfg.Port)

	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      150 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		if cfg.TrustProxyTLS {
			log.Printf("[Server] Listening on %s (TLS terminated by reverse proxy)", addr)
		} else {
			log.Printf("[Server] Listening on %s (no TLS — ensure network is secure)", addr)
		}
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[Server] Error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("[Server] Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[Server] HTTP shutdown error: %v", err)
	}

	<-productionDone
	select {
	case <-writeDone:
	case <-time.After(7 * time.Second):
		log.Printf("[AI writes] Shutdown deadline reached; durable reservations require read-only recovery")
	}
	syncService.wg.Wait()
	dawarichSync.wg.Wait()
	log.Println("[Server] All sync goroutines completed")
}

const maxRequestBodyBytes = 10_000_000
const maxQueryLength = 2048

func requestHardeningMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.RawQuery) > maxQueryLength {
			writeError(w, http.StatusRequestURITooLong, "query string too long")
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
		}
		next.ServeHTTP(w, r)
	})
}

func sessionMiddleware(db *Database, next http.Handler) http.Handler {
	return sessionMiddlewareWithErrors(db, next, writeError)
}

func sessionMiddlewareWithErrors(db *Database, next http.Handler, respond func(http.ResponseWriter, int, string)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || cookie.Value == "" {
			respond(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		hash := sha256.Sum256([]byte(cookie.Value))
		tokenHash := hex.EncodeToString(hash[:])

		user, err := db.getSessionUser(r.Context(), tokenHash)
		if err != nil {
			log.Printf("[Auth] Session DB error: %v", err)
			respond(w, http.StatusInternalServerError, "internal error")
			return
		}
		if user == nil {
			respond(w, http.StatusUnauthorized, "session expired")
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
