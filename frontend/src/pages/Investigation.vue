<!-- Investigation Page - Phase 4: AI-Assisted Threat Hunting -->
<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useInvestigation } from '../composables/useInvestigation'
import EventList from '../components/investigation/EventList.vue'
import AIContextPanel from '../components/investigation/AIContextPanel.vue'
import { Lightbulb, Search, X } from 'lucide-vue-next'
const { state, searchEvents, explainSelectedEvent, loading, loadMoreEvents, hasMore, loadingMore, refreshEvents, typeCounts } = useInvestigation()
const filterType = ref<string>('all')
const searchQuery = ref('')
const sortBy = ref<'time' | 'pid' | 'type' | 'process'>('time')
const sortDir = ref<'asc' | 'desc'>('desc')

// Auto-refresh events every 5 seconds
let refreshInterval: ReturnType<typeof setInterval> | null = null

const startAutoRefresh = () => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
  refreshInterval = setInterval(async () => {
    // Only refresh if not currently loading and no active search query
    if (!loading.value && !searchQuery.value.trim()) {
      await refreshEvents()
    }
  }, 5000) // Refresh every 5 seconds
}

const stopAutoRefresh = () => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
    refreshInterval = null
  }
}

const handleEventSelect = async (event: any) => {
  // Do not auto-analyze. Only set the selected event.
  state.value.selectedEvent = event
}

const filteredEvents = computed(() => {
  let result = state.value.events

  // Filter by type
  if (filterType.value !== 'all') {
    result = result.filter(e => e.type === filterType.value)
  }

  // Filter by search query
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(e =>
      (e.header?.comm || '').toLowerCase().includes(query) ||
      String(e.header?.pid || '').includes(query) ||
      (e.type || '').toLowerCase().includes(query)
    )
  }

  return result
})

// Use type counts from backend instead of calculating from loaded events
const eventTypeCounts = computed(() => typeCounts.value)

const sortedEvents = computed(() => {
  const arr = [...filteredEvents.value]
  const dir = sortDir.value === 'asc' ? 1 : -1
  const cmpStr = (a: string, b: string) => a.localeCompare(b) * dir
  const cmpNum = (a: number, b: number) => ((a ?? 0) - (b ?? 0)) * dir
  switch (sortBy.value) {
    case 'time':
      arr.sort((a, b) => cmpNum(a.header?.timestamp || 0, b.header?.timestamp || 0))
      break
    case 'pid':
      arr.sort((a, b) => cmpNum(a.header?.pid || 0, b.header?.pid || 0))
      break
    case 'type':
      arr.sort((a, b) => cmpStr(a.type || '', b.type || ''))
      break
    case 'process':
      arr.sort((a, b) => cmpStr(a.header?.comm || '', b.header?.comm || ''))
      break
  }
  return arr
})

const changeSort = (field: 'time' | 'pid' | 'type' | 'process') => {
  if (sortBy.value === field) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = field
    // default directions: newest first for time, largest first for pid
    sortDir.value = field === 'time' || field === 'pid' ? 'desc' : 'asc'
  }
}

