<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { AlertTriangle, Filter, Search, X } from 'lucide-vue-next'
import AlertCard from '../components/alerts/AlertCard.vue'
import AttackChain from '../components/topology/AttackChain.vue'
import { useAlerts } from '../composables/useAlerts'
import { getAncestors, type Alert, type ProcessInfo } from '../lib/api'

const route = useRoute()
const { alerts, getAlertsBySeverity } = useAlerts()

// State
const selectedAlert = ref<Alert | null>(null)
const ancestors = ref<ProcessInfo[]>([])
const loadingAncestors = ref(false)
const filterSeverity = ref<string>('all')
const filterWorkload = ref<string>('')
const searchQuery = ref('')

// Initialize workload filter from route query
onMounted(() => {
  if (route.query.workload) {
    filterWorkload.value = route.query.workload as string
  }
})

// Watch for route changes
watch(() => route.query.workload, (newWorkload) => {
  filterWorkload.value = (newWorkload as string) || ''
})

// Filtered alerts
const filteredAlerts = computed(() => {
  let result = alerts.value

  // Filter by workload (cgroup ID)
  if (filterWorkload.value) {
    result = result.filter(a => a.cgroupId === filterWorkload.value)
  }

  // Filter by severity
  if (filterSeverity.value !== 'all') {
    result = result.filter(a => a.severity === filterSeverity.value)
  }

  // Filter by search query
  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(a => 
      a.ruleName.toLowerCase().includes(query) ||
      a.processName.toLowerCase().includes(query) ||
      a.description.toLowerCase().includes(query) ||
      String(a.pid).includes(query)
    )
  }

  return result
})

const clearWorkloadFilter = () => {
  filterWorkload.value = ''
}

// Stats
const severityCounts = computed(() => getAlertsBySeverity())

// Select alert and load ancestors
const selectAlert = async (alert: Alert) => {
  selectedAlert.value = alert
  await loadAncestors(alert.pid)
}

const loadAncestors = async (pid: number) => {
  loadingAncestors.value = true
  try {
    ancestors.value = await getAncestors(pid)
  } catch (e) {
    console.error('Failed to load ancestors:', e)
    ancestors.value = []
  } finally {
    loadingAncestors.value = false
  }
}

const refreshAncestors = () => {
  if (selectedAlert.value) {
    loadAncestors(selectedAlert.value.pid)
  }
}

const clearSelection = () => {
  selectedAlert.value = null
  ancestors.value = []
}

// Clear selection when alerts change significantly
watch(() => alerts.value.length, (newLen, oldLen) => {
  if (selectedAlert.value && newLen < oldLen) {
    // Check if selected alert still exists
    const stillExists = alerts.value.some(a => a.id === selectedAlert.value?.id)
    if (!stillExists) {
      clearSelection()
    }
  }
})
</script>

<template>
  <div class="alerts-page">
    <!-- Header -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">
          <AlertTriangle :size="24" class="title-icon" />
          Threat Hunting
        </h1>
        <span class="page-subtitle">Alert analysis and attack chain visualization</span>
      </div>
      <div class="header-stats">
        <div class="stat-badge high">
          <span class="stat-value">{{ severityCounts.high }}</span>
          <span class="stat-label">High</span>
        </div>
        <div class="stat-badge warning">
          <span class="stat-value">{{ severityCounts.warning }}</span>
          <span class="stat-label">Warning</span>
        </div>
        <div class="stat-badge info">
          <span class="stat-value">{{ severityCounts.info }}</span>
          <span class="stat-label">Info</span>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <div class="alerts-content">
      <!-- Left Panel: Alert Queue -->
      <div class="alert-queue">
        <div class="queue-header">
          <h2 class="queue-title">Alert Queue</h2>
          <span class="queue-count">{{ filteredAlerts.length }}</span>
        </div>

        <!-- Filters -->
        <div class="queue-filters">
          <div class="filter-search">
            <Search :size="16" class="search-icon" />
            <input 
              v-model="searchQuery"
              type="text" 
              placeholder="Search alerts..."
              class="search-input"
            />
            <button 
              v-if="searchQuery" 
              class="search-clear" 
              @click="searchQuery = ''"
            >
              <X :size="14" />
            </button>
          </div>
          
          <div class="filter-row">
            <div class="filter-severity">
              <Filter :size="14" class="filter-icon" />
              <select v-model="filterSeverity" class="severity-select">
                <option value="all">All Severity</option>
                <option value="high">High</option>
                <option value="warning">Warning</option>
                <option value="info">Info</option>
              </select>
            </div>
            
            <div class="filter-workload">
              <input 
                v-model="filterWorkload"
                type="text" 
                placeholder="Cgroup ID..."
                class="workload-input"
              />
              <button 
                v-if="filterWorkload" 
                class="workload-clear" 
                @click="clearWorkloadFilter"
                title="Clear workload filter"
              >
                <X :size="12" />
              </button>
            </div>
          </div>
        </div>

        <!-- Alert List -->
        <div class="queue-list">
          <div v-if="filteredAlerts.length === 0" class="queue-empty">
            <span class="empty-icon">✓</span>
            <span class="empty-text">
              {{ alerts.length === 0 ? 'No alerts detected' : 'No matching alerts' }}
            </span>
          </div>
          <AlertCard
            v-for="alert in filteredAlerts"
            :key="alert.id"
            :alert="alert"
            :is-selected="selectedAlert?.id === alert.id"
            @select="selectAlert"
          />
        </div>
      </div>

      <!-- Right Panel: Attack Chain Visualization -->
      <div class="chain-panel">
        <AttackChain
          :ancestors="ancestors"
          :alert="selectedAlert"
          :loading="loadingAncestors"
          @refresh="refreshAncestors"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.alerts-page {
  height: calc(100vh - var(--topbar-height) - var(--footer-height) - 48px); /* viewport - topbar - footer - padding */
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* Header */
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 16px;
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
  color: var(--status-warning);
}

