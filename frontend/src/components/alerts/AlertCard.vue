<script setup lang="ts">
import { computed } from 'vue'
import type { Alert } from '../../lib/api'

const props = defineProps<{
  alert: Alert
  isSelected: boolean
}>()

defineEmits<{
  select: [alert: Alert]
}>()

const severityLabel = computed(() => {
  switch (props.alert.severity) {
    case 'high': return 'HIGH'
    case 'warning': return 'WARN'
    default: return 'INFO'
  }
})

const formatTime = (timestamp: number) => {
  return new Date(timestamp).toLocaleTimeString('en-US', {
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}
</script>

<template>
  <div 
    class="alert-card" 
    :class="[`severity-${alert.severity}`, { 'is-selected': isSelected }]"
    @click="$emit('select', alert)"
  >
    <div class="alert-indicator"></div>
    <div class="alert-content">
      <div class="alert-header">
        <span class="alert-severity" :class="alert.severity">{{ severityLabel }}</span>
        <span class="alert-time font-mono">{{ formatTime(alert.timestamp) }}</span>
      </div>
      <div class="alert-title">{{ alert.ruleName }}</div>
      <div class="alert-process">
        <code>{{ alert.processName }}</code>
        <span class="alert-pid">PID {{ alert.pid }}</span>
        <span v-if="alert.cgroupId" class="alert-cgroup font-mono">cg:{{ alert.cgroupId.slice(0, 8) }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.alert-card {
  position: relative;
  padding: 12px 16px 12px 20px;
  background: var(--bg-elevated);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  margin-bottom: 8px;
}

.alert-card:hover {
  background: var(--bg-overlay);
}

.alert-indicator {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 4px;
  border-radius: 8px 0 0 8px;
  background: var(--text-muted);
}

.severity-high .alert-indicator { background: var(--status-critical); }
.severity-warning .alert-indicator { background: var(--status-warning); }
.severity-info .alert-indicator { background: var(--status-info); }

.alert-card.is-selected {
  background: var(--bg-overlay);
  box-shadow: 0 0 0 2px var(--border-focus);
}

.severity-high.is-selected {
  box-shadow: 0 0 0 2px rgba(239, 68, 68, 0.5), var(--glow-critical);
}

.severity-warning.is-selected {
  box-shadow: 0 0 0 2px rgba(245, 158, 11, 0.5), 0 0 20px rgba(245, 158, 11, 0.3);
}

.alert-content {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.alert-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.alert-severity {
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
}

.alert-severity.high {
  background: var(--status-critical-dim);
  color: var(--status-critical);
}

.alert-severity.warning {
  background: var(--status-warning-dim);
  color: var(--status-warning);
}

.alert-severity.info {
  background: var(--status-info-dim);
  color: var(--status-info);
}

.alert-time {
  font-size: 11px;
  color: var(--text-muted);
}

.alert-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  line-height: 1.3;
}

.alert-process {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.alert-process code {
  font-family: var(--font-mono);
  color: var(--text-secondary);
  background: var(--bg-surface);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
}

.alert-pid {
  color: var(--text-muted);
  font-family: var(--font-mono);
  font-size: 11px;
}

.alert-cgroup {
  color: var(--text-muted);
  font-size: 10px;
  padding: 1px 4px;
  background: var(--bg-surface);
  border-radius: var(--radius-sm);
}

/* New alert pulse animation */
@keyframes alert-pulse {
  0%, 100% { 
    box-shadow: 0 0 0 0 var(--glow-critical); 
  }
  50% { 
    box-shadow: 0 0 20px 4px var(--glow-critical); 
  }
}

.alert-card.severity-high.is-new {
  animation: alert-pulse 0.6s ease-out 2;
}
</style>

