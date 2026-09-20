package main

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/providers"
)

type aiProviderHandlers struct {
	db                *Database
	enabled           bool
	origin            string
	policy            providers.EgressPolicy
	dispatcher        *providers.Dispatcher
	limiter           *capabilities.Limiter
	executionPolicies jobs.ExecutionPolicies
}

type aiProviderRequest struct {
	providers.Input
	ExpectedRevision int `json:"expectedRevision"`
}

func newAIProviderHandler(db *Database, cfg *Config, dispatcher *providers.Dispatcher) http.Handler {
	h := &aiProviderHandlers{
		db:                db,
		enabled:           cfg.AIEnabled,
		origin:            cfg.AIPublicOrigin,
		policy:            cfg.AIProviderEgressPolicy,
		dispatcher:        dispatcher,
		limiter:           capabilities.NewLimiter(2, 1),
		executionPolicies: cfg.AIExecutionPolicies,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ai/providers", h.list)
	mux.HandleFunc("POST /ai/providers", h.save)
	mux.HandleFunc("PUT /ai/providers/{id}", h.save)
	mux.HandleFunc("POST /ai/providers/{id}/test", h.testProvider)
	return sessionMiddlewareWithErrors(db, mux, func(w http.ResponseWriter, status int, message string) {
		code := "UNAUTHENTICATED"
		if status == http.StatusInternalServerError {
			code = "STORAGE_ERROR"
		}
		writeAIProviderError(w, status, code, message)
	})
}

func (h *aiProviderHandlers) list(w http.ResponseWriter, r *http.Request) {
	items := make([]aiProviderProfile, 0)
	if h.enabled {
		var err error
		items, err = h.db.listAIProviders(r.Context(), getUserFromContext(r).ID, capabilities.ApplicabilityContext{
			AIEnabled:                true,
			CurrentPolicyFingerprint: policyFingerprint(h.policy),
		})
		if err != nil {
			writeAIProviderFailure(w, err)
			return
		}
		for i := range items {
			items[i].ExecutionReadiness, err = h.executionReadiness(r.Context(), getUserFromContext(r).ID, items[i])
			if err != nil {
				writeAIProviderFailure(w, err)
				return
			}
		}
	}
	writeJSON(w, http.StatusOK, struct {
		Enabled bool                `json:"enabled"`
		Items   []aiProviderProfile `json:"items"`
	}{h.enabled, items})
}

func (h *aiProviderHandlers) save(w http.ResponseWriter, r *http.Request) {
	if !h.enabled {
		writeAIProviderError(w, http.StatusServiceUnavailable, "AI_DISABLED", "AI is disabled by the administrator")
		return
	}
	if len(r.Header.Values("Origin")) != 1 || r.Header.Get("Origin") == "" || r.Header.Get("Origin") != h.origin {
		writeAIProviderError(w, http.StatusForbidden, "ORIGIN_REJECTED", "request origin is not allowed")
		return
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeAIProviderError(w, http.StatusUnsupportedMediaType, "INVALID_REQUEST", "Content-Type must be application/json")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request *aiProviderRequest
	if err := decoder.Decode(&request); err != nil || request == nil {
		writeAIProviderError(w, http.StatusBadRequest, "INVALID_REQUEST", "provide one valid provider JSON object, at most 16 KiB, without unsupported fields")
		return
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		writeAIProviderError(w, http.StatusBadRequest, "INVALID_REQUEST", "provide one JSON object within the 16 KiB limit")
		return
	}
	var profile aiProviderProfile
	status := http.StatusOK
	if r.Method == http.MethodPost {
		if request.ExpectedRevision != 0 {
			writeAIProviderError(w, http.StatusBadRequest, "INVALID_REQUEST", "new profiles must not set expectedRevision")
			return
		}
		profile, err = h.db.createAIProvider(r.Context(), getUserFromContext(r).ID, uuid.NewString(), request.Input)
		status = http.StatusCreated
	} else {
		profile, err = h.db.updateAIProvider(r.Context(), getUserFromContext(r).ID, r.PathValue("id"), request.ExpectedRevision, request.Input)
	}
	if err != nil {
		writeAIProviderFailure(w, err)
		return
	}
	profile.ExecutionReadiness, err = h.executionReadiness(r.Context(), getUserFromContext(r).ID, profile)
	if err != nil {
		writeAIProviderFailure(w, err)
		return
	}
	writeJSON(w, status, profile)
}

func writeAIProviderFailure(w http.ResponseWriter, err error) {
	var invalid *aiProviderInputError
	switch {
	case errors.As(err, &invalid):
		writeAIProviderError(w, http.StatusBadRequest, "INVALID_PROVIDER", invalid.Error())
	case errors.Is(err, errAIProviderConflict):
		writeAIProviderError(w, http.StatusConflict, "PROVIDER_CONFLICT", errAIProviderConflict.Error())
	case errors.Is(err, errAIProviderNotFound):
		writeAIProviderError(w, http.StatusNotFound, "PROVIDER_NOT_FOUND", errAIProviderNotFound.Error())
	default:
		writeAIProviderError(w, http.StatusInternalServerError, "STORAGE_ERROR", "provider settings could not be saved or loaded")
	}
}

func writeAIProviderError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		Retryable bool   `json:"retryable"`
		RequestID string `json:"requestID"`
	}{code, message, code == "STORAGE_ERROR", uuid.NewString()})
}
