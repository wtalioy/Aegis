<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  value: number | string
  label: string
  icon?: any
  trend?: 'up' | 'down' | 'neutral'
  color?: 'default' | 'safe' | 'warning' | 'critical' | 'info'
}>()

const colorClass = computed(() => {
  return props.color ? `color-${props.color}` : 'color-default'
})
</script>

<template>
  <div class="stat-card" :class="colorClass">
    <div class="stat-icon" v-if="icon">
      <component :is="icon" :size="24" />
    </div>
    <div class="stat-content">
      <div class="stat-value">{{ typeof value === 'number' ? value.toLocaleString() : value }}</div>
      <div class="stat-label">{{ label }}</div>
    </div>
  </div>
</template>

<style scoped>
.stat-card {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md);
  background: var(--bg-elevated);
  color: var(--text-muted);
}

.color-safe .stat-icon {
  background: var(--status-safe-dim);
  color: var(--status-safe);
}

.color-warning .stat-icon {
  background: var(--status-warning-dim);
  color: var(--status-warning);
}

.color-critical .stat-icon {
  background: var(--status-critical-dim);
  color: var(--status-critical);
}

.color-info .stat-icon {
  background: var(--status-info-dim);
  color: var(--status-info);
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  font-family: var(--font-mono);
  color: var(--text-primary);
  line-height: 1.2;
}

.stat-label {
  font-size: 13px;
  color: var(--text-muted);
  margin-top: 4px;
}
</style>

