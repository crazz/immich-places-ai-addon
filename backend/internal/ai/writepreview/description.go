package writepreview

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/text/language"
)

type TextObservation struct {
	Presence string `json:"presence"`
	Value    string `json:"value"`
}

type AppendLineage struct {
	ID    string `json:"id"`
	Block string `json:"block"`
	Hash  string `json:"hash"`
}

type DescriptionInput struct {
	Text, Language, Policy, NewLineageID string
	Owned                                *AppendLineage
}

type DescriptionPlan struct {
	Before   TextObservation `json:"before"`
	Intended string          `json:"intended"`
	Language string          `json:"language"`
	Policy   string          `json:"policy"`
	Lineage  *AppendLineage  `json:"lineage,omitempty"`
}

func PlanDescription(before TextObservation, input DescriptionInput) (*DescriptionPlan, error) {
	tag, err := language.Parse(input.Language)
	if !ValidTextObservation(before) || len(input.Text) > 16<<10 || !utf8.ValidString(input.Text) || strings.TrimSpace(input.Text) == "" || err != nil || tag.String() != input.Language || (input.Policy != "managed_append" && input.Policy != "replace") {
		return nil, Failure{Code: "DESCRIPTION_INVALID"}
	}
	if input.Policy == "replace" {
		return &DescriptionPlan{Before: before, Intended: input.Text, Language: input.Language, Policy: input.Policy}, nil
	}
	if hasManagedMarker(input.Text) {
		return nil, Failure{Code: "DESCRIPTION_CONFLICT"}
	}
	id := input.NewLineageID
	if input.Owned != nil {
		id = input.Owned.ID
		if !validOwnedBlock(before.Value, *input.Owned) {
			return nil, Failure{Code: "DESCRIPTION_CONFLICT"}
		}
	} else if hasManagedMarker(before.Value) {
		return nil, Failure{Code: "DESCRIPTION_CONFLICT"}
	}
	if _, err = uuid.Parse(id); err != nil {
		return nil, Failure{Code: "DESCRIPTION_INVALID"}
	}
	block := "[[Immich Places AI v1:" + id + "]]\nLanguage: " + input.Language + "\n" + input.Text + "\n[[/Immich Places AI v1:" + id + "]]"
	sum := sha256.Sum256([]byte(block))
	separator := ""
	if before.Value != "" {
		separator = "\n\n"
	}
	intended := before.Value + separator + block
	if input.Owned != nil {
		intended = strings.Replace(before.Value, input.Owned.Block, block, 1)
	}
	if len(intended) > 64<<10 {
		return nil, Failure{Code: "DESCRIPTION_TOO_LARGE"}
	}
	return &DescriptionPlan{Before: before, Intended: intended, Language: input.Language, Policy: input.Policy, Lineage: &AppendLineage{ID: id, Block: block, Hash: hex.EncodeToString(sum[:])}}, nil
}

func ValidTextObservation(value TextObservation) bool {
	if len(value.Value) > 64<<10 || !utf8.ValidString(value.Value) {
		return false
	}
	return value.Presence == "value" || ((value.Presence == "absent" || value.Presence == "null") && value.Value == "")
}

func hasManagedMarker(text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(lower, "[[immich places ai") || strings.Contains(lower, "[[/immich places ai")
}

func validOwnedBlock(before string, owned AppendLineage) bool {
	sum := sha256.Sum256([]byte(owned.Block))
	if owned.Hash != hex.EncodeToString(sum[:]) || strings.Count(before, owned.Block) != 1 || !strings.HasPrefix(owned.Block, "[[Immich Places AI v1:"+owned.ID+"]]\n") || !strings.HasSuffix(owned.Block, "\n[[/Immich Places AI v1:"+owned.ID+"]]") {
		return false
	}
	lower := strings.ToLower(owned.Block)
	return strings.Count(lower, "[[immich places ai") == 1 && strings.Count(lower, "[[/immich places ai") == 1 && !hasManagedMarker(strings.Replace(before, owned.Block, "", 1))
}

func ValidDescriptionPlan(plan DescriptionPlan) bool {
	tag, err := language.Parse(plan.Language)
	if err != nil || tag.String() != plan.Language || !ValidTextObservation(plan.Before) || len(plan.Intended) > 64<<10 || !utf8.ValidString(plan.Intended) || strings.TrimSpace(plan.Intended) == "" {
		return false
	}
	switch plan.Policy {
	case "replace":
		return plan.Lineage == nil && len(plan.Intended) <= 16<<10
	case "managed_append":
		if plan.Lineage == nil || !validOwnedBlock(plan.Intended, *plan.Lineage) {
			return false
		}
		if _, err = uuid.Parse(plan.Lineage.ID); err != nil {
			return false
		}
		prefix := "[[Immich Places AI v1:" + plan.Lineage.ID + "]]\nLanguage: " + plan.Language + "\n"
		suffix := "\n[[/Immich Places AI v1:" + plan.Lineage.ID + "]]"
		if !strings.HasPrefix(plan.Lineage.Block, prefix) {
			return false
		}
		text := strings.TrimSuffix(strings.TrimPrefix(plan.Lineage.Block, prefix), suffix)
		return len(text) <= 16<<10 && strings.TrimSpace(text) != ""
	}
	return false
}
