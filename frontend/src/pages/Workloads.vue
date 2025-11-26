<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { 
  Boxes, RefreshCw, ArrowUpDown, Activity, FileText, Network, 
  AlertTriangle, ChevronRight, Clock, Play, Bell
} from 'lucide-vue-next'
import Card from '../components/common/Card.vue'
import { getWorkloads, type Workload } from '../lib/api'

const router = useRouter()
const workloads = ref<Workload[]>([])
const loading = ref(false)
const sortKey = ref<keyof Workload>('lastSeen')
const sortDesc = ref(true)
const expandedId = ref<string | null>(null)

const fetchWorkloads = async () => {
  loading.value = true
  try {
    workloads.value = await getWorkloads()
  } catch (e) {
    console.error('Failed to fetch workloads:', e)
  } finally {
    loading.value = false
  }
}

const sortedWorkloads = computed(() => {
  return [...workloads.value].sort((a, b) => {
    const aVal = a[sortKey.value]
    const bVal = b[sortKey.value]
    if (typeof aVal === 'number' && typeof bVal === 'number') {
      return sortDesc.value ? bVal - aVal : aVal - bVal
    }
    return sortDesc.value 
      ? String(bVal).localeCompare(String(aVal))
      : String(aVal).localeCompare(String(bVal))
  })
})

const toggleSort = (key: keyof Workload) => {
  if (sortKey.value === key) {
    sortDesc.value = !sortDesc.value
  } else {
    sortKey.value = key
    sortDesc.value = true
  }
}

const toggleExpand = (id: string) => {
  expandedId.value = expandedId.value === id ? null : id
}

