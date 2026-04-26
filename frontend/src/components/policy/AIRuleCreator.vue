<!-- AI Rule Creator Component - Updated to use global styles -->
<script setup lang="ts">
import { ref } from 'vue'
import { Sparkles, Loader2 } from 'lucide-vue-next'
import { useAI } from '../../composables/useAI'
import type { RuleGenResponse } from '../../types/ai'

const { generateRule, loading } = useAI()

const emit = defineEmits<{
  'rule-generated': [rule: RuleGenResponse]
}>()

const description = ref('')
const error = ref<string | null>(null)

const createRule = async () => {
  if (!description.value.trim()) return

  error.value = null
  const result = await generateRule({
    description: description.value,
    context: {
      currentPage: 'policy-studio'
    }
  })

  if (result) {
    emit('rule-generated', result)
  } else {
    error.value = 'Failed to generate rule'
  }
}
</script>

<template>
  <div class="card-base">
    <div class="card-content-base">
      <div class="input-section">
        <label class="input-label">Describe your security intent:</label>
        <textarea v-model="description" class="description-input"
          placeholder="e.g., Block all outbound connections from nginx to port 3306" rows="6" />
        <button class="btn btn-primary" @click="createRule" :disabled="loading || !description.trim()">
          <Loader2 v-if="loading" :size="16" class="spin" />
          <Sparkles v-else :size="16" />
          <span>{{ loading ? 'Generating...' : 'Generate Rule' }}</span>
        </button>
      </div>

      <div v-if="error" class="error-message">
        {{ error }}
      </div>
    </div>
  </div>
</template>

<style scoped>
.input-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.input-label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 0;
}

.description-input {
  width: 100%;
  padding: 12px 14px;
  background: var(--bg-overlay);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  font-size: 14px;
  font-family: inherit;
  color: var(--text-primary);
  resize: vertical;
  min-height: 120px;
  line-height: 1.6;
  transition: all var(--transition-fast);
}

.description-input:focus {
  outline: none;
  border-color: var(--accent-primary);
  background: var(--bg-surface);
}

.btn-primary {
  width: 100%;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.error-message {
  padding: 12px 16px;
  background: var(--status-critical-dim);
  border: 1px solid var(--status-critical);
  border-radius: var(--radius-md);
  color: var(--status-critical);
  font-size: 13px;
  font-weight: 500;
  margin-top: 16px;
}
</style>
