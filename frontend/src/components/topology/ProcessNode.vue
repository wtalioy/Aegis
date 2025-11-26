<script setup lang="ts">
import { computed } from 'vue'
import { Box, AlertTriangle } from 'lucide-vue-next'

export interface ProcessNodeData {
  pid: number
  ppid: number
  comm: string
  timestamp: number
  cgroupId?: string
  isTarget?: boolean
  isAlertSource?: boolean
  severity?: 'high' | 'warning' | 'info'
}

const props = defineProps<{
  data: ProcessNodeData
}>()

const formatTime = (timestamp: number) => {
  if (!timestamp) return ''
  return new Date(timestamp).toLocaleTimeString('en-US', {
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

const nodeClass = computed(() => ({
  'is-target': props.data.isTarget,
  'is-alert-source': props.data.isAlertSource,
  [`severity-${props.data.severity}`]: props.data.severity
}))

// Generate a consistent color from cgroup ID for workload indicator
const workloadColor = computed(() => {
  if (!props.data.cgroupId || props.data.cgroupId === '0') return null
  let hash = 0
  for (let i = 0; i < props.data.cgroupId.length; i++) {
    hash = props.data.cgroupId.charCodeAt(i) + ((hash << 5) - hash)
  }
  const hue = Math.abs(hash % 360)
  return `hsl(${hue}, 60%, 50%)`
})

// Short cgroup ID for display
const shortCgroupId = computed(() => {
  if (!props.data.cgroupId) return ''
  const id = props.data.cgroupId
  return id.length > 8 ? id.slice(0, 8) + '…' : id
})
</script>

<template>
  <div class="process-node" :class="nodeClass">
    <div v-if="workloadColor" class="workload-indicator" :style="{ borderColor: workloadColor }"></div>
    
    <div class="node-header">
      <div class="node-icon">
        <AlertTriangle v-if="data.isAlertSource" :size="14" class="alert-icon" />
        <Box v-else :size="14" />
      </div>
      <span class="node-comm">{{ data.comm }}</span>
      <span v-if="data.severity" class="severity-badge" :class="data.severity">
        {{ data.severity.toUpperCase() }}
      </span>
    </div>
    
    <div class="node-meta">
      <div class="meta-row">
        <span class="meta-label">PID:</span>
        <span class="meta-value">{{ data.pid }}</span>
      </div>
      <div v-if="data.ppid" class="meta-row">
        <span class="meta-label">PPID:</span>
        <span class="meta-value">{{ data.ppid }}</span>
      </div>
      <div v-if="data.timestamp" class="meta-row">
        <span class="meta-label">Time:</span>
        <span class="meta-value">{{ formatTime(data.timestamp) }}</span>
      </div>
      <div v-if="shortCgroupId" class="meta-row">
        <span class="meta-label">Cgroup:</span>
        <span class="meta-value">{{ shortCgroupId }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.process-node {
  position: relative;
  background: var(--bg-elevated);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  padding: 12px 16px;
  min-width: 200px;
  max-width: 280px;
  user-select: none;
}

.process-node.is-target {
  background: var(--bg-overlay);
  border-color: var(--accent-primary);
  box-shadow: var(--glow-accent);
}

.process-node.is-alert-source {
  border-color: var(--status-critical);
}

.process-node.is-alert-source.severity-high {
  box-shadow: var(--glow-critical);
}

.process-node.is-alert-source.severity-warning {
  border-color: var(--status-warning);
  box-shadow: 0 0 20px rgba(245, 158, 11, 0.3);
}

.workload-indicator {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  border-top: 3px solid;
  border-radius: var(--radius-md) var(--radius-md) 0 0;
}

.node-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.node-icon {
  color: var(--text-muted);
  display: flex;
  align-items: center;
}

.node-icon .alert-icon {
  color: var(--status-critical);
}

.process-node.severity-warning .node-icon .alert-icon {
  color: var(--status-warning);
}

.node-comm {
  font-family: var(--font-mono);
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  flex: 1;
}

.severity-badge {
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  font-size: 9px;
  font-weight: 700;
  text-transform: uppercase;
}

.severity-badge.high {
  background: var(--status-critical-dim);
  color: var(--status-critical);
}

.severity-badge.warning {
  background: var(--status-warning-dim);
  color: var(--status-warning);
}

.severity-badge.info {
  background: var(--status-info-dim);
  color: var(--status-info);
}

.node-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.meta-row {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
}

.meta-label {
  color: var(--text-muted);
}

.meta-value {
  font-family: var(--font-mono);
  color: var(--text-secondary);
}
</style>
