package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"time"

	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/ai/providers"
)

type capabilityTestRequest struct {
	ExpectedRevision int `json:"expectedRevision"`
}

func (h *aiProviderHandlers) testProvider(w http.ResponseWriter, r *http.Request) {
	if !h.enabled {
		writeAIProviderError(w, http.StatusServiceUnavailable, "AI_DISABLED", "AI is disabled by the administrator")
		return
	}
	if h.dispatcher == nil {
		writeAIProviderError(w, http.StatusServiceUnavailable, "AI_DISABLED", "AI provider dispatch is unavailable")
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
	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request *capabilityTestRequest
	if err := decoder.Decode(&request); err != nil || request == nil || request.ExpectedRevision < 1 {
		writeAIProviderError(w, http.StatusBadRequest, "INVALID_REQUEST", "provide one valid capability test JSON object, at most 1 KiB, with a positive expectedRevision")
		return
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		writeAIProviderError(w, http.StatusBadRequest, "INVALID_REQUEST", "provide one JSON object within the 1 KiB limit")
		return
	}

	user := getUserFromContext(r)
	if err := h.limiter.Acquire(user.ID); err != nil {
		writeAIProviderError(w, http.StatusConflict, "PROVIDER_BUSY", "another capability test is already running")
		return
	}
	defer h.limiter.Release(user.ID)

	destination, err := h.db.loadAIProviderCapabilityDestination(r.Context(), user.ID, r.PathValue("id"))
	if err != nil {
		writeAIProviderCapabilityAdmissionFailure(w, err)
		return
	}
	if _, err := providers.MatchEgressDestination(h.policy, destination); err != nil {
		writeAIProviderError(w, http.StatusForbidden, "DESTINATION_DENIED", "provider destination is not approved")
		return
	}

	now := time.Now().UTC()
	deadline := capabilities.SharedDeadline(now)
	fingerprint := policyFingerprint(h.policy)
	admission, err := h.db.admitAIProviderCapability(r.Context(), user.ID, r.PathValue("id"), request.ExpectedRevision, fingerprint, now, deadline)
	if err != nil {
		writeAIProviderCapabilityAdmissionFailure(w, err)
		return
	}

	workCtx, cancel := context.WithDeadline(r.Context(), deadline)
	defer cancel()

	var sessionTokenHash string
	if cookie, err := r.Cookie(sessionCookieName); err == nil && cookie.Value != "" {
		hash := sha256.Sum256([]byte(cookie.Value))
		sessionTokenHash = hex.EncodeToString(hash[:])
	}
	ownerID := user.ID
	runner := &capabilities.Runner{
		Protocol: capabilityChatProtocol{},
		Admit: func(ctx context.Context, admitOwnerID, profileID string, revision int) error {
			if !h.enabled {
				return capabilities.ErrDisabled
			}
			sessionUser, err := h.db.getSessionUser(ctx, sessionTokenHash)
			if err != nil {
				return err
			}
			if sessionUser == nil || sessionUser.ID != ownerID {
				return capabilities.ErrUnavailable
			}
			return h.db.admitAIProviderCapabilityProbe(ctx, admitOwnerID, profileID, revision)
		},
		Dispatch: func(ctx context.Context, ownerID, profileID string, revision int, body []byte) ([]byte, error) {
			result, err := h.dispatcher.Dispatch(ctx, providers.DispatchRequest{
				OwnerID:   ownerID,
				ProfileID: profileID,
				Revision:  revision,
				Body:      body,
			})
			if err != nil {
				return nil, classifyCapabilityProviderFailure(err)
			}
			return result.Body, nil
		},
	}
	// An accepted test reports probe failures through its persisted report, not as an HTTP error.
	report, _ := runner.Run(workCtx, capabilities.ProbeRequest{
		OwnerID:           user.ID,
		ProfileID:         admission.ProfileID,
		Revision:          admission.Revision,
		Model:             admission.Model,
		PolicyFingerprint: admission.PolicyFingerprint,
		AttemptID:         admission.AttemptID,
		StartedAt:         admission.StartedAt,
		DeadlineAt:        admission.DeadlineAt,
		Applicability: capabilities.ApplicabilityContext{
			AIEnabled:                true,
			ProfileEnabled:           true,
			ActiveRevision:           admission.Revision,
			CurrentProtocolVersion:   capabilities.ProtocolVersion,
			CurrentPolicyFingerprint: admission.PolicyFingerprint,
		},
	})
	completeCtx, completeCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer completeCancel()
	if err := h.db.completeAIProviderCapability(completeCtx, user.ID, report); err != nil {
		writeAIProviderError(w, http.StatusInternalServerError, "STORAGE_ERROR", "capability test result could not be saved")
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func writeAIProviderCapabilityAdmissionFailure(w http.ResponseWriter, err error) {
	if writeAIProviderCapabilityKnownFailure(w, err) {
		return
	}
	writeAIProviderError(w, http.StatusInternalServerError, "STORAGE_ERROR", "capability test could not be saved")
}

func writeAIProviderCapabilityKnownFailure(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, errAIProviderNotFound), errors.Is(err, capabilities.ErrUnavailable):
		writeAIProviderError(w, http.StatusNotFound, "PROVIDER_NOT_FOUND", errAIProviderNotFound.Error())
	case errors.Is(err, errAIProviderConflict):
		writeAIProviderError(w, http.StatusConflict, "PROVIDER_CONFLICT", errAIProviderConflict.Error())
	case errors.Is(err, capabilities.ErrDisabled), errors.Is(err, providers.ErrDisabled):
		writeAIProviderError(w, http.StatusConflict, "PROVIDER_DISABLED", "provider profile is disabled")
	case errors.Is(err, capabilities.ErrBusy):
		writeAIProviderError(w, http.StatusConflict, "PROVIDER_BUSY", "another capability test is already running")
	default:
		return false
	}
	return true
}
