<!-- Quick Ask AI Button Component - Phase 4 -->
<script setup lang="ts">
import { MessageSquare } from 'lucide-vue-next'
import { useOmnibox } from '../../composables/useOmnibox'

const { askQuestion, toggle } = useOmnibox()

const props = defineProps<{
  context?: string
  question?: string
}>()

const askAI = async (e: Event) => {
  e.preventDefault()
  e.stopPropagation()
  if (props.question) {
    await askQuestion(props.question)
  } else {
    toggle()
  }
}
</script>

<template>
  <button class="quick-ask-btn" @click="askAI" :title="question || 'Ask AI'">
    <MessageSquare :size="14" class="icon" />
    <span v-if="!question">Ask AI</span>
    <span v-else>{{ question }}</span>
  </button>
</template>

<style scoped>
.quick-ask-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 8px 14px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  height: auto; /* Auto height */
  width: 100%;
  box-sizing: border-box;
}

.quick-ask-btn .icon {
  flex-shrink: 0;
}

.quick-ask-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
  border-color: var(--border-default);
}
</style>