.page-subtitle {
  font-size: 14px;
  color: var(--text-muted);
}

.header-stats {
  display: flex;
  gap: 12px;
}

.stat-badge {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 16px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
  min-width: 70px;
}

.stat-value {
  font-size: 20px;
  font-weight: 700;
  font-family: var(--font-mono);
}

.stat-label {
  font-size: 11px;
  color: var(--text-muted);
  text-transform: uppercase;
}

.stat-badge.high .stat-value { color: var(--status-critical); }
.stat-badge.warning .stat-value { color: var(--status-warning); }
.stat-badge.info .stat-value { color: var(--status-info); }

/* Main Content Layout */
.alerts-content {
  flex: 1;
  display: grid;
  grid-template-columns: 320px 1fr;
  gap: 20px;
  min-height: 0;
}

/* Alert Queue Panel */
.alert-queue {
  display: flex;
  flex-direction: column;
  background: var(--bg-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  overflow: hidden;
}

.queue-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.queue-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.queue-count {
  padding: 2px 10px;
  background: var(--bg-overlay);
  border-radius: var(--radius-full);
  font-size: 12px;
  font-weight: 600;
  font-family: var(--font-mono);
  color: var(--text-secondary);
}

/* Filters */
.queue-filters {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  border-bottom: 1px solid var(--border-subtle);
  background: var(--bg-elevated);
}

.filter-search {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: var(--bg-surface);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
}

.filter-search:focus-within {
  border-color: var(--border-focus);
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
  width: 20px;
  height: 20px;
  border-radius: var(--radius-sm);
  color: var(--text-muted);
  transition: all var(--transition-fast);
}

.search-clear:hover {
  background: var(--bg-overlay);
  color: var(--text-primary);
}

.filter-row {
  display: flex;
  gap: 8px;
}

.filter-severity {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.filter-workload {
  display: flex;
  align-items: center;
  position: relative;
  flex: 1;
}

.workload-input {
  width: 100%;
  padding: 6px 28px 6px 10px;
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  font-size: 12px;
  font-family: var(--font-mono);
  color: var(--text-primary);
  outline: none;
  transition: border-color var(--transition-fast);
}

.workload-input::placeholder {
  color: var(--text-muted);
  font-family: var(--font-sans);
}

.workload-input:focus {
  border-color: var(--border-focus);
}

.workload-clear {
  position: absolute;
  right: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: var(--radius-sm);
  color: var(--text-muted);
  transition: all var(--transition-fast);
}

.workload-clear:hover {
  background: var(--bg-overlay);
  color: var(--text-primary);
}

.filter-icon {
  color: var(--text-muted);
}

.severity-select {
  flex: 1;
  padding: 6px 10px;
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-size: 12px;
  cursor: pointer;
}

.severity-select:focus {
  border-color: var(--border-focus);
  outline: none;
}

/* Alert List */
.queue-list {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}

.queue-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  gap: 8px;
}

.empty-icon {
  font-size: 32px;
  color: var(--status-safe);
}

.empty-text {
  font-size: 14px;
  color: var(--text-muted);
  text-align: center;
}

/* Chain Panel */
.chain-panel {
  min-height: 0;
}

/* Responsive */
@media (max-width: 900px) {
  .alerts-content {
    grid-template-columns: 1fr;
    grid-template-rows: 1fr 1fr;
  }
}
</style>
