<!-- Event List - Redesigned for clear table layout -->
<script setup lang="ts">
import { FileText, Terminal, Globe } from 'lucide-vue-next'
import type { SecurityEvent } from '../../types/events'

const props = defineProps<{
  events: SecurityEvent[]
  selectedEventId?: string
  loading?: boolean
  sortBy?: 'time' | 'pid' | 'type' | 'process'
  sortDir?: 'asc' | 'desc'
  hasMore?: boolean
  loadingMore?: boolean
}>()

const emit = defineEmits<{
  select: [event: SecurityEvent]
  loadMore: []
  changeSort: [field: 'time' | 'pid' | 'type' | 'process']
}>()

const eventIcon = (type: string) => {
  switch (type) {
    case 'exec': return Terminal
    case 'file': return FileText
    case 'connect': return Globe
    default: return FileText
  }
}

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
  <div class="event-list">
    <div v-if="loading && events.length === 0" class="loading-state">Loading events...</div>
    <div v-else-if="events.length === 0" class="empty-state">No events found</div>

    <div v-else class="table">
      <!-- Sticky Header -->
      <div class="thead">
        <div class="th type sortable" role="button" tabindex="0" @click="$emit('changeSort', 'type')">
          Type
          <span v-if="sortBy === 'type'" class="sort-ind">{{ sortDir === 'asc' ? '▲' : '▼' }}</span>
        </div>
        <div class="th process sortable" role="button" tabindex="0" @click="$emit('changeSort', 'process')">
          Process
          <span v-if="sortBy === 'process'" class="sort-ind">{{ sortDir === 'asc' ? '▲' : '▼' }}</span>
        </div>
        <div class="th details">Details</div>
        <div class="th pid sortable" role="button" tabindex="0" @click="$emit('changeSort', 'pid')">
          PID
          <span v-if="sortBy === 'pid'" class="sort-ind">{{ sortDir === 'asc' ? '▲' : '▼' }}</span>
        </div>
        <div class="th time sortable" role="button" tabindex="0" @click="$emit('changeSort', 'time')">
          Time
          <span v-if="sortBy === 'time'" class="sort-ind">{{ sortDir === 'asc' ? '▲' : '▼' }}</span>
        </div>
      </div>

      <div class="tbody">
        <div v-for="event in events" :key="event.id" class="tr" :class="{ selected: event.id === selectedEventId }"
          @click="$emit('select', event)">
          <div class="td type">
            <component :is="eventIcon(event.type)" :size="16" class="icon" />
            <span class="badge" :data-type="event.type">{{ event.type }}</span>
          </div>
          <div class="td process" :title="event.header?.comm">{{ event.header?.comm || 'Unknown' }}</div>
          <div class="td details">
            <span v-if="event.type === 'exec'" class="details-text"
              :title="(event as any).commandLine || (event as any).filename || (event as any).header?.comm">
              {{ (event as any).commandLine || (event as any).filename || (event as any).header?.comm || '—' }}
            </span>
            <span v-else-if="event.type === 'file'" class="details-text" :title="(event as any).filename">
              {{ (event as any).filename || '—' }}
            </span>
            <span v-else-if="event.type === 'connect'" class="details-text">
              <template v-if="(event as any).addr && (event as any).port">
                {{ (event as any).addr }}:{{ (event as any).port }}
              </template>
              <template v-else-if="(event as any).addr">
                {{ (event as any).addr }}
              </template>
              <template v-else>—</template>
            </span>
            <span v-else class="details-text">—</span>
          </div>
          <div class="td pid">{{ event.header?.pid ?? '—' }}</div>
          <div class="td time">{{ event.header?.timestamp ? formatTime(event.header.timestamp) : 'Unknown' }}</div>
        </div>
      </div>

      <div v-if="hasMore" class="tfoot">
        <button class="load-more-btn" :disabled="loadingMore" @click="$emit('loadMore')">
          <span v-if="loadingMore">Loading…</span>
          <span v-else>Load More</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.event-list {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.loading-state,
.empty-state {
  padding: 48px 24px;
  text-align: center;
  color: var(--text-muted);
  font-size: 14px;
}

/* Table skeleton */
.table {
  display: grid;
  grid-template-rows: auto 1fr auto;
  min-height: 0;
}

.thead {
  display: grid;
  grid-template-columns: 160px 180px 1fr 120px 140px;
  gap: 0;
  position: sticky;
  top: 0;
  background: var(--bg-elevated);
  border-bottom: 1px solid var(--border-default);
  z-index: 10;
  backdrop-filter: blur(8px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.th {
  padding: 14px 16px;
  font-size: 11px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.8px;
  font-weight: 600;
  align-items: center;
  background: transparent;
}

.th.sortable {
  cursor: pointer;
  user-select: none;
  display: flex;
  gap: 8px;
  transition: all var(--transition-fast);
  border-radius: var(--radius-sm);
}

.th.sortable:hover {
  color: var(--text-primary);
  background: var(--bg-overlay);
}

.sort-ind {
  font-size: 10px;
  color: var(--accent-primary);
  font-weight: 700;
  margin-left: 2px;
}

.tbody {
  overflow-y: auto;
  background: var(--bg-surface);
}

.tr {
  position: relative;
  display: grid;
  grid-template-columns: 160px 180px 1fr 120px 140px;
  align-items: center;
  border-bottom: 1px solid var(--border-subtle);
  cursor: pointer;
  background: var(--bg-surface);
  transition: all var(--transition-normal);
}

.tr::after {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 0;
  background: var(--accent-primary);
  transition: width var(--transition-normal);
  opacity: 0;
}

/* Enhanced hover feedback */
.tr:hover {
  background: rgba(96, 165, 250, 0.08);
  box-shadow: inset 0 0 0 1px var(--border-default), 0 2px 4px rgba(0, 0, 0, 0.1);
  transform: translateX(2px);
}

.tr:hover::after {
  width: 3px;
  opacity: 1;
}

.tr.selected {
  background: rgba(96, 165, 250, 0.15);
  box-shadow: inset 0 0 0 1px var(--accent-primary), 0 2px 8px rgba(96, 165, 250, 0.15);
  border-bottom-color: var(--accent-primary);
}

.tr.selected::after {
  width: 3px;
  opacity: 1;
  background: var(--accent-primary);
  box-shadow: 0 0 8px var(--accent-glow);
}

.tr:nth-child(even) {
  background: var(--bg-elevated);
}

.tr:nth-child(even):hover {
  background: rgba(96, 165, 250, 0.08);
}

.tr:nth-child(even).selected {
  background: rgba(96, 165, 250, 0.15);
}

/* Keyboard focus parity */
.tr:focus-visible {
  outline: none;
  background: var(--bg-hover);
  box-shadow: inset 0 0 0 2px var(--border-focus);
}

.td {
  padding: 12px 16px;
  min-width: 0;
}

.type {
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
}

.icon {
  color: var(--text-muted);
  transition: color var(--transition-fast);
  flex-shrink: 0;
}

.tr:hover .icon {
  color: var(--accent-primary);
}

.badge {
  text-transform: uppercase;
  font-weight: 700;
  font-size: 10px;
  letter-spacing: 0.6px;
  padding: 3px 8px;
  border-radius: var(--radius-sm);
  background: rgba(96, 165, 250, 0.1);
  transition: all var(--transition-fast);
}

.tr:hover .badge {
  transform: scale(1.05);
}

.badge[data-type="exec"] {
  color: var(--chart-exec);
  background: rgba(96, 165, 250, 0.12);
}

.badge[data-type="file"] {
  color: var(--chart-file);
  background: rgba(16, 185, 129, 0.12);
}

.badge[data-type="connect"] {
  color: var(--chart-network);
  background: rgba(245, 158, 11, 0.12);
}

.process {
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 13px;
  transition: color var(--transition-fast);
}

.tr:hover .process {
  color: var(--accent-primary);
}

.details {
  font-weight: 500;
  color: var(--text-secondary);
  font-size: 13px;
  transition: color var(--transition-fast);
}

.tr:hover .details {
  color: var(--text-primary);
}

.details-text {
  display: block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}

.pid,
.time {
  font-family: var(--font-mono);
  color: var(--text-muted);
  font-size: 14px;
  transition: color var(--transition-fast);
}

.tr:hover .pid,
.tr:hover .time {
  color: var(--text-secondary);
}

.tfoot {
  display: flex;
  justify-content: center;
  padding: 16px;
  background: linear-gradient(to top, var(--bg-elevated) 0%, var(--bg-surface) 100%);
  border-top: 1px solid var(--border-subtle);
}

.load-more-btn {
  padding: 10px 20px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-normal);
  box-shadow: var(--shadow-sm);
}

.load-more-btn:hover {
  background: var(--bg-overlay);
  color: var(--text-primary);
  border-color: var(--accent-primary);
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
}

.load-more-btn:active {
  transform: translateY(0);
}

.load-more-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

@media (max-width: 900px) {

  .thead,
  .tr {
    grid-template-columns: 120px 150px 1fr 90px 120px;
  }

  .th,
  .td {
    padding: 10px 12px;
  }
}
</style>
