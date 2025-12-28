// AI Omnibox Composable - Simplified chat-only approach
// Using singleton pattern to share state across all components
import { ref, computed, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { sendChatMessage } from '../lib/api'

// Shared state - singleton pattern
const isOpen = ref(false)
const input = ref('')
const loading = ref(false)
const aiAnswer = ref<string | null>(null)
const answering = ref(false)
const recentQueries = ref<Array<{ input: string; timestamp: number }>>([])

// Clean AI response to remove prompt content and internal markers
function cleanAIResponse(message: string): string {
  let cleaned = message

  // Remove context markers (these are definitely not user-facing)
  cleaned = cleaned.replace(/\[END CONTEXT[^\]]*\]/gi, '').trim()
  cleaned = cleaned.replace(/\[LIVE SYSTEM CONTEXT[^\]]*\]/gi, '').trim()
  cleaned = cleaned.replace(/\[END CONTEXT - The data above is for your reference only[^\]]*\]/gi, '').trim()
  cleaned = cleaned.replace(/\[INTERNAL NOTE[^\]]*\]/gi, '').trim()

  // Remove specific instruction patterns that might leak from system prompt
  cleaned = cleaned.replace(/IMPORTANT:.*?use the real data.*?\./gis, '').trim()
  cleaned = cleaned.replace(/CRITICAL:.*?Answer the User's Question Directly[^\n]*/gi, '').trim()
  cleaned = cleaned.replace(/NEVER include context markers.*?in your response/gi, '').trim()
  cleaned = cleaned.replace(/Your response should only contain.*?no internal markers/gi, '').trim()
  cleaned = cleaned.replace(/Do not include this marker or any instructions in your response/gi, '').trim()

  // Remove internal context notes about false positives (common pattern that leaks)
  cleaned = cleaned.replace(/NOTE:\s*Not all alerts indicate real threats[^\n]*/gi, '').trim()
  cleaned = cleaned.replace(/Not all alerts indicate real threats[^\n]*/gi, '').trim()
  cleaned = cleaned.replace(/Common system utilities \(ls, cat, grep, ps, etc\.\) doing normal operations may be false positives/gi, '').trim()
  cleaned = cleaned.replace(/Common utilities \(ls, cat, grep\) accessing normal files are typically NOT suspicious/gi, '').trim()

  // Remove system prompt introduction fragments (but preserve legitimate content)
  cleaned = cleaned.replace(/^You are Aegis AI, an intelligent assistant[^\n]*$/gm, '').trim()
  cleaned = cleaned.replace(/^Your capabilities:[^\n]*$/gm, '').trim()

  // Remove instruction sections that are clearly from the prompt
  cleaned = cleaned.replace(/Answer ONLY what the user asked - be contextual and concise/gi, '').trim()
  cleaned = cleaned.replace(/Response Guidelines:[^\n]*/gi, '').trim()
  cleaned = cleaned.replace(/NEVER provide UI navigation instructions/gi, '').trim()
  cleaned = cleaned.replace(/NEVER tell users how to navigate the UI/gi, '').trim()
  cleaned = cleaned.replace(/Focus on providing information, not teaching users how to use the interface/gi, '').trim()

  // Remove prompt-like instruction blocks (multi-line patterns that are clearly instructions)
  cleaned = cleaned.replace(/1\. \*\*Understand the Query Intent\*\*:.*?2\. \*\*Answer Directly/gs, '').trim()
  cleaned = cleaned.replace(/ALWAYS answer the question using the actual data/gi, '').trim()
  cleaned = cleaned.replace(/NEVER output template placeholders/gi, '').trim()

  // Clean up multiple consecutive newlines
  cleaned = cleaned.replace(/\n{3,}/g, '\n\n').trim()

  return cleaned
}

export function useOmnibox() {
  const router = useRouter()

  const toggle = () => {
    isOpen.value = !isOpen.value
    if (isOpen.value) {
      input.value = ''
      aiAnswer.value = null
    }
  }

  const openWithQuery = (query: string) => {
    input.value = query
    isOpen.value = true
    aiAnswer.value = null
    // Don't auto-send - wait for user to press Enter
    // This gives them a chance to edit the query if needed
  }

  const askQuestion = async (question: string) => {
    // Set the input and open omnibox
    input.value = question
    isOpen.value = true
    aiAnswer.value = null

    // Wait for next tick to ensure omnibox is rendered
    await nextTick()

    // Send to AI chat
    answering.value = true
    try {
      const chatResponse = await sendChatMessage(question, 'omnibox-session')
      aiAnswer.value = cleanAIResponse(chatResponse.message)

      // Store in recent queries
      recentQueries.value.push({
        input: question,
        timestamp: Date.now()
      })
      if (recentQueries.value.length > 10) {
        recentQueries.value = recentQueries.value.slice(-10)
      }
    } catch (err) {
      console.error('Failed to get AI answer:', err)
    } finally {
      answering.value = false
    }
  }

  const close = () => {
    isOpen.value = false
    input.value = ''
    aiAnswer.value = null
  }

  const parseInput = async () => {
    if (!input.value.trim()) return

    aiAnswer.value = null

    // Show "AI is thinking" immediately
    answering.value = true
    loading.value = true

    try {
      const chatResponse = await sendChatMessage(input.value, 'omnibox-session')
      aiAnswer.value = cleanAIResponse(chatResponse.message)

      // Store in recent queries
      recentQueries.value.push({
        input: input.value,
        timestamp: Date.now()
      })
      if (recentQueries.value.length > 10) {
        recentQueries.value = recentQueries.value.slice(-10)
      }
    } catch (err) {
      console.error('Failed to get AI answer:', err)
    } finally {
      answering.value = false
      loading.value = false
    }
  }

  // Keyboard shortcuts
  const handleKeydown = (e: KeyboardEvent) => {
    // Cmd/Ctrl + K to toggle
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault()
      toggle()
    }
    // Escape to close
    if (e.key === 'Escape' && isOpen.value) {
      e.preventDefault()
      close()
    }
  }

  return {
    isOpen,
    input,
    loading,
    aiAnswer,
    answering,
    recentQueries: computed(() => recentQueries.value.slice().reverse()),
    toggle,
    openWithQuery,
    askQuestion,
    close,
    parseInput,
    handleKeydown
  }
}
