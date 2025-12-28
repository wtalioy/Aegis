<!-- AI Omnibox - Simplified chat-only approach -->
<script setup lang="ts">
import { computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { Search, Sparkles, Loader2, ArrowRight, X } from 'lucide-vue-next'
import { marked } from 'marked'
import { useOmnibox } from '../../composables/useOmnibox'

const { isOpen, input, loading, aiAnswer, answering, recentQueries, toggle, close, parseInput, handleKeydown } = useOmnibox()

// Convert AI answer to markdown HTML
const aiAnswerHtml = computed(() => {
  if (!aiAnswer.value) return ''
  try {
    return marked.parse(aiAnswer.value, { breaks: true })
  } catch (err) {
    console.error('Failed to parse markdown:', err)
    return aiAnswer.value
  }
})

const suggestions = [
  'Block all outbound connections from nginx to port 3306',
  'Why was this process blocked?',
  'Show me suspicious file access in the last hour',
  'Create a whitelist for my Redis container'
]

// Clear answer when input is cleared
watch(input, () => {
  if (!input.value.trim()) {
    aiAnswer.value = null
  }
})

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})

const handleEnter = async (e: KeyboardEvent) => {
  if (e.shiftKey) return // Allow Shift+Enter for new lines if needed
  
  e.preventDefault()
  
  if (input.value.trim() && !loading.value) {
    await parseInput()
  }
}

const handleSuggestionClick = async (e: Event, suggestion: string) => {
  e.preventDefault()
  e.stopPropagation()
  input.value = suggestion
  // Wait for next tick to ensure input is set
  await nextTick()
  // Automatically send to AI
  await parseInput()
}
</script>

<template>
  <Teleport to="body">
    <Transition name="omnibox">
      <div v-if="isOpen" class="omnibox-overlay" @click.self="close">
        <div class="omnibox-container">
          <!-- Header -->
          <div class="omnibox-header">
            <div class="omnibox-title">
              <Sparkles :size="20" />
              <span>Ask Aegis anything...</span>
            </div>
            <button class="close-btn" @click="close">
              <X :size="18" />
            </button>
          </div>

          <!-- Input -->
          <div class="omnibox-input-wrapper">
            <Search :size="18" class="search-icon" />
            <input
              v-model="input"
              type="text"
              class="omnibox-input"
              placeholder="Ask a question or describe what you want to do..."
              autofocus
              @keydown.enter="handleEnter"
            />
          </div>

          <!-- Scrollable Content Area -->
          <div class="omnibox-content">
            <!-- Loading State -->
            <div v-if="loading || answering" class="loading-state">
              <Loader2 :size="20" class="spin" />
              <span>{{ answering ? 'AI is thinking...' : 'Processing...' }}</span>
            </div>

            <!-- AI Answer -->
            <div v-else-if="aiAnswer" class="ai-answer">
              <div class="ai-answer-header">
                <Sparkles :size="16" />
                <span>AI Answer</span>
              </div>
              <div class="ai-answer-content" v-html="aiAnswerHtml"></div>
            </div>

            <!-- Suggestions -->
            <div v-else-if="!input && !loading" class="suggestions">
              <div class="suggestions-header">
                <span>💡 Try:</span>
              </div>
              <div class="suggestions-list">
                <button
                  v-for="(suggestion, idx) in suggestions"
                  :key="idx"
                  class="suggestion-item"
                  @click="(e) => handleSuggestionClick(e, suggestion)"
                >
                  <ArrowRight :size="14" />
                  <span>{{ suggestion }}</span>
                </button>
              </div>
            </div>

            <!-- Recent Queries -->
            <div v-if="recentQueries.length > 0 && !input && !loading && !aiAnswer" class="recent">
              <div class="recent-header">
                <span>📝 Recent:</span>
              </div>
              <div class="recent-list">
                <button
                  v-for="(recent, idx) in recentQueries.slice(0, 5)"
                  :key="idx"
                  class="recent-item"
                  @click="(e) => handleSuggestionClick(e, recent.input)"
                >
                  <span class="recent-input">"{{ recent.input }}"</span>
                </button>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div class="omnibox-footer">
            <div class="shortcuts">
              <kbd>Enter</kbd> to send
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.omnibox-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  z-index: 9999;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 10vh;
}

.omnibox-container {
  width: 100%;
  max-width: 700px;
  max-height: 85vh;
  background: var(--bg-surface);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  animation: slideDown 0.2s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.omnibox-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-subtle);
}

