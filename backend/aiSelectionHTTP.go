package main

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"immich-places-backend/internal/ai/selection"
	"io"
	"mime"
	"net/http"
	"time"
)

type aiSelectionHTTP struct {
	http.Handler
	store *aiSelectionStore
}

func newAISelectionHandler(db *Database, cfg *Config) *aiSelectionHTTP {
	store := newAISelectionStore(db)
	store.enabled = cfg.AIEnabled
	if cfg.AISelectionMaxAssets != 0 {
		store.maxAssets = cfg.AISelectionMaxAssets
	}
	if cfg.AISelectionTTLSeconds != 0 {
		store.ttl = time.Duration(cfg.AISelectionTTLSeconds) * time.Second
	}
	epoch := cfg.AIInstanceEpoch
	if epoch == "" {
		epoch = "1"
	}
	initializationError := store.bind(context.Background(), cfg.ImmichURL, epoch)
	if initializationError == nil {
		initializationError = store.cleanup(context.Background())
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /ai/selection-preview", func(w http.ResponseWriter, r *http.Request) {

		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			writeAIProviderError(w, 415, "INVALID_SELECTION", "Content-Type must be application/json")
			return
		}
		data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, selection.MaxBytes))
		if err != nil {
			var oversized *http.MaxBytesError
			if errors.As(err, &oversized) {
				writeAISelectionFailure(w, selection.ErrLimit)
			} else {
				writeAISelectionFailure(w, selection.ErrInvalid)
			}
			return
		}
		input, err := selection.Decode(data)
		if err != nil {
			writeAISelectionFailure(w, err)
			return
		}
		manifest, err := store.preview(r.Context(), input, getUserFromContext(r).ID)
		if err != nil {
			writeAISelectionFailure(w, err)
			return
		}
		writeJSON(w, 200, manifest)
	})
	mux.HandleFunc("GET /ai/selections/{id}", func(w http.ResponseWriter, r *http.Request) {
		manifest, err := store.load(r.Context(), getUserFromContext(r).ID, r.PathValue("id"))
		if err != nil {
			writeAISelectionFailure(w, err)
			return
		}
		writeJSON(w, 200, manifest)
	})
	guarded := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !cfg.AIEnabled {
			writeAIProviderError(w, 503, "AI_DISABLED", "AI is disabled")
			return
		}
		if r.Method == http.MethodPost && (len(r.Header.Values("Origin")) != 1 || r.Header.Get("Origin") == "" || r.Header.Get("Origin") != cfg.AIPublicOrigin) {
			writeAIProviderError(w, 403, "ORIGIN_REJECTED", "request origin is not allowed")
			return
		}
		if initializationError != nil {
			writeAIProviderError(w, 503, "SELECTION_UNAVAILABLE", "selection initialization failed")
			return
		}
		mux.ServeHTTP(w, r)
	})
	protected := sessionMiddlewareWithErrors(db, guarded, func(w http.ResponseWriter, status int, message string) {
		code := "UNAUTHENTICATED"
		if status == http.StatusInternalServerError {
			code = "STORAGE_ERROR"
		}
		writeAIProviderError(w, status, code, message)
	})
	return &aiSelectionHTTP{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		protected.ServeHTTP(w, r)
	}), store: store}
}

func writeAISelectionFailure(w http.ResponseWriter, err error) {
	var limit *selection.MatchingLimitError
	if errors.As(err, &limit) {
		writeJSON(w, 413, struct {
			selection.MatchingCounts
			SnapshotID *string `json:"snapshotID"`
			Code       string  `json:"code"`
			Message    string  `json:"message"`
			Retryable  bool    `json:"retryable"`
			RequestID  string  `json:"requestID"`
		}{MatchingCounts: limit.MatchingCounts, Code: "SELECTION_LIMIT_EXCEEDED", Message: "matching selection exceeds eligible asset limit", RequestID: uuid.NewString()})
		return
	}
	status, code, message := 500, "STORAGE_ERROR", "selection storage is unavailable"
	var sqliteError interface{ Code() int }
	busy := errors.As(err, &sqliteError) && (sqliteError.Code()&255 == 5 || sqliteError.Code()&255 == 6)
	switch {
	case busy || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled):
		status, code, message = 503, "SELECTION_BUSY", "selection could not complete within its deadline"
	case errors.Is(err, errAISelectionDisabled):
		status, code, message = 503, "AI_DISABLED", "AI is disabled"
	case errors.Is(err, selection.ErrInvalid):
		status, code, message = 400, "INVALID_SELECTION", "provide a supported selection and scope"
	case errors.Is(err, selection.ErrLimit):
		status, code, message = 413, "SELECTION_LIMIT", "selection exceeds a resource limit"
	case errors.Is(err, errAISelectionOwnerQuota):
		status, code, message = 429, "SELECTION_QUOTA", "too many active selections"
	case errors.Is(err, errAISelectionCapacity):
		status, code, message = 503, "SELECTION_CAPACITY", "selection storage capacity is temporarily exhausted"
	case errors.Is(err, errAISelectionStale):
		status, code, message = 409, "SELECTION_STALE", "selection must be previewed again"
	case errors.Is(err, sql.ErrNoRows):
		status, code, message = 404, "SELECTION_NOT_FOUND", "selection is unavailable"
	}
	writeJSON(w, status, struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		Retryable bool   `json:"retryable"`
		RequestID string `json:"requestID"`
	}{code, message, status == 500 || status == 503, uuid.NewString()})
}