onMounted(async () => {
  // Load initial events on page load
  await searchEvents({
    filter: {
      types: [],
      processes: [],
      pids: []
    },
    page: 1,
    limit: 50
  })
  // Start auto-refresh after initial load
  startAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<template>
  <div class="investigation-page">
    <!-- Header -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">
          <Lightbulb :size="24" class="title-icon" />
          Investigation
        </h1>
        <span class="page-subtitle">AI-assisted threat hunting and event analysis</span>
      </div>
      <div class="header-stats">
        <div class="stat-group">
          <div class="stat-badge exec" :class="{ active: filterType === 'exec' }"
            @click="filterType = filterType === 'exec' ? 'all' : 'exec'">
            <span class="stat-value">{{ eventTypeCounts.exec }}</span>
            <span class="stat-label">Exec</span>
          </div>
          <div class="stat-badge file" :class="{ active: filterType === 'file' }"
            @click="filterType = filterType === 'file' ? 'all' : 'file'">
            <span class="stat-value">{{ eventTypeCounts.file }}</span>
            <span class="stat-label">File</span>
          </div>
          <div class="stat-badge connect" :class="{ active: filterType === 'connect' }"
            @click="filterType = filterType === 'connect' ? 'all' : 'connect'">
            <span class="stat-value">{{ eventTypeCounts.connect }}</span>
            <span class="stat-label">Network</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <div class="investigation-content">
      <!-- Left Panel: Event Queue -->
      <div class="event-queue">
        <!-- Filters -->
        <div class="queue-filters">
          <div class="filter-search">
            <Search :size="16" class="search-icon" />
            <input v-model="searchQuery" type="text" placeholder="Search events..." class="search-input" />
            <button v-if="searchQuery" class="search-clear" @click="searchQuery = ''">
              <X :size="14" />
            </button>
          </div>
        </div>

        <!-- Events Display -->
        <div class="events-display">
          <div v-if="loading && sortedEvents.length === 0" class="loading-state">
            <div class="spinner"></div>
            <span>Loading events...</span>
          </div>
          <div v-else-if="sortedEvents.length === 0" class="empty-state">
            <span class="empty-icon">✓</span>
            <span class="empty-text">
              {{ state.events.length === 0 ? 'No events detected' : 'No matching events' }}
            </span>
            <span v-if="state.events.length > 0" class="empty-hint">
              Try adjusting your filters
            </span>
          </div>
          <EventList v-else :events="sortedEvents" :selected-event-id="state.selectedEvent?.id" :sort-by="sortBy"
            :sort-dir="sortDir" :has-more="hasMore" :loading-more="loadingMore" @select="handleEventSelect"
            @changeSort="changeSort" @loadMore="loadMoreEvents" />
        </div>
      </div>

      <!-- Right Panel: AI Context -->
      <div class="context-panel">
        <AIContextPanel :event="state.selectedEvent" :process-id="state.selectedEvent?.header?.pid"
          style="flex:1 1 0; min-height:0;" />

      </div>
    </div>
  </div>
</template>

<style scoped>
.investigation-page {
  height: calc(100vh - var(--topbar-height) - var(--footer-height) - 48px);
  display: flex;
  flex-direction: column;
  gap: 24px;
  padding: 0 4px;
}

/* Header */
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 20px;
  padding: 0 4px;
}

.header-content {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.page-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
  letter-spacing: -0.5px;
}

.title-icon {
  color: var(--accent-primary);
  filter: drop-shadow(0 0 8px var(--accent-glow));
}

.page-subtitle {
  font-size: 14px;
  color: var(--text-muted);
  font-weight: 400;
  letter-spacing: 0.2px;
}

.header-stats {
  display: flex;
  gap: 12px;
  align-items: center;
}

.stat-group {
  display: flex;
  gap: 10px;
  background: var(--bg-elevated);
  padding: 6px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  box-shadow: var(--shadow-sm);
}

.stat-badge {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 10px 18px;
  background: var(--bg-surface);
  border-radius: var(--radius-md);
  border: 1px solid transparent;
  min-width: 70px;
  cursor: pointer;
  transition: all var(--transition-normal);
  position: relative;
  overflow: hidden;
}

.stat-badge::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, transparent 0%, rgba(96, 165, 250, 0.05) 100%);
  opacity: 0;
  transition: opacity var(--transition-normal);
}

.stat-badge:hover {
  background: var(--bg-overlay);
  border-color: var(--border-default);
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
}

.stat-badge:hover::before {
  opacity: 1;
}

.stat-badge.active {
  border-color: var(--accent-primary);
  background: linear-gradient(135deg, var(--accent-glow) 0%, var(--bg-overlay) 100%);
  box-shadow: 0 0 0 1px var(--accent-primary), var(--shadow-md);
}

