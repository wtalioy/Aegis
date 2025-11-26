<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { RecycleScroller } from 'vue-virtual-scroller'
import 'vue-virtual-scroller/dist/vue-virtual-scroller.css'
import { Radio } from 'lucide-vue-next'
import StreamControls from '../components/stream/StreamControls.vue'
import EventRow from '../components/stream/EventRow.vue'
import EventDetailsPanel from '../components/stream/EventDetailsPanel.vue'
import { subscribeToAllEvents, type StreamEvent } from '../lib/api'

const MAX_BUFFER_SIZE = 1000
const events = ref<(StreamEvent & { id: string })[]>([])
const isPaused = ref(false)
const selectedEvent = ref<(StreamEvent & { id: string }) | null>(null)
let eventIdCounter = 0

const filters = ref({
  exec: true,
  connect: true,
  file: true,
  workloadId: '' // Filter by specific workload (cgroup ID)
})

const filteredEvents = computed(() => {
  const filtered = events.value.filter(event => {
    if (!filters.value[event.type]) return false
    if (filters.value.workloadId && event.cgroupId !== filters.value.workloadId) return false
    return true
  })
  return [...filtered].sort((a, b) => b.timestamp - a.timestamp)
})

let unsubscribe: (() => void) | null = null

const handleEvent = (event: StreamEvent) => {
  if (isPaused.value) return

  const eventWithId = { ...event, id: `evt-${eventIdCounter++}` }
  events.value.push(eventWithId)
  
  if (events.value.length > MAX_BUFFER_SIZE) {
    events.value = events.value.slice(-MAX_BUFFER_SIZE)
  }
}

const togglePause = () => {
  isPaused.value = !isPaused.value
}

const clearEvents = () => {
  events.value = []
  selectedEvent.value = null
}

const selectEvent = (event: StreamEvent & { id: string }) => {
  selectedEvent.value = event
}

const closeDetails = () => {
  selectedEvent.value = null
}

const huntSimilar = (event: StreamEvent) => {
  if (event.type === 'exec') {
    console.log('Hunt similar exec:', event.comm)
  }
}

onMounted(() => {
  unsubscribe = subscribeToAllEvents(handleEvent)
})

onUnmounted(() => {
  unsubscribe?.()
})
</script>

<template>
  <div class="live-stream">
    <!-- Header -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">
          <Radio :size="24" class="title-icon" :class="{ pulsing: !isPaused }" />
          Live Stream
        </h1>
        <span class="page-subtitle">Real-time event monitoring</span>
      </div>
    </div>

    <!-- Main Content -->
    <div class="stream-container">
      <!-- Controls -->
      <StreamControls
        :is-paused="isPaused"
        :event-count="events.length"
        :filters="filters"
        @toggle-pause="togglePause"
        @clear="clearEvents"
        @update:filters="filters = $event"
      />

      <!-- Table Header -->
      <div class="table-header">
        <span class="col-time">Time</span>
        <span class="col-type">Type</span>
        <span class="col-process">Process</span>
        <span class="col-details">Details</span>
        <span class="col-workload"></span>
      </div>

      <!-- Event List with Virtual Scrolling -->
      <div class="table-wrapper" :class="{ 'has-panel': selectedEvent }">
        <RecycleScroller
          v-if="filteredEvents.length > 0"
          class="scroller"
          :items="filteredEvents"
          :item-size="40"
          key-field="id"
          v-slot="{ item }"
        >
          <EventRow
            :event="item"
            :is-selected="selectedEvent?.id === item.id"
            @select="selectEvent"
          />
        </RecycleScroller>

        <!-- Empty State -->
        <div v-else class="empty-state">
          <div class="empty-icon">📡</div>
          <div class="empty-title">
            {{ events.length === 0 ? 'Waiting for events...' : 'No matching events' }}
          </div>
          <div class="empty-description">
            {{ events.length === 0 
              ? 'Events will appear here as they are captured by eBPF probes'
              : 'Try adjusting your filters to see more events' 
            }}
          </div>
        </div>

        <!-- Details Panel -->
        <EventDetailsPanel
          :event="selectedEvent"
          @close="closeDetails"
          @hunt-similar="huntSimilar"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.live-stream {
  height: calc(100vh - var(--topbar-height) - var(--footer-height) - 48px);
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* Header */
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.header-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.title-icon {
  color: var(--text-muted);
}

.title-icon.pulsing {
  color: var(--status-safe);
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.page-subtitle {
  font-size: 14px;
  color: var(--text-muted);
}

/* Stream Container */
.stream-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: var(--bg-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  overflow: hidden;
  min-height: 0;
}

/* Table Header */
.table-header {
  display: grid;
  grid-template-columns: 100px 32px 120px 1fr 28px;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  background: var(--bg-void);
  border-bottom: 1px solid var(--border-subtle);
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

/* Table Wrapper */
.table-wrapper {
  flex: 1;
  position: relative;
  overflow: hidden;
  min-height: 0;
}

.table-wrapper.has-panel {
  margin-right: 380px;
}

.scroller {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  overflow: auto;
}

/* Empty State */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: 60px 40px;
  text-align: center;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
  opacity: 0.5;
}

.empty-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.empty-description {
  font-size: 14px;
  color: var(--text-muted);
  max-width: 300px;
}

/* Virtual scroller overrides */
:deep(.vue-recycle-scroller__item-wrapper) {
  overflow: visible;
}

:deep(.vue-recycle-scroller__item-view) {
  box-sizing: border-box;
}
</style>
