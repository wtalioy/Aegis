<!-- Deep Ask AI Component - Phase 4 -->
<script setup lang="ts">
import { ref } from 'vue'
import { MessageSquare, Send, Loader2 } from 'lucide-vue-next'
import { useAI } from '../../composables/useAI'
import type { Insight } from '../../types/sentinel'

const props = defineProps<{
  insight: Insight
}>()

const { analyzeContext, loading } = useAI()
const question = ref('')
const response = ref<string | null>(null)

const ask = async () => {
  if (!question.value.trim()) return

  // Use analyzeContext to get deep insights
  const analysis = await analyzeContext({
    type: 'process',
    id: props.insight.data.process_id?.toString() || ''
  })

  if (analysis) {
    response.value = analysis.summary
  }
}

const clear = () => {
  question.value = ''
  response.value = null
}
</script>

<template>
  <div class="deep-ask-ai">
    <div class="header">
      <MessageSquare :size="18" />
      <h3>Deep Ask AI</h3>
    </div>

    <div class="insight-context">
      <div class="context-label">About this insight:</div>
      <div class="context-text">{{ insight.summary }}</div>
    </div>

    <div class="input-section">
      <textarea
        v-model="question"
        class="question-input"
        placeholder="Ask a deeper question about this insight..."
        rows="3"
      />
      <button class="ask-btn" @click="ask" :disabled="loading || !question.trim()">
        <Loader2 v-if="loading" :size="16" class="spin" />
        <Send v-else :size="16" />
        <span>{{ loading ? 'Thinking...' : 'Ask' }}</span>
      </button>
    </div>

    <div v-if="response" class="response-section">
      <div class="response-label">AI Response:</div>
      <div class="response-text">{{ response }}</div>
      <button class="clear-btn" @click="clear">Clear</button>
    </div>
  </div>
</template>

<style scoped>
.deep-ask-ai {
  padding: 20px;
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
}

.header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.header h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.insight-context {
  margin-bottom: 20px;
  padding: 12px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
}

.context-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 6px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.context-text {
  font-size: 13px;
  line-height: 1.5;
  color: var(--text-secondary);
}

.input-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 20px;
}

.question-input {
  width: 100%;
  padding: 12px;
  background: var(--bg-void);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  font-size: 14px;
  font-family: inherit;
  color: var(--text-primary);
  resize: vertical;
}

.question-input:focus {
  outline: none;
  border-color: var(--chart-network);
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.ask-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: var(--chart-network);
  color: white;
  border: none;
  border-radius: var(--radius-md);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
  align-self: flex-end;
}

.ask-btn:hover:not(:disabled) {
  opacity: 0.9;
}

.ask-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.response-section {
  padding: 16px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
}

.response-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.response-text {
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-primary);
  margin-bottom: 12px;
}

.clear-btn {
  padding: 6px 12px;
  background: var(--bg-void);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  font-size: 12px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.15s;
}

.clear-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}
</style>