.omnibox-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.close-btn {
  padding: 4px;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: var(--radius-sm);
  transition: all 0.15s;
}

.close-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.omnibox-input-wrapper {
  position: relative;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 2px;
  flex-shrink: 0;
  border-bottom: 1px solid var(--border-subtle);
}

.omnibox-content {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  min-height: 0;
}

.search-icon {
  color: var(--text-muted);
  flex-shrink: 0;
}

.omnibox-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  font-size: 16px;
  color: var(--text-primary);
  font-family: inherit;
}

.omnibox-input::placeholder {
  color: var(--text-muted);
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.suggestions,
.recent {
  padding: 20px;
}

.suggestions-header,
.recent-header {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.suggestions-list,
.recent-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.suggestion-item,
.recent-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.15s;
  text-align: left;
  width: 100%;
  color: var(--text-secondary);
}

.suggestion-item:hover,
.recent-item:hover {
  background: var(--bg-hover);
  border-color: var(--border-default);
  color: var(--text-primary);
}

.recent-input {
  flex: 1;
  font-style: italic;
}

.omnibox-footer {
  padding: 12px 20px;
  border-top: 1px solid var(--border-subtle);
  background: var(--bg-elevated);
}

.shortcuts {
  display: flex;
  gap: 6px;
  font-size: 12px;
  color: var(--text-muted);
  align-items: center;
}

kbd {
  padding: 2px 6px;
  background: var(--bg-void);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  font-family: monospace;
  font-size: 11px;
}

.ai-answer {
  padding: 20px;
  border-top: 1px solid var(--border-subtle);
  background: var(--bg-elevated);
}

.ai-answer-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.ai-answer-header svg {
  color: var(--accent-primary);
}

.ai-answer-content {
  font-size: 14px;
  line-height: 1.7;
  color: var(--text-secondary);
  word-wrap: break-word;
  overflow-wrap: break-word;
}

/* Markdown styling */
.ai-answer-content :deep(h1),
.ai-answer-content :deep(h2),
.ai-answer-content :deep(h3) {
  color: var(--text-primary);
  font-weight: 600;
  margin-top: 16px;
  margin-bottom: 8px;
}

.ai-answer-content :deep(h1) {
  font-size: 18px;
}

.ai-answer-content :deep(h2) {
  font-size: 16px;
}

.ai-answer-content :deep(h3) {
  font-size: 14px;
}

.ai-answer-content :deep(p) {
  margin-bottom: 12px;
}

.ai-answer-content :deep(ul),
.ai-answer-content :deep(ol) {
  margin-bottom: 12px;
  padding-left: 24px;
}

.ai-answer-content :deep(li) {
  margin-bottom: 6px;
}

.ai-answer-content :deep(strong) {
  color: var(--text-primary);
  font-weight: 600;
}

.ai-answer-content :deep(code) {
  background: var(--bg-void);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 12px;
  color: var(--accent-primary);
}

.ai-answer-content :deep(pre) {
  background: var(--bg-void);
  padding: 12px;
  border-radius: var(--radius-md);
  overflow-x: auto;
  margin-bottom: 12px;
}

.ai-answer-content :deep(pre code) {
  background: transparent;
  padding: 0;
  color: var(--text-primary);
}

.ai-answer-content :deep(blockquote) {
  border-left: 3px solid var(--accent-primary);
  padding-left: 12px;
  margin: 12px 0;
  color: var(--text-muted);
  font-style: italic;
}

/* Custom scrollbar for content area */
.omnibox-content::-webkit-scrollbar {
  width: 6px;
}

.omnibox-content::-webkit-scrollbar-track {
  background: var(--bg-void);
  border-radius: var(--radius-sm);
}

.omnibox-content::-webkit-scrollbar-thumb {
  background: var(--border-subtle);
  border-radius: var(--radius-sm);
}

.omnibox-content::-webkit-scrollbar-thumb:hover {
  background: var(--border-default);
}

.loading-state {
  padding: 32px 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  color: var(--text-secondary);
  font-size: 14px;
  border-top: 1px solid var(--border-subtle);
  min-height: 120px;
}

.loading-state svg {
  color: var(--accent-primary);
}

.omnibox-enter-active,
.omnibox-leave-active {
  transition: opacity 0.2s;
}

.omnibox-enter-from,
.omnibox-leave-to {
  opacity: 0;
}
</style>
