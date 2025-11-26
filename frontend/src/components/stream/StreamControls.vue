<script setup lang="ts">
import { Play, Pause, Trash2, Terminal, Globe, FileText } from 'lucide-vue-next'

interface Filters {
  exec: boolean
  connect: boolean
  file: boolean
  workloadId: string
}

defineProps<{
  isPaused: boolean
  eventCount: number
  filters: Filters
}>()

defineEmits<{
  togglePause: []
  clear: []
  'update:filters': [filters: Filters]
}>()
</script>

<template>
  <div class="stream-controls">
    <div class="controls-left">
      <!-- Play/Pause -->
      <button 
        class="control-btn primary"
        :class="{ paused: isPaused }"
        @click="$emit('togglePause')"
      >
        <Pause v-if="!isPaused" :size="16" />
        <Play v-else :size="16" />
        <span>{{ isPaused ? 'Resume' : 'Live' }}</span>
      </button>

      <!-- Clear -->
      <button class="control-btn" @click="$emit('clear')">
        <Trash2 :size="16" />
        <span>Clear</span>
      </button>

      <div class="control-divider"></div>

      <!-- Type Filters -->
      <div class="filter-group">
        <label class="filter-toggle" :class="{ active: filters.exec }">
          <input 
            type="checkbox" 
            :checked="filters.exec"
            @change="$emit('update:filters', { ...filters, exec: !filters.exec })"
          />
          <Terminal :size="14" />
          <span>Exec</span>
        </label>

        <label class="filter-toggle" :class="{ active: filters.connect }">
          <input 
            type="checkbox" 
            :checked="filters.connect"
            @change="$emit('update:filters', { ...filters, connect: !filters.connect })"
          />
          <Globe :size="14" />
          <span>Network</span>
        </label>

        <label class="filter-toggle" :class="{ active: filters.file }">
          <input 
            type="checkbox" 
            :checked="filters.file"
            @change="$emit('update:filters', { ...filters, file: !filters.file })"
          />
          <FileText :size="14" />
          <span>File</span>
        </label>
      </div>

    </div>

    <div class="controls-right">
      <div class="event-counter">
        <span class="counter-label">Events buffered:</span>
        <span class="counter-value font-mono">{{ eventCount.toLocaleString() }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.stream-controls {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--bg-elevated);
  border-bottom: 1px solid var(--border-subtle);
  gap: 16px;
  flex-wrap: wrap;
}

.controls-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.control-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  background: var(--bg-overlay);
  border-radius: var(--radius-md);
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  transition: all var(--transition-fast);
}

.control-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.control-btn.primary {
  background: var(--status-safe-dim);
  color: var(--status-safe);
}

.control-btn.primary:hover {
  background: var(--status-safe);
  color: var(--bg-void);
}

.control-btn.primary.paused {
  background: var(--status-warning-dim);
  color: var(--status-warning);
}

.control-btn.primary.paused:hover {
  background: var(--status-warning);
  color: var(--bg-void);
}

.control-divider {
  width: 1px;
  height: 24px;
  background: var(--border-subtle);
  margin: 0 4px;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 4px;
}

.filter-toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  border-radius: var(--radius-md);
  font-size: 11px;
  color: var(--text-muted);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.filter-toggle input {
  display: none;
}

.filter-toggle:hover {
  background: var(--bg-overlay);
  color: var(--text-secondary);
}

.filter-toggle.active {
  background: var(--bg-overlay);
  color: var(--text-primary);
}

.filter-toggle.active:first-of-type {
  color: var(--status-info);
}

.filter-group .filter-toggle:nth-child(2).active {
  color: var(--status-warning);
}

.filter-group .filter-toggle:nth-child(3).active {
  color: var(--status-safe);
}

.controls-right {
  display: flex;
  align-items: center;
}

.event-counter {
  display: flex;
  align-items: center;
  gap: 8px;
}

.counter-label {
  font-size: 12px;
  color: var(--text-muted);
}

.counter-value {
  font-size: 13px;
  color: var(--text-primary);
  font-weight: 600;
}
</style>