.stat-badge.active::before {
  opacity: 1;
}

.stat-value {
  font-size: 20px;
  font-weight: 700;
  font-family: var(--font-mono);
  line-height: 1.2;
  letter-spacing: -0.5px;
}

.stat-label {
  font-size: 10px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.8px;
  font-weight: 600;
  margin-top: 2px;
}

/* Stat badge colors */
.stat-badge.exec .stat-value {
  color: var(--chart-exec);
}

.stat-badge.file .stat-value {
  color: var(--chart-file);
}

.stat-badge.connect .stat-value {
  color: var(--chart-network);
}

.stat-badge.active.exec {
  box-shadow: 0 0 0 1px var(--chart-exec), 0 4px 12px rgba(96, 165, 250, 0.2);
}

.stat-badge.active.file {
  box-shadow: 0 0 0 1px var(--chart-file), 0 4px 12px rgba(16, 185, 129, 0.2);
}

.stat-badge.active.connect {
  box-shadow: 0 0 0 1px var(--chart-network), 0 4px 12px rgba(245, 158, 11, 0.2);
}

/* Main Content Layout */
.investigation-content {
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 420px;
  gap: 24px;
  min-height: 0;
}

/* Event Queue Panel */
.event-queue {
  display: flex;
  flex-direction: column;
  background: var(--bg-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  overflow: hidden;
  box-shadow: var(--shadow-md);
  backdrop-filter: blur(10px);
}

/* Filters */
.queue-filters {
  display: flex;
  flex-direction: column;
  gap: 0;
  padding: 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.filter-search {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
  height: 40px;
  box-sizing: border-box;
  transition: all var(--transition-normal);
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.1);
}

.filter-search:focus-within {
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 3px var(--accent-glow), inset 0 1px 2px rgba(0, 0, 0, 0.1);
  background: var(--bg-surface);
}

.search-icon {
  color: var(--text-muted);
  flex-shrink: 0;
  transition: color var(--transition-fast);
}

.filter-search:focus-within .search-icon {
  color: var(--accent-primary);
}

.search-input {
  flex: 1;
  background: transparent;
  border: none;
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
  font-weight: 400;
}

.search-input::placeholder {
  color: var(--text-muted);
}

.search-clear {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: var(--radius-sm);
  color: var(--text-muted);
  background: transparent;
  border: none;
  cursor: pointer;
  transition: all var(--transition-fast);
  opacity: 0.7;
}

.search-clear:hover {
  background: var(--bg-overlay);
  color: var(--text-primary);
  opacity: 1;
  transform: scale(1.1);
}

/* Events Display */
.events-display {
  flex: 1;
  overflow-y: auto;
  padding: 0;
  display: flex;
  flex-direction: column;
  background: var(--bg-surface);
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 64px 24px;
  gap: 16px;
  color: var(--text-muted);
}

.spinner {
  width: 28px;
  height: 28px;
  border: 3px solid var(--border-subtle);
  border-top-color: var(--accent-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 64px 24px;
  gap: 12px;
}

.empty-icon {
  font-size: 48px;
  color: var(--status-safe);
  opacity: 0.6;
  margin-bottom: 4px;
}

.empty-text {
  font-size: 15px;
  color: var(--text-secondary);
  text-align: center;
  font-weight: 500;
}

.empty-hint {
  font-size: 13px;
  color: var(--text-muted);
  text-align: center;
}

/* Context Panel */
.context-panel {
  display: flex;
  flex-direction: column;
  gap: 0;
  min-height: 0;
}

/* Responsive */
@media (max-width: 1100px) {
  .header-stats {
    flex-direction: column;
    gap: 12px;
  }

  .stat-group {
    width: 100%;
    justify-content: space-around;
  }
}

@media (max-width: 900px) {
  .investigation-content {
    grid-template-columns: 1fr;
    grid-template-rows: 1fr 1fr;
  }
}
</style>
