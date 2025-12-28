package chat

import (
	"aegis/pkg/ai/prompt"
	"aegis/pkg/ai/types"
	"aegis/pkg/ai/snapshot"
	"aegis/pkg/proc"
)

func BuildMessages(history []types.Message, state snapshot.SystemState, userMessage string, processTree *proc.ProcessTree, processKeyToChain, processNameToChain map[string]string) []types.Message {
	messages := make([]types.Message, 0, len(history)+3)

	messages = append(messages, types.Message{
		Role:    "system",
		Content: prompt.ChatSystemPrompt,
	})

	// Use intelligent context filtering based on user query
	contextMsg := prompt.FormatContextForChatWithFilter(state, userMessage, processTree, processKeyToChain, processNameToChain)
	messages = append(messages, types.Message{
		Role:    "system",
		Content: contextMsg,
	})

	for _, msg := range history {
		messages = append(messages, types.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	messages = append(messages, types.Message{
		Role:    "user",
		Content: userMessage,
	})

	return messages
}

