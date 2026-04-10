package diagnostics

import (
	"aegis/internal/platform/ai/prompt"
	"aegis/internal/platform/ai/snapshot"
)

// BuildPrompt renders the diagnosis prompt from system state.
func BuildPrompt(state snapshot.SystemState) (string, error) {
	return prompt.GeneratePrompt(state)
}

func SnapshotSummary(state snapshot.SystemState) string {
	return prompt.FormatSnapshotSummary(state)
}
