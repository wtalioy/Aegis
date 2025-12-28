<!-- Query Translation Component - Phase 4 -->
<script setup lang="ts">
import { computed } from 'vue'
import { Sparkles, CheckCircle2 } from 'lucide-vue-next'

const props = defineProps<{
  naturalLanguage: string
  translatedQuery?: {
    filter?: any
    semantic?: string
  }
}>()

const querySummary = computed(() => {
  if (!props.translatedQuery?.filter) return null
  
  const filter = props.translatedQuery.filter
  const parts: string[] = []
  
  if (filter.types && filter.types.length > 0) {
    parts.push(`Event types: ${filter.types.join(', ')}`)
  }
  if (filter.processes && filter.processes.length > 0) {
    parts.push(`Processes: ${filter.processes.join(', ')}`)
  }
  if (filter.pids && filter.pids.length > 0) {
    parts.push(`PIDs: ${filter.pids.join(', ')}`)
  }
  if (filter.timeWindow) {
    parts.push('Time window specified')
  }
  
  return parts.length > 0 ? parts.join(' • ') : 'All events'
})
</script>

<template>
  <Transition name="slide">
    <div v-if="translatedQuery" class="query-understanding">
      <div class="understanding-header">
        <CheckCircle2 :size="16" class="icon" />
        <span class="header-text">AI Understanding</span>
      </div>
      <div class="understanding-content">
        <div class="query-display">
          <span class="query-text">"{{ naturalLanguage }}"</span>
        </div>
        <div v-if="querySummary" class="summary">
          <Sparkles :size="12" />
          <span>{{ querySummary }}</span>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.query-understanding {
  padding: 14px 16px;
  background: rgba(59, 130, 246, 0.05);
  border: 1px solid rgba(59, 130, 246, 0.2);
  border-radius: var(--radius-md);
  margin-bottom: 16px;
  animation: slideDown 0.2s ease-out;
}

.understanding-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.icon {
  color: var(--chart-network);
  flex-shrink: 0;
}

.header-text {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.understanding-content {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.query-display {
  padding: 10px 12px;
  background: var(--bg-void);
  border-radius: var(--radius-sm);
  border-left: 3px solid var(--chart-network);
}

.query-text {
  font-size: 13px;
  line-height: 1.5;
  color: var(--text-primary);
  font-style: italic;
}

.summary {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-secondary);
  padding-left: 4px;
}

.summary svg {
  color: var(--chart-network);
  flex-shrink: 0;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.slide-enter-active {
  transition: all 0.2s ease-out;
}

.slide-leave-active {
  transition: all 0.15s ease-in;
}

.slide-enter-from {
  opacity: 0;
  transform: translateY(-8px);
}

.slide-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>

