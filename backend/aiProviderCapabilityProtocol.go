package main

import (
	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

type capabilityChatProtocol struct{}

func (capabilityChatProtocol) EncodeImageProbe(model, instruction, imageDataURL string) ([]byte, error) {
	return providerhttp.EncodeImageProbe(model, instruction, imageDataURL)
}

func (capabilityChatProtocol) EncodeJSONProbe(model, instruction, imageDataURL string) ([]byte, error) {
	return providerhttp.EncodeJSONProbe(model, instruction, imageDataURL)
}

func (capabilityChatProtocol) EncodeStrictProbe(model, instruction, imageDataURL string) ([]byte, error) {
	return providerhttp.EncodeStrictProbe(model, instruction, imageDataURL)
}

func (capabilityChatProtocol) ParseAssistantText(body []byte) (string, string, error) {
	return providerhttp.ParseAssistantText(body)
}

func (capabilityChatProtocol) ParseUsage(body []byte) *capabilities.Usage {
	return providerhttp.ParseUsage(body)
}
