<!-- AI Rule Creator Component - Phase 4 -->
<script setup lang="ts">
import { ref } from 'vue'
import { Sparkles, Loader2 } from 'lucide-vue-next'
import { useAI } from '../../composables/useAI'

const { generateRule, loading } = useAI()

const emit = defineEmits<{
  'rule-generated': [rule: any]
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
  <div class="rule-creator">
    <div class="creator-body">
      <div class="input-section">
        <label class="input-label">Describe your security intent:</label>
        <textarea v-model="description" class="description-input"
          placeholder="e.g., Block all outbound connections from nginx to port 3306" rows="6" />
        <button class="generate-btn" @click="createRule" :disabled="loading || !description.trim()">
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
.rule-creator {
  padding: 28px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-md);
  backdrop-filter: blur(10px);
}

.creator-body {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.input-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.input-label {
  display: block;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  margin-bottom: 4px;
  text-transform: uppercase;
  letter-spacing: 0.6px;
}

.description-input {
  width: 100%;
  padding: 14px 16px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  font-size: 14px;
  font-family: inherit;
  color: var(--text-primary);
  resize: none;
  min-height: 140px;
  max-height: 240px;
  overflow-y: auto;
  line-height: 1.6;
  transition: all var(--transition-normal);
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.1);
}

.description-input:hover {
  border-color: var(--border-default);
  background: var(--bg-surface);
}

.description-input:focus {
  outline: none;
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 3px var(--accent-glow), inset 0 1px 2px rgba(0, 0, 0, 0.1);
  background: var(--bg-surface);
}

.generate-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 14px 24px;
  background: var(--accent-primary);
  color: white;
  border: none;
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all var(--transition-normal);
  box-shadow: 0 2px 8px rgba(96, 165, 250, 0.3);
  width: 100%;
  position: relative;
  overflow: hidden;
}

.generate-btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
  transition: left var(--transition-slow);
}

.generate-btn:hover:not(:disabled) {
  background: var(--accent-primary-hover);
  box-shadow: 0 4px 12px rgba(96, 165, 250, 0.4);
  transform: translateY(-2px);
}

.generate-btn:hover:not(:disabled)::before {
  left: 100%;
}

.generate-btn:active:not(:disabled) {
  transform: translateY(0);
  box-shadow: 0 2px 6px rgba(96, 165, 250, 0.3);
}

.generate-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}

.error-message {
  padding: 14px 16px;
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.12) 0%, rgba(239, 68, 68, 0.06) 100%);
  border: 1px solid var(--status-critical);
  border-radius: var(--radius-md);
  color: var(--status-critical);
  font-size: 13px;
  font-weight: 500;
  box-shadow: 0 0 0 1px rgba(239, 68, 68, 0.2);
}
</style>
