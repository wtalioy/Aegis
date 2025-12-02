<script setup lang="ts">
import { computed } from 'vue'
import { Terminal, Globe, FileText } from 'lucide-vue-next'
import type { StreamEvent } from '../../lib/api'

type EventWithId = StreamEvent & { id: string }

const props = defineProps<{
  event: EventWithId
  isSelected: boolean
}>()

defineEmits<{
  select: [event: EventWithId]
}>()

const formatTime = (timestamp: number) => {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('en-US', {
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  }) + '.' + String(date.getMilliseconds()).padStart(3, '0')
}

const typeIcon = computed(() => {
  switch (props.event.type) {
    case 'exec': return Terminal
    case 'connect': return Globe
    case 'file': return FileText
  }
})

const typeClass = computed(() => `type-${props.event.type}`)

const details = computed(() => {
  switch (props.event.type) {
    case 'exec':
      return `${props.event.parentComm} → ${props.event.comm}`
    case 'connect':
      return `${props.event.addr}`
    case 'file':
      return props.event.filename
  }
})

const processName = computed(() => {
  if (props.event.type === 'exec') {
    return props.event.comm
  }
  if (props.event.type === 'connect' && props.event.processName) {
    return props.event.processName
  }
  return `PID ${props.event.pid}`
})

// Generate a consistent color from cgroup ID for workload badge
const workloadColor = computed(() => {
  if (!props.event.cgroupId || props.event.cgroupId === '0') return null
  // Hash the cgroup ID to get a hue value
  let hash = 0
  for (let i = 0; i < props.event.cgroupId.length; i++) {
    hash = props.event.cgroupId.charCodeAt(i) + ((hash << 5) - hash)
  }
  const hue = Math.abs(hash % 360)
  return `hsl(${hue}, 60%, 50%)`
})
</script>

<template>
  <div 
    class="event-row" 
    :class="[typeClass, { 'is-selected': isSelected }]"
    @click="$emit('select', event)"
  >
    <span class="event-time font-mono">{{ formatTime(event.timestamp) }}</span>
    
    <span class="event-type">
      <component :is="typeIcon" :size="14" />
    </span>
    
    <span class="event-process font-mono">{{ processName }}</span>
    
    <span class="event-details">{{ details }}</span>
    
    <span v-if="workloadColor" class="event-workload" :style="{ background: workloadColor }"></span>
  </div>
</template>

<style scoped>
.event-row {
  display: grid;
  grid-template-columns: 100px 32px 120px 1fr 28px;
  align-items: center;
  gap: 12px;
  padding: 0 16px;
  height: 40px;
  border-bottom: 1px solid var(--border-subtle);
  cursor: pointer;
  transition: background var(--transition-fast);
  box-sizing: border-box;
}

.event-row:hover {
  background: var(--bg-hover);
}

.event-row.is-selected {
  background: var(--bg-overlay);
}

.event-time {
  font-size: 11px;
  color: var(--text-muted);
}

.event-type {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: var(--radius-sm);
}

.type-exec .event-type {
  background: var(--status-info-dim);
  color: var(--status-info);
}

.type-connect .event-type {
  background: var(--status-warning-dim);
  color: var(--status-warning);
}

.type-file .event-type {
  background: var(--status-safe-dim);
  color: var(--status-safe);
}

.event-process {
  font-size: 12px;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.event-details {
  font-size: 12px;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.type-exec .event-details {
  font-family: var(--font-mono);
}

.type-file .event-details {
  color: var(--status-warning);
  font-family: var(--font-mono);
  font-size: 11px;
}

.event-workload {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin: auto;
}
</style>

