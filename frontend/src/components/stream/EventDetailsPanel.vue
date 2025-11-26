<script setup lang="ts">
import { ref, watch } from 'vue'
import { X, Copy, Search, Terminal, Globe, FileText, Container } from 'lucide-vue-next'
import type { StreamEvent, ProcessInfo } from '../../lib/api'
import { getAncestors } from '../../lib/api'

type EventWithId = StreamEvent & { id: string }

const props = defineProps<{
  event: EventWithId | null
}>()

defineEmits<{
  close: []
  huntSimilar: [event: EventWithId]
}>()

const ancestors = ref<ProcessInfo[]>([])
const loadingAncestors = ref(false)

const formatTime = (timestamp: number) => {
  return new Date(timestamp).toLocaleString('en-US', {
    hour12: false,
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

const copyEventJson = () => {
  if (props.event) {
    navigator.clipboard.writeText(JSON.stringify(props.event, null, 2))
  }
}

watch(() => props.event, async (newEvent) => {
  if (newEvent) {
    loadingAncestors.value = true
    try {
      ancestors.value = await getAncestors(newEvent.pid)
    } catch (e) {
      ancestors.value = []
    } finally {
      loadingAncestors.value = false
    }
  } else {
    ancestors.value = []
  }
}, { immediate: true })
</script>

<template>
  <Transition name="slide">
    <aside v-if="event" class="details-panel">
      <div class="panel-header">
        <h3 class="panel-title">Event Details</h3>
        <button class="close-btn" @click="$emit('close')">
          <X :size="18" />
        </button>
      </div>

      <div class="panel-content">
        <!-- Event Type Badge -->
        <div class="event-badge" :class="`type-${event.type}`">
          <Terminal v-if="event.type === 'exec'" :size="16" />
          <Globe v-else-if="event.type === 'connect'" :size="16" />
          <FileText v-else :size="16" />
          <span>{{ event.type.toUpperCase() }} EVENT</span>
        </div>

        <!-- Basic Info -->
        <section class="detail-section">
          <h4 class="section-title">Process Information</h4>
          <div class="detail-grid">
            <div class="detail-row">
              <span class="detail-label">PID</span>
              <span class="detail-value font-mono">{{ event.pid }}</span>
            </div>
            <div v-if="event.type === 'exec'" class="detail-row">
              <span class="detail-label">PPID</span>
              <span class="detail-value font-mono">{{ event.ppid }}</span>
            </div>
            <div v-if="event.type === 'exec'" class="detail-row">
              <span class="detail-label">Command</span>
              <span class="detail-value font-mono">{{ event.comm }}</span>
            </div>
            <div v-if="event.type === 'exec'" class="detail-row">
              <span class="detail-label">Parent</span>
              <span class="detail-value font-mono">{{ event.parentComm }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Cgroup ID</span>
              <span class="detail-value font-mono">{{ event.cgroupId }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Container</span>
              <span class="detail-value">
                <span v-if="event.inContainer" class="container-yes">
                  <Container :size="12" /> Yes
                </span>
                <span v-else class="container-no">No</span>
              </span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Timestamp</span>
              <span class="detail-value font-mono">{{ formatTime(event.timestamp) }}</span>
            </div>
          </div>
        </section>

        <!-- Type-specific Info -->
        <section v-if="event.type === 'connect'" class="detail-section">
          <h4 class="section-title">Connection Details</h4>
          <div class="detail-grid">
            <div class="detail-row">
              <span class="detail-label">Address</span>
              <span class="detail-value font-mono">{{ event.addr }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Port</span>
              <span class="detail-value font-mono">{{ event.port }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Family</span>
              <span class="detail-value font-mono">{{ event.family === 2 ? 'IPv4' : 'IPv6' }}</span>
            </div>
          </div>
        </section>

        <section v-if="event.type === 'file'" class="detail-section">
          <h4 class="section-title">File Access Details</h4>
          <div class="detail-grid">
            <div class="detail-row">
              <span class="detail-label">Filename</span>
              <span class="detail-value font-mono filename">{{ event.filename }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Flags</span>
              <span class="detail-value font-mono">{{ event.flags }}</span>
            </div>
          </div>
        </section>

        <!-- Ancestry Chain -->
        <section class="detail-section">
          <h4 class="section-title">Ancestry Chain</h4>
          <div v-if="loadingAncestors" class="ancestry-loading">
            Loading...
          </div>
          <div v-else-if="ancestors.length === 0" class="ancestry-empty">
            No ancestry data available
          </div>
          <div v-else class="ancestry-chain">
            <div 
              v-for="(proc, index) in ancestors" 
              :key="proc.pid"
              class="ancestry-node"
              :class="{ 'is-current': index === ancestors.length - 1 }"
            >
              <span class="ancestry-indent">{{ '  '.repeat(index) }}{{ index > 0 ? '└─ ' : '' }}</span>
              <span class="ancestry-comm">{{ proc.comm }}</span>
              <span class="ancestry-pid">({{ proc.pid }})</span>
              <span v-if="index === ancestors.length - 1" class="ancestry-current">← current</span>
            </div>
          </div>
        </section>
      </div>

      <!-- Actions -->
      <div class="panel-actions">
        <button class="action-btn" @click="copyEventJson">
          <Copy :size="14" />
          Copy JSON
        </button>
        <button class="action-btn" @click="$emit('huntSimilar', event)">
          <Search :size="14" />
          Hunt Similar
        </button>
      </div>
    </aside>
  </Transition>
</template>

<style scoped>
.details-panel {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  width: 380px;
  background: var(--bg-surface);
  border-left: 1px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  z-index: 10;
}

.slide-enter-active,
.slide-leave-active {
  transition: transform 0.2s ease;
}

.slide-enter-from,
.slide-leave-to {
  transform: translateX(100%);
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.panel-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.close-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  transition: all var(--transition-fast);
}

.close-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.panel-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.event-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  border-radius: var(--radius-md);
  font-size: 11px;
  font-weight: 600;
  margin-bottom: 16px;
}

.event-badge.type-exec {
  background: var(--status-info-dim);
  color: var(--status-info);
}

.event-badge.type-connect {
  background: var(--status-warning-dim);
  color: var(--status-warning);
}

.event-badge.type-file {
  background: var(--status-safe-dim);
  color: var(--status-safe);
}

.detail-section {
  margin-bottom: 20px;
}

.section-title {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin: 0 0 12px 0;
}

.detail-grid {
  display: flex;
  flex-direction: column;
  gap: 8px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
  padding: 12px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.detail-label {
  font-size: 12px;
  color: var(--text-muted);
  flex-shrink: 0;
}

.detail-value {
  font-size: 12px;
  color: var(--text-primary);
  text-align: right;
  word-break: break-all;
}

.detail-value.filename {
  font-size: 11px;
  color: var(--status-warning);
}

.container-yes {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--status-info);
}

.container-no {
  color: var(--text-muted);
}

.ancestry-chain {
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
  padding: 12px;
  font-family: var(--font-mono);
  font-size: 12px;
}

.ancestry-node {
  line-height: 1.6;
  white-space: pre;
}

.ancestry-indent {
  color: var(--text-muted);
}

.ancestry-comm {
  color: var(--text-primary);
}

.ancestry-pid {
  color: var(--text-muted);
  margin-left: 4px;
}

.ancestry-current {
  color: var(--accent-primary);
  margin-left: 8px;
  font-size: 10px;
}

.ancestry-node.is-current .ancestry-comm {
  color: var(--accent-primary);
  font-weight: 600;
}

.ancestry-loading,
.ancestry-empty {
  font-size: 12px;
  color: var(--text-muted);
  text-align: center;
  padding: 16px;
}

.panel-actions {
  display: flex;
  gap: 8px;
  padding: 16px;
  border-top: 1px solid var(--border-subtle);
}

.action-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 16px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  transition: all var(--transition-fast);
}

.action-btn:hover {
  background: var(--bg-overlay);
  color: var(--text-primary);
}
</style>

