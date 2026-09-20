package analysis

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"sync/atomic"
	"time"

	"immich-places-backend/internal/ai/contextual"
	"immich-places-backend/internal/ai/images"
	"immich-places-backend/internal/ai/results"
)

type Protocol struct {
	Encode func(model, instruction, imageDataURL, format string) ([]byte, error)
	Parse  func([]byte) (Response, error)
}
type DispatchRequest struct {
	OwnerID, ProfileID string
	Revision           int
	Body               []byte
	Authorize          func(context.Context) error
}
type DispatchReply struct{ Body []byte }
type DispatchFunc func(context.Context, DispatchRequest) (DispatchReply, error)
type Guard struct{ Authorize, Reserve func(context.Context) error }
type Request struct {
	Owner, Installation, Asset, ProfileID, Model, Format string
	Revision                                             int
	AllowJSON                                            bool
	Languages                                            []string
	PrimaryLanguage                                      string
	Image                                                *images.Prepared
	Context                                              *contextual.Bundle
	Guard                                                Guard
}
type Runner struct {
	protocol  Protocol
	dispatch  DispatchFunc
	validator *results.Validator
}

func New(protocol Protocol, dispatch DispatchFunc) (*Runner, error) {
	if protocol.Encode == nil || protocol.Parse == nil || dispatch == nil {
		return nil, ErrInvalidRequest
	}
	validator, err := results.New()
	if err != nil {
		return nil, err
	}
	return &Runner{protocol: protocol, dispatch: dispatch, validator: validator}, nil
}

func (r *Runner) RunVisual(ctx context.Context, req Request) (result *Result, err error) {
	if req.Context != nil {
		return nil, ErrInvalidRequest
	}
	return r.run(ctx, req, results.Visual)
}

func (r *Runner) run(ctx context.Context, req Request, mode results.Mode) (result *Result, err error) {
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	defer func() {
		if ctx.Err() != nil {
			result, err = nil, ctx.Err()
		}
	}()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	tags, primary, err := results.NormalizeLanguages(req.Languages, req.PrimaryLanguage)
	if err != nil {
		return nil, err
	}
	info, valid := req.Image.Info()
	if !valid || info.Binding.Owner != req.Owner || info.Binding.Installation != req.Installation || info.Binding.Asset != req.Asset || req.ProfileID == "" || len(req.ProfileID) > 128 || req.Revision < 1 || req.Guard.Authorize == nil || req.Guard.Reserve == nil || (req.Format != "strict" && req.Format != "json") || (req.Format == "json" && !req.AllowJSON) {
		return nil, ErrInvalidRequest
	}
	authorize := func(ctx context.Context) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		guardErr := req.Guard.Authorize(ctx)
		if err := ctx.Err(); err != nil {
			return err
		}
		if guardErr != nil {
			if errors.Is(guardErr, ErrUpstream) || errors.Is(guardErr, ErrLimit) {
				return dispatchFailure(guardErr)
			}
			return ErrDenied
		}
		return nil
	}
	if err := authorize(ctx); err != nil {
		return nil, err
	}
	data, err := req.Image.Bytes()
	if err != nil {
		return nil, err
	}
	defer clear(data)
	languageSettings, _ := json.Marshal(struct {
		Languages []string `json:"languages"`
		Primary   string   `json:"primary_language"`
	}{tags, primary})
	prompt, validation, err := attemptContext(req, mode)
	if err != nil {
		return nil, err
	}
	validation.Languages, validation.PrimaryLanguage = tags, primary
	instruction := prompt + "\nLanguage settings: " + string(languageSettings) + "\nCanonical JSON schema: " + string(results.CanonicalSchema())
	body, err := r.protocol.Encode(req.Model, instruction, "data:image/jpeg;base64,"+base64.StdEncoding.EncodeToString(data), req.Format)
	if err != nil {
		return nil, ErrInvalidRequest
	}
	defer clear(body)
	if len(body) > 15<<20 {
		return nil, ErrLimit
	}
	var reserved, admitted atomic.Bool
	reply, err := r.dispatch(ctx, DispatchRequest{OwnerID: req.Owner, ProfileID: req.ProfileID, Revision: req.Revision, Body: body, Authorize: func(ctx context.Context) error {
		if !reserved.CompareAndSwap(false, true) {
			return ErrBudget
		}
		if err := authorize(ctx); err != nil {
			return err
		}
		reserveErr := req.Guard.Reserve(ctx)
		if err := ctx.Err(); err != nil {
			return err
		}
		if reserveErr != nil {
			return ErrBudget
		}
		admitted.Store(true)
		return nil
	}})
	if err != nil {
		return nil, dispatchFailure(err)
	}
	defer clear(reply.Body)
	if !admitted.Load() {
		return nil, ErrBudget
	}
	if len(reply.Body) > 1<<20 {
		return nil, ErrLimit
	}
	parsed, err := r.protocol.Parse(reply.Body)
	if err != nil {
		return nil, ErrInvalidResponse
	}
	defer clear(parsed.Content)
	proposal, err := r.validator.Validate(parsed.Content, validation)
	if err != nil {
		return nil, err
	}
	if err := authorize(ctx); err != nil {
		return nil, err
	}
	metadata := Metadata{Mode: mode, Image: info, ProfileID: req.ProfileID, Revision: req.Revision, Model: req.Model, Format: req.Format, PromptVersion: PromptVersion, SchemaVersion: "1.0", Languages: tags, PrimaryLanguage: primary, Usage: parsed.Usage}
	if mode == results.ContextAssisted {
		contextInfo := req.Context.Info()
		metadata.Context = &contextInfo
		metadata.PromptVersion = ContextPromptVersion
	}
	return newResult(proposal, metadata), nil
}
