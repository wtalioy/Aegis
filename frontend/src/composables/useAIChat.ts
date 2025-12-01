import { ref, computed } from 'vue'
import {
    sendChatMessageStream,
    getChatHistory,
    clearChatHistory,
    type ChatMessage,
    type ChatStreamToken
} from '../lib/api'

// Generate a unique session ID
function generateSessionId(): string {
    return `chat-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
}

// Singleton state (persists across component instances)
const sessionId = ref<string>(generateSessionId())
const messages = ref<ChatMessage[]>([])
const isLoading = ref(false)
const error = ref<string | null>(null)
const lastContextSummary = ref<string>('')
const streamingMessageIndex = ref<number | null>(null)

export function useAIChat() {
    const hasMessages = computed(() => messages.value.length > 0)

    async function sendMessage(content: string): Promise<void> {
        if (!content.trim() || isLoading.value) return

        // Add user message immediately (optimistic)
        const userMessage: ChatMessage = {
            role: 'user',
            content: content.trim(),
            timestamp: Date.now()
        }
        messages.value.push(userMessage)

        // Add placeholder assistant message for streaming
        const assistantMessage: ChatMessage = {
            role: 'assistant',
            content: '',
            timestamp: Date.now()
        }
        messages.value.push(assistantMessage)
        const assistantIndex = messages.value.length - 1
        streamingMessageIndex.value = assistantIndex

        isLoading.value = true
        error.value = null

        try {
            await sendChatMessageStream(
                content.trim(),
                sessionId.value,
                // onToken - append content to assistant message
                (token: ChatStreamToken) => {
                    if (token.sessionId) {
                        sessionId.value = token.sessionId
                    }
                    messages.value[assistantIndex].content += token.content
                },
                // onError
                (err: Error) => {
                    error.value = err.message
                    // Remove the empty assistant message on error
                    if (messages.value[assistantIndex].content === '') {
                        messages.value.splice(assistantIndex, 1)
                        // Also remove the user message
                        messages.value.splice(assistantIndex - 1, 1)
                    }
                    streamingMessageIndex.value = null
                    isLoading.value = false
                },
                // onComplete
                () => {
                    messages.value[assistantIndex].timestamp = Date.now()
                    streamingMessageIndex.value = null
                    isLoading.value = false
                }
            )
        } catch (e) {
            error.value = e instanceof Error ? e.message : 'Failed to send message'
            // Remove the placeholder messages on error
            messages.value.splice(assistantIndex - 1, 2)
            streamingMessageIndex.value = null
            isLoading.value = false
        }
    }

    async function loadHistory(): Promise<void> {
        if (!sessionId.value) return

        try {
            const history = await getChatHistory(sessionId.value)
            if (history.length > 0) {
                messages.value = history
            }
        } catch (e) {
            console.error('Failed to load chat history:', e)
        }
    }

    async function clearChat(): Promise<void> {
        try {
            await clearChatHistory(sessionId.value)
        } catch (e) {
            console.error('Failed to clear chat:', e)
        }

        // Reset local state
        messages.value = []
        sessionId.value = generateSessionId()
        lastContextSummary.value = ''
        error.value = null
        streamingMessageIndex.value = null
    }

    function isStreamingMessage(index: number): boolean {
        return streamingMessageIndex.value === index
    }

    return {
        // State
        sessionId,
        messages,
        isLoading,
        error,
        lastContextSummary,
        hasMessages,
        streamingMessageIndex,

        // Actions
        sendMessage,
        loadHistory,
        clearChat,
        isStreamingMessage
    }
}
