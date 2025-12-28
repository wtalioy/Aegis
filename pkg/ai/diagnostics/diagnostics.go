package diagnostics

import (
	"aegis/pkg/ai/prompt"
	"aegis/pkg/ai/snapshot"
)

// BuildPrompt renders the diagnosis prompt from system state.
func BuildPrompt(state snapshot.SystemState) (string, error) {
	return prompt.GeneratePrompt(state)
}

func SnapshotSummary(state snapshot.SystemState) string {
	return prompt.FormatSnapshotSummary(state)
}

