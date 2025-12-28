<!-- AI Confidence Badge Component - Phase 4 -->
<script setup lang="ts">
import { computed } from 'vue'
import { CheckCircle2, AlertTriangle, HelpCircle } from 'lucide-vue-next'

const props = defineProps<{
  confidence: number
  size?: 'sm' | 'md' | 'lg'
}>()

const level = computed(() => {
  if (props.confidence >= 0.8) return 'high'
  if (props.confidence >= 0.5) return 'medium'
  return 'low'
})

const icon = computed(() => {
  switch (level.value) {
    case 'high':
      return CheckCircle2
    case 'medium':
      return AlertTriangle
    default:
      return HelpCircle
  }
})

const sizeClass = computed(() => `size-${props.size || 'md'}`)
</script>

<template>
  <div class="confidence-badge" :class="[level, sizeClass]">
    <component :is="icon" :size="size === 'sm' ? 12 : size === 'lg' ? 18 : 14" />
    <span class="confidence-value">{{ Math.round(confidence * 100) }}%</span>
  </div>
</template>

<style scoped>
.confidence-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  font-weight: 600;
  font-size: 12px;
}

.confidence-badge.size-sm {
  padding: 2px 6px;
  font-size: 11px;
}

.confidence-badge.size-lg {
  padding: 6px 12px;
  font-size: 14px;
}

.confidence-badge.high {
  background: rgba(34, 197, 94, 0.1);
  color: rgb(34, 197, 94);
}

.confidence-badge.medium {
  background: rgba(251, 191, 36, 0.1);
  color: rgb(251, 191, 36);
}

.confidence-badge.low {
  background: rgba(239, 68, 68, 0.1);
  color: rgb(239, 68, 68);
}

.confidence-value {
  font-variant-numeric: tabular-nums;
}
</style>

