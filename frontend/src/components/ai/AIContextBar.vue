<!-- AI Context Bar - Phase 4: Contextual AI Assistant -->
<script setup lang="ts">
import { ref, computed } from 'vue'
import { MessageSquare, X, ChevronDown, ChevronUp } from 'lucide-vue-next'
import { useAI } from '../../composables/useAI'
import { useOmnibox } from '../../composables/useOmnibox'

const props = defineProps<{
  context?: {
    page?: string
    summary?: string
    insights?: string[]
  }
}>()

const { analyzeContext } = useAI()
const { toggle } = useOmnibox()

const expanded = ref(false)
const loading = ref(false)
const analysis = ref<string | null>(null)

const hasContext = computed(() => props.context && (props.context.summary || props.context.insights?.length))

const toggleExpand = async () => {
  expanded.value = !expanded.value
  if (expanded.value && !analysis.value && props.context) {
    // Analyze current context
    loading.value = true
    // This would call analyzeContext based on current page
    // For now, just set a placeholder
    setTimeout(() => {
      analysis.value = "I notice some interesting patterns in the current view. Would you like me to explain?"
      loading.value = false
    }, 1000)
  }
}

const askAI = () => {
  // Open omnibox instead of navigating to a non-existent page
  toggle()
}
</script>

<template>
  <Transition name="context-bar">
    <div v-if="hasContext" class="context-bar">
      <div class="context-bar-content">
        <div class="context-summary" @click="toggleExpand">
          <MessageSquare :size="16" />
          <span class="summary-text">
            {{ context?.summary || 'AI has insights about this page...' }}
          </span>
          <button class="expand-btn">
            <ChevronDown v-if="!expanded" :size="16" />
            <ChevronUp v-else :size="16" />
          </button>
        </div>

        <Transition name="expand">
          <div v-if="expanded" class="context-expanded">
            <div v-if="loading" class="loading">
              Analyzing context...
            </div>
            <div v-else-if="analysis" class="analysis">
              {{ analysis }}
            </div>
            <div v-if="context?.insights && context.insights.length > 0" class="insights">
              <div v-for="(insight, idx) in context.insights" :key="idx" class="insight-item">
                • {{ insight }}
              </div>
            </div>
            <div class="actions">
              <button class="action-btn primary" @click="askAI">
                <MessageSquare :size="14" />
                <span>Ask AI</span>
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.context-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 100;
  background: var(--bg-surface);
  border-top: 1px solid var(--border-default);
  box-shadow: 0 -4px 12px rgba(0, 0, 0, 0.1);
}

.context-bar-content {
  max-width: 1400px;
  margin: 0 auto;
  padding: 12px 24px;
}

.context-summary {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  padding: 8px;
  border-radius: var(--radius-md);
  transition: background 0.15s;
}

.context-summary:hover {
  background: var(--bg-hover);
}

.summary-text {
  flex: 1;
  font-size: 13px;
  color: var(--text-secondary);
}

.expand-btn {
  padding: 4px;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  border-radius: var(--radius-sm);
}

.context-expanded {
  margin-top: 12px;
  padding: 16px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
}

.loading {
  color: var(--text-muted);
  font-size: 13px;
}

.analysis {
  color: var(--text-primary);
  font-size: 14px;
  line-height: 1.6;
  margin-bottom: 12px;
}

.insights {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--border-subtle);
}

.insight-item {
  font-size: 13px;
  color: var(--text-secondary);
  margin-top: 6px;
  line-height: 1.5;
}

.actions {
  margin-top: 16px;
  display: flex;
  gap: 8px;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-surface);
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.action-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.action-btn.primary {
  background: var(--chart-network);
  color: white;
  border-color: var(--chart-network);
}

.action-btn.primary:hover {
  opacity: 0.9;
}

.context-bar-enter-active,
.context-bar-leave-active {
  transition: transform 0.3s ease, opacity 0.3s ease;
}

.context-bar-enter-from,
.context-bar-leave-to {
  transform: translateY(100%);
  opacity: 0;
}

.expand-enter-active,
.expand-leave-active {
  transition: all 0.2s ease;
}

.expand-enter-from,
.expand-leave-to {
  opacity: 0;
  max-height: 0;
  margin-top: 0;
  padding-top: 0;
  padding-bottom: 0;
}
</style>

