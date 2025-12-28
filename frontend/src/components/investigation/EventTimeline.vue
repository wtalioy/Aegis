<!-- Event Timeline Component - Phase 4 -->
<script setup lang="ts">
import { computed } from 'vue'
import { Clock, Circle } from 'lucide-vue-next'
import type { SecurityEvent } from '../../types/events'

const props = defineProps<{
  events: SecurityEvent[]
  selectedEventId?: string
}>()

const emit = defineEmits<{
  select: [event: SecurityEvent]
}>()

const sortedEvents = computed(() => {
  return [...props.events].sort((a, b) =>
    (b.header?.timestamp || 0) - (a.header?.timestamp || 0)
  )
})

const formatTime = (timestamp: number) => {
  return new Date(timestamp).toLocaleTimeString('en-US', {
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

const eventTypeColor = (type: string) => {
  switch (type) {
    case 'exec': return 'var(--chart-exec)'
    case 'file': return 'var(--chart-file)'
    case 'connect': return 'var(--chart-network)'
    default: return 'var(--text-muted)'
  }
}
</script>

<template>
  <div class="event-timeline">
    <div class="timeline-header">
      <Clock :size="18" />
      <h3>Event Timeline</h3>
      <span class="event-count">{{ sortedEvents.length }} events</span>
    </div>

    <div class="timeline-container">
      <div v-for="event in sortedEvents" :key="event.id" class="timeline-item"
        :class="{ selected: event.id === selectedEventId }" @click="$emit('select', event)">
        <div class="marker" :style="{ background: eventTypeColor(event.type), boxShadow: `0 0 0 2px var(--bg-surface)` }"></div>
        <div class="timeline-content">
          <div class="top-row">
            <div class="process" :title="event.header?.comm">{{ event.header?.comm || 'Unknown' }}</div>
            <div class="time">{{ event.header?.timestamp ? formatTime(event.header.timestamp) : 'Unknown' }}</div>
          </div>
          <div class="bottom-row">
            <span class="type-badge" :data-type="event.type">{{ event.type }}</span>
            <span v-if="event.header?.pid" class="pid">PID: {{ event.header.pid }}</span>
          </div>
        </div>
      </div>

      <div v-if="sortedEvents.length === 0" class="empty-state">
        No events in timeline
      </div>
    </div>
  </div>
</template>

<style scoped>
.event-timeline {
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* Header hidden in this view (kept for future use) */
.timeline-header { display: none; }

.event-count { font-size: 13px; color: var(--text-muted); }

/* Container gets a single left rail so items don't draw their own */
.timeline-container {
  position: relative;
  flex: 1;
  overflow-y: auto;
  padding: 4px 0 4px 28px; /* space for the rail and marker */
}
.timeline-container::before {
  content: '';
  position: absolute;
  left: 12px; /* center of marker column */
  top: 0;
  bottom: 0;
  width: 2px;
  background: var(--border-subtle);
}

/* Items: marker + content in two columns */
.timeline-item {
  display: grid;
  grid-template-columns: 16px 1fr;
  column-gap: 12px;
  align-items: start;
  padding: 10px 8px;
  border-radius: var(--radius-md);
  cursor: pointer;
}
.timeline-item:hover { background: var(--bg-hover); }
.timeline-item.selected { background: rgba(59,130,246,.10); }

.marker {
  width: 10px; height: 10px; border-radius: 50%;
  box-shadow: 0 0 0 2px var(--bg-surface);
  align-self: center;
}

.timeline-content { min-width: 0; display: flex; flex-direction: column; gap: 6px; }
.top-row { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.process { font-weight: 600; color: var(--text-primary); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.time { font-family: var(--font-mono); color: var(--text-secondary); white-space: nowrap; }

.bottom-row { display: flex; align-items: center; gap: 12px; color: var(--text-secondary); }
.type-badge { text-transform: uppercase; font-weight: 700; font-size: 11px; letter-spacing: .5px; }
.type-badge[data-type="exec"] { color: var(--chart-exec); }
.type-badge[data-type="file"] { color: var(--chart-file); }
.type-badge[data-type="connect"] { color: var(--chart-network); }
.pid { color: var(--text-muted); white-space: nowrap; }

.empty-state { padding: 40px; text-align: center; color: var(--text-muted); font-size: 14px; }
</style>
