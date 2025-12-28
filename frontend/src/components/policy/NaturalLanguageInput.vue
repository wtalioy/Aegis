<!-- Natural Language Input Component - Phase 4 -->
<script setup lang="ts">
import { ref } from 'vue'
import { Sparkles } from 'lucide-vue-next'

const props = defineProps<{
  placeholder?: string
  value?: string
}>()

const emit = defineEmits<{
  'update:value': [value: string]
  submit: [value: string]
}>()

const input = ref(props.value || '')

const handleSubmit = () => {
  if (input.value.trim()) {
    emit('submit', input.value.trim())
  }
}

const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSubmit()
  }
}
</script>

<template>
  <div class="nl-input">
    <div class="input-wrapper">
      <Sparkles :size="18" class="icon" />
      <textarea
        v-model="input"
        class="input"
        :placeholder="placeholder || 'Describe your security intent in natural language...'"
        rows="4"
        @keydown="handleKeydown"
        @input="$emit('update:value', input)"
      />
    </div>
    <button class="submit-btn" @click="handleSubmit" :disabled="!input.trim()">
      <Sparkles :size="16" />
      <span>Generate</span>
    </button>
  </div>
</template>

<style scoped>
.nl-input {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.input-wrapper {
  position: relative;
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  transition: all 0.15s;
}

.input-wrapper:focus-within {
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 2px var(--accent-glow);
}

.icon {
  color: var(--accent-primary);
  flex-shrink: 0;
  margin-top: 2px;
}

.input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  font-size: 14px;
  font-family: inherit;
  color: var(--text-primary);
  resize: vertical;
  min-height: 80px;
}

.input::placeholder {
  color: var(--text-muted);
}

.submit-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: var(--accent-primary);
  color: white;
  border: none;
  border-radius: var(--radius-md);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
  align-self: flex-end;
}

.submit-btn:hover:not(:disabled) {
  opacity: 0.9;
}

.submit-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>

