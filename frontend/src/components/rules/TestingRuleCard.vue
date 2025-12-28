<script setup lang="ts">
import { Activity, AlertTriangle, Clock } from 'lucide-vue-next'
import type { Rule } from '../../types/rules'

defineProps<{
  rule: Rule
  selected?: boolean
}>()

defineEmits<{
  select: []
}>()
</script>

<template>
  <div class="testing-rule-card" :class="{ selected }" @click="$emit('select')">
    <div class="card-header">
      <h4 class="rule-name">{{ rule.name }}</h4>
      <span class="mode-badge">Testing</span>
    </div>

    <div class="card-stats">
      <div class="stat">
        <Activity :size="14" />
        <span class="stat-label">Hits</span>
        <span class="stat-value">{{ (rule as any).actual_testing_hits || 0 }}</span>
      </div>
      <div class="stat">
        <Clock :size="14" />
        <span class="stat-label">Observed</span>
        <span class="stat-value">{{ (() => {
          const minutes = (rule as any).observationMinutes || ((rule as any).observationHours ? Math.round((rule as any).observationHours * 60) : 0) || 0
          if (minutes < 60) return `${minutes}min`
          return `${(minutes / 60).toFixed(1)}h`
        })() }}</span>
      </div>
    </div>

    <div class="readiness-bar">
      <div class="bar-fill" :style="{ width: `${(((rule as any).promotion_score || (rule as any).promotionScore || 0) * 100)}%` }"></div>
    </div>
  </div>
</template>

<style scoped>
.testing-rule-card {
  padding: 12px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.15s;
}

.testing-rule-card:hover {
  background: var(--bg-hover);
  border-color: var(--border-default);
}

.testing-rule-card.selected {
  border-color: var(--accent-primary);
  background: rgba(59, 130, 246, 0.05);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.rule-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.mode-badge {
  padding: 2px 6px;
  background: rgba(59, 130, 246, 0.1);
  color: rgb(59, 130, 246);
  font-size: 10px;
  font-weight: 600;
  border-radius: var(--radius-sm);
  text-transform: uppercase;
  letter-spacing: 0.3px;
  flex-shrink: 0;
  margin-left: 8px;
}

.card-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  margin-bottom: 8px;
}

.stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 6px;
  background: var(--bg-surface);
  border-radius: var(--radius-sm);
}

.stat svg {
  color: var(--text-muted);
  flex-shrink: 0;
}

.stat-label {
  font-size: 9px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.stat-value {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
}

.readiness-bar {
  height: 4px;
  background: var(--bg-surface);
  border-radius: var(--radius-full);
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  background: linear-gradient(90deg, rgb(251, 191, 36), rgb(34, 197, 94));
  transition: width 0.3s;
}
</style>