const formatTime = (timestamp: number) => {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString('en-US', {
    hour12: false,
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatRelativeTime = (timestamp: number) => {
  if (!timestamp) return '-'
  const now = Date.now()
  const diff = now - timestamp
  if (diff < 60000) return 'Just now'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`
  return `${Math.floor(diff / 86400000)}d ago`
}

const shortenPath = (path: string) => {
  if (!path) return '-'
  if (path.length <= 45) return path
  return '…' + path.slice(-45)
}

// Generate consistent color from cgroup ID
const workloadColor = (id: string) => {
  if (!id || id === '0') return 'var(--text-muted)'
  let hash = 0
  for (let i = 0; i < id.length; i++) {
    hash = id.charCodeAt(i) + ((hash << 5) - hash)
  }
  const hue = Math.abs(hash % 360)
  return `hsl(${hue}, 60%, 50%)`
}

const totalEvents = computed(() => {
  return workloads.value.reduce((sum, w) => 
    sum + w.execCount + w.fileCount + w.connectCount, 0)
})

const totalAlerts = computed(() => {
  return workloads.value.reduce((sum, w) => sum + w.alertCount, 0)
})

const navigateToStream = (workloadId: string) => {
  router.push({ path: '/stream', query: { workload: workloadId } })
}

const navigateToAlerts = (workloadId: string) => {
  router.push({ path: '/alerts', query: { workload: workloadId } })
}

onMounted(() => {
  fetchWorkloads()
  setInterval(fetchWorkloads, 5000)
})
</script>

<template>
  <div class="workloads-page">
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">
          <Boxes :size="24" class="title-icon" />
          Workloads
        </h1>
        <span class="page-subtitle">Active cgroup-based workload groups</span>
      </div>
      <button class="refresh-btn" @click="fetchWorkloads" :disabled="loading">
        <RefreshCw :size="16" :class="{ spinning: loading }" />
        Refresh
      </button>
    </div>

    <!-- Stats Row -->
    <div class="stats-row">
      <div class="stat-item">
        <span class="stat-value">{{ workloads.length }}</span>
        <span class="stat-label">Workloads</span>
      </div>
      <div class="stat-item">
        <span class="stat-value">{{ totalEvents }}</span>
        <span class="stat-label">Total Events</span>
      </div>
      <div class="stat-item" :class="{ 'has-alerts': totalAlerts > 0 }">
        <span class="stat-value">{{ totalAlerts }}</span>
        <span class="stat-label">Alerts</span>
      </div>
    </div>

    <!-- Workloads Table -->
    <Card class="table-card">
      <div class="table-container">
        <table class="workloads-table">
          <thead>
            <tr>
              <th class="col-expand"></th>
              <th class="col-workload" @click="toggleSort('id')">
                Workload
                <ArrowUpDown :size="12" class="sort-icon" />
              </th>
              <th class="col-activity" @click="toggleSort('execCount')">
                Activity
                <ArrowUpDown :size="12" class="sort-icon" />
              </th>
              <th class="col-alerts" @click="toggleSort('alertCount')">
                Status
                <ArrowUpDown :size="12" class="sort-icon" />
              </th>
              <th class="col-last" @click="toggleSort('lastSeen')">
                Last Active
                <ArrowUpDown :size="12" class="sort-icon" />
              </th>
            </tr>
          </thead>
          <tbody>
            <template v-for="w in sortedWorkloads" :key="w.id">
              <tr 
                class="workload-row" 
                :class="{ expanded: expandedId === w.id, 'has-alerts': w.alertCount > 0 }"
                @click="toggleExpand(w.id)"
              >
                <td class="col-expand">
                  <ChevronRight 
                    :size="16" 
                    class="expand-icon" 
                    :class="{ rotated: expandedId === w.id }" 
                  />
                </td>
                <td class="col-workload">
                  <div class="workload-info">
                    <div class="workload-header">
                      <span class="workload-dot" :style="{ background: workloadColor(w.id) }"></span>
                      <code class="workload-id">{{ w.id }}</code>
                    </div>
                    <span class="workload-path" :title="w.cgroupPath">{{ shortenPath(w.cgroupPath) }}</span>
                  </div>
                </td>
                <td class="col-activity">
                  <div class="activity-summary">
                    <span class="activity-item" :class="{ active: w.execCount > 0 }">
                      <Play :size="12" />
                      {{ w.execCount }}
                    </span>
                    <span class="activity-item" :class="{ active: w.fileCount > 0 }">
                      <FileText :size="12" />
                      {{ w.fileCount }}
                    </span>
                    <span class="activity-item" :class="{ active: w.connectCount > 0 }">
                      <Network :size="12" />
                      {{ w.connectCount }}
                    </span>
                  </div>
                </td>
                <td class="col-alerts">
                  <div v-if="w.alertCount > 0" class="alert-indicator">
                    <AlertTriangle :size="14" />
                    <span>{{ w.alertCount }}</span>
                  </div>
                  <span v-else class="status-ok">OK</span>
                </td>
                <td class="col-last">
                  <span class="last-seen">{{ formatRelativeTime(w.lastSeen) }}</span>
                </td>
              </tr>
              
              <!-- Expanded Details Row -->
              <tr v-if="expandedId === w.id" class="details-row">
                <td colspan="5">
                  <div class="details-panel">
                    <div class="details-grid">
                      <div class="detail-section">
                        <h4>Activity Breakdown</h4>
                        <div class="detail-stats">
                          <div class="detail-stat">
                            <Play :size="16" class="stat-icon exec" />
                            <div class="stat-content">
                              <span class="stat-num">{{ w.execCount }}</span>
                              <span class="stat-name">Process Executions</span>
                            </div>
                          </div>
                          <div class="detail-stat">
                            <FileText :size="16" class="stat-icon file" />
                            <div class="stat-content">
                              <span class="stat-num">{{ w.fileCount }}</span>
                              <span class="stat-name">File Operations</span>
                            </div>
                          </div>
                          <div class="detail-stat">
                            <Network :size="16" class="stat-icon net" />
                            <div class="stat-content">
                              <span class="stat-num">{{ w.connectCount }}</span>
                              <span class="stat-name">Network Connections</span>
                            </div>
                          </div>
                          <div class="detail-stat" :class="{ 'has-alerts': w.alertCount > 0 }">
                            <Bell :size="16" class="stat-icon alert" />
                            <div class="stat-content">
                              <span class="stat-num">{{ w.alertCount }}</span>
                              <span class="stat-name">Security Alerts</span>
                            </div>
                          </div>
                        </div>
                      </div>
                      
                      <div class="detail-section">
                        <h4>Timeline</h4>
                        <div class="timeline-info">
                          <div class="timeline-item">
                            <Clock :size="14" />
                            <span class="timeline-label">First seen:</span>
                            <span class="timeline-value">{{ formatTime(w.firstSeen) }}</span>
                          </div>
                          <div class="timeline-item">
                            <Activity :size="14" />
                            <span class="timeline-label">Last active:</span>
                            <span class="timeline-value">{{ formatTime(w.lastSeen) }}</span>
                          </div>
                        </div>
                      </div>
                    </div>
                    
                    <div class="details-actions">
                      <button class="action-btn primary" @click.stop="navigateToStream(w.id)">
                        <Activity :size="14" />
                        View Events
                      </button>
                      <button 
                        v-if="w.alertCount > 0" 
                        class="action-btn alert" 
                        @click.stop="navigateToAlerts(w.id)"
                      >
                        <AlertTriangle :size="14" />
                        View Alerts
                      </button>
                    </div>
                  </div>
                </td>
              </tr>
            </template>
          </tbody>
        </table>

        <div v-if="workloads.length === 0 && !loading" class="empty-state">
          <Boxes :size="48" class="empty-icon" />
          <span class="empty-title">No workloads detected</span>
          <span class="empty-desc">Workloads will appear here as events are captured</span>
        </div>
      </div>
    </Card>
  </div>
</template>

<style scoped>
.workloads-page {
  max-width: 1400px;
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 24px;
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
  color: var(--accent-primary);
}

.page-subtitle {
  font-size: 14px;
  color: var(--text-muted);
}

.refresh-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: var(--bg-overlay);
  border-radius: var(--radius-md);
  font-size: 13px;
  color: var(--text-secondary);
  transition: all var(--transition-fast);
}

.refresh-btn:hover:not(:disabled) {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.refresh-btn .spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* Stats Row */
.stats-row {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
}

.stat-item {
  flex: 1;
  max-width: 180px;
  display: flex;
  flex-direction: column;
  padding: 16px 24px;
  background: var(--bg-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  text-align: center;
}

.stat-item.has-alerts {
  border-color: var(--status-critical);
  background: color-mix(in srgb, var(--status-critical) 5%, var(--bg-surface));
}

.stat-item.has-alerts .stat-value {
  color: var(--status-critical);
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  font-family: var(--font-mono);
  color: var(--text-primary);
}

.stat-label {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 4px;
}

/* Table */
.table-card :deep(.card-content) {
  padding: 0;
}

.table-container {
  overflow-x: auto;
}

.workloads-table {
  width: 100%;
  border-collapse: collapse;
}

.workloads-table th {
  padding: 12px 16px;
  text-align: left;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
  background: var(--bg-void);
  border-bottom: 1px solid var(--border-subtle);
  cursor: pointer;
  user-select: none;
  white-space: nowrap;
}

.workloads-table th:hover {
  color: var(--text-secondary);
}

.col-expand {
  width: 32px;
  cursor: default !important;
}

.sort-icon {
  opacity: 0.5;
  margin-left: 4px;
  vertical-align: middle;
}

.workload-row {
  border-bottom: 1px solid var(--border-subtle);
  transition: background var(--transition-fast);
  cursor: pointer;
}

.workload-row:hover {
  background: var(--bg-hover);
}

.workload-row.expanded {
  background: var(--bg-overlay);
}

.workload-row.has-alerts {
  border-left: 3px solid var(--status-critical);
}

.workloads-table td {
  padding: 14px 16px;
  font-size: 13px;
  color: var(--text-secondary);
}

.expand-icon {
  color: var(--text-muted);
  transition: transform var(--transition-fast);
}

.expand-icon.rotated {
  transform: rotate(90deg);
}

.workload-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.workload-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.workload-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.workload-id {
  font-family: var(--font-mono);
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.workload-path {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-muted);
  padding-left: 18px;
}

/* Activity Summary */
.activity-summary {
  display: flex;
  gap: 12px;
}

.activity-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--text-muted);
  padding: 4px 8px;
  background: var(--bg-void);
  border-radius: var(--radius-sm);
}

.activity-item.active {
  color: var(--text-secondary);
  background: var(--bg-overlay);
}

.activity-item svg {
  opacity: 0.7;
}

/* Alert Indicator */
.alert-indicator {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--status-critical);
  font-size: 13px;
  font-weight: 600;
}

.alert-indicator svg {
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.status-ok {
  color: var(--status-ok);
  font-size: 12px;
  font-weight: 500;
}

.last-seen {
  font-size: 12px;
  color: var(--text-muted);
}

/* Details Row */
.details-row td {
  padding: 0 !important;
  background: var(--bg-void);
}

.details-panel {
  padding: 20px 24px;
  border-top: 1px solid var(--border-subtle);
}

.details-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 32px;
  margin-bottom: 20px;
}

.detail-section h4 {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
  margin: 0 0 12px 0;
}

.detail-stats {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.detail-stat {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: var(--bg-surface);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
}

.detail-stat.has-alerts {
  border-color: var(--status-critical);
  background: color-mix(in srgb, var(--status-critical) 8%, var(--bg-surface));
}

.stat-icon {
  color: var(--text-muted);
}

.stat-icon.exec { color: var(--accent-primary); }
.stat-icon.file { color: var(--status-warning); }
.stat-icon.net { color: var(--status-info); }
.stat-icon.alert { color: var(--status-critical); }

.stat-content {
  display: flex;
  flex-direction: column;
}

.stat-num {
  font-family: var(--font-mono);
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
}

.stat-name {
  font-size: 11px;
  color: var(--text-muted);
}

.timeline-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.timeline-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-muted);
}

.timeline-label {
  color: var(--text-muted);
}

.timeline-value {
  font-family: var(--font-mono);
  color: var(--text-secondary);
}

.details-actions {
  display: flex;
  gap: 12px;
  padding-top: 16px;
  border-top: 1px solid var(--border-subtle);
}

.action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px 20px;
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 500;
  border: 1px solid transparent;
  cursor: pointer;
  transition: all var(--transition-fast);
  min-width: 140px;
  height: 40px;
}

.action-btn.primary {
  background: var(--accent-primary);
  color: #fff;
  border-color: var(--accent-primary);
}

.action-btn.primary:hover {
  background: var(--accent-hover);
  border-color: var(--accent-hover);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px color-mix(in srgb, var(--accent-primary) 30%, transparent);
}

.action-btn.secondary {
  background: var(--bg-surface);
  color: var(--text-secondary);
  border-color: var(--border-subtle);
}

.action-btn.secondary:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
  border-color: var(--border-default);
}

.action-btn.alert {
  background: color-mix(in srgb, var(--status-critical) 12%, var(--bg-surface));
  color: var(--status-critical);
  border-color: color-mix(in srgb, var(--status-critical) 40%, transparent);
}

.action-btn.alert:hover {
  background: color-mix(in srgb, var(--status-critical) 20%, var(--bg-surface));
  border-color: var(--status-critical);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px color-mix(in srgb, var(--status-critical) 20%, transparent);
}

/* Empty State */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 64px 24px;
  text-align: center;
}

.empty-icon {
  color: var(--text-muted);
  opacity: 0.5;
  margin-bottom: 16px;
}

.empty-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.empty-desc {
  font-size: 14px;
  color: var(--text-muted);
}
</style>
