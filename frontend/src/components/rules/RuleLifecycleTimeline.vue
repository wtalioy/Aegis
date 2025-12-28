<script setup lang="ts">
import { CheckCircle2, Circle } from 'lucide-vue-next'
import type { Rule } from '../../types/rules'

defineProps<{
  rule: Rule
}>()

const formatDate = (date: string | number | null | undefined, fallbackState?: string) => {
  if (!date) {
    // If no date but we have state info, show appropriate message
    if (fallbackState === 'deployed') {
      return 'Deployed'
    }
    return 'Not yet'
  }
  try {
    const d = typeof date === 'string' ? new Date(date) : new Date(date)
    // Check if date is valid (not NaN and not epoch 0)
    if (isNaN(d.getTime()) || d.getTime() === 0) {
      if (fallbackState === 'deployed') {
        return 'Deployed'
      }
      return 'Not yet'
    }
    // Format as YYYY-MM-DD or locale string
    const year = d.getFullYear()
    const month = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    return `${year}-${month}-${day}`
  } catch {
    if (fallbackState === 'deployed') {
      return 'Deployed'
    }
    return 'Not yet'
  }
}

const isDeployed = (rule: Rule) => {
  const state = (rule as any).state
  return state === 'testing' || state === 'production'
}

const isProduction = (rule: Rule) => {
  return (rule as any).state === 'production'
}
</script>

<template>
  <div class="lifecycle-timeline">
    <h3>Rule Lifecycle</h3>

    <div class="timeline">
      <div class="timeline-item">
        <div class="timeline-marker done">
          <CheckCircle2 :size="16" />
        </div>
        <div class="timeline-content">
          <div class="timeline-label">Created</div>
          <div class="timeline-date">{{ formatDate((rule as any).created_at) }}</div>
        </div>
      </div>

      <div class="timeline-item">
        <div class="timeline-marker" :class="{ done: isDeployed(rule) }">
          <CheckCircle2 v-if="isDeployed(rule)" :size="16" />
          <Circle v-else :size="16" />
        </div>
        <div class="timeline-content">
          <div class="timeline-label">Deployed to Testing</div>
          <div class="timeline-date">{{ formatDate((rule as any).deployed_at, isDeployed(rule) ? 'deployed' : undefined)
          }}</div>
        </div>
      </div>

      <div class="timeline-item">
        <div class="timeline-marker" :class="{ done: isProduction(rule) }">
          <CheckCircle2 v-if="isProduction(rule)" :size="16" />
          <Circle v-else :size="16" />
        </div>
        <div class="timeline-content">
          <div class="timeline-label">Promoted to Production</div>
          <div class="timeline-date">{{ formatDate((rule as any).promoted_at, isProduction(rule) ? 'deployed' :
            undefined) }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.lifecycle-timeline {
  padding: 16px;
  background: var(--bg-elevated);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
}

.lifecycle-timeline h3 {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 16px 0;
}

.timeline {
  display: flex;
  flex-direction: column;
  gap: 16px;
  position: relative;
}

.timeline::before {
  content: '';
  position: absolute;
  left: 7px;
  top: 24px;
  bottom: 0;
  width: 2px;
  background: var(--border-subtle);
}

.timeline-item {
  display: flex;
  gap: 12px;
  position: relative;
  z-index: 1;
}

.timeline-marker {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-surface);
  color: var(--text-muted);
  flex-shrink: 0;
  margin-top: 2px;
}

.timeline-marker.done {
  background: rgba(34, 197, 94, 0.1);
  color: rgb(34, 197, 94);
}

.timeline-content {
  flex: 1;
  padding-top: 2px;
}

.timeline-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
}

.timeline-date {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 2px;
}
</style>
