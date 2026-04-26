<!-- Investigation Page - Phase 4: AI-Assisted Threat Hunting -->
<script setup lang="ts">
import { useInvestigationPage } from '../composables/useInvestigationPage'
import EventList from '../components/investigation/EventList.vue'
import AIContextPanel from '../components/investigation/AIContextPanel.vue'
import { Search, X, CheckCircle2 } from 'lucide-vue-next'

const {
  state,
  loading,
  loadMoreEvents,
  hasMore,
  loadingMore,
  filterType,
  searchQuery,
  sortBy,
  sortDir,
  eventTypeCounts,
  sortedEvents,
  handleEventSelect,
  changeSort
} = useInvestigationPage()
</script>

<template>
  <div class="investigation-page">
    <!-- Header -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">
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
            <CheckCircle2 :size="40" class="empty-icon" />
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
        <AIContextPanel :event="state.selectedEvent" :process-id="state.selectedEvent?.pid"
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
  padding: 24px;
}

/* Header */
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 20px;
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
  font-size: 24px;
  /* Softened */
  font-weight: 600;
  /* Softened */
  color: var(--text-primary);
  margin: 0;
}

.title-icon {
  color: var(--accent-primary);
}

.page-subtitle {
  font-size: 14px;
  color: var(--text-muted);
  font-weight: 400;
}

.header-stats {
  display: flex;
  gap: 12px;
  align-items: center;
}

.stat-group {
  display: flex;
  gap: 8px;
  background: var(--bg-overlay);
  padding: 6px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
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
  transition: all var(--transition-fast);
}

.stat-badge:hover {
  background: var(--bg-hover);
  border-color: var(--border-default);
}

.stat-badge.active {
  border-color: var(--accent-primary);
  background: var(--bg-hover);
}

.stat-value {
  font-size: 20px;
  font-weight: 600;
  /* Softened */
  font-family: var(--font-mono);
  line-height: 1.2;
}

.stat-label {
  font-size: 10px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-weight: 500;
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
}

/* Filters */
.queue-filters {
  padding: 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.filter-search {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 14px;
  background: var(--bg-overlay);
  border-radius: var(--radius-md);
  border: 1px solid transparent;
  height: 40px;
  box-sizing: border-box;
  transition: all var(--transition-fast);
}

.filter-search:focus-within {
  border-color: var(--accent-primary);
  background: var(--bg-surface);
  box-shadow: 0 0 0 3px var(--accent-glow);
}

.search-icon {
  color: var(--text-muted);
  flex-shrink: 0;
}

.search-input {
  flex: 1;
  background: transparent;
  border: none;
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
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
}

.search-clear:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

/* Events Display */
.events-display {
  flex: 1;
  overflow-y: auto;
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
  color: var(--text-muted);
  opacity: 0.6;
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
