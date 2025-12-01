<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import {
  Boxes, RefreshCw, ArrowUpDown, Activity, FileText, Network,
  ChevronRight, Clock, Play, ShieldOff, AlertTriangle, ShieldCheck, Shield,
  Ban, Eye, Radar
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

// Security status for workload
const getSecurityStatus = (w: Workload) => {
  if (w.blockedCount > 0) return 'blocked'
  if (w.alertCount > 0) return 'alert'
  return 'ok'
}

const totalBlocked = computed(() => {
  return workloads.value.reduce((sum, w) => sum + w.blockedCount, 0)
})

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
      <div class="stat-card">
        <div class="stat-icon-wrap workloads">
          <Boxes :size="20" />
        </div>
        <div class="stat-content">
          <span class="stat-value">{{ workloads.length }}</span>
          <span class="stat-label">Workloads</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon-wrap events">
          <Activity :size="20" />
        </div>
        <div class="stat-content">
          <span class="stat-value">{{ totalEvents }}</span>
          <span class="stat-label">Total Events</span>
        </div>
      </div>
      <div class="stat-card" :class="{ blocked: totalBlocked > 0, alert: totalBlocked === 0 && totalAlerts > 0 }">
        <div class="stat-icon-wrap" :class="totalBlocked > 0 ? 'blocked' : (totalAlerts > 0 ? 'alert' : 'safe')">
          <Ban v-if="totalBlocked > 0" :size="20" />
          <AlertTriangle v-else-if="totalAlerts > 0" :size="20" />
          <ShieldCheck v-else :size="20" />
        </div>
        <div class="stat-content">
          <span class="stat-value">{{ totalBlocked > 0 ? totalBlocked : totalAlerts }}</span>
          <span class="stat-label">{{ totalBlocked > 0 ? 'Threats Blocked' : 'Security Events' }}</span>
        </div>
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
              <th class="col-status" @click="toggleSort('alertCount')">
                Security Status
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
              <tr class="workload-row" :class="[
                { expanded: expandedId === w.id },
                `status-${getSecurityStatus(w)}`
              ]" @click="toggleExpand(w.id)">
                <td class="col-expand">
                  <ChevronRight :size="16" class="expand-icon" :class="{ rotated: expandedId === w.id }" />
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
                <td class="col-status">
                  <div class="status-badges">
                    <div v-if="w.blockedCount > 0" class="status-badge blocked">
                      <Ban :size="14" />
                      <span>{{ w.blockedCount }} Blocked</span>
                    </div>
                    <div v-if="w.alertCount - w.blockedCount > 0" class="status-badge alert">
                      <AlertTriangle :size="14" />
                      <span>{{ w.alertCount - w.blockedCount }} Alert{{ w.alertCount - w.blockedCount > 1 ? 's' : ''
                        }}</span>
                    </div>
                    <div v-if="w.alertCount === 0" class="status-badge ok">
                      <ShieldCheck :size="14" />
                      <span>Secure</span>
                    </div>
                  </div>
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
                            <div class="stat-info">
                              <span class="stat-num">{{ w.execCount }}</span>
                              <span class="stat-name">Process Executions</span>
                            </div>
                          </div>
                          <div class="detail-stat">
                            <FileText :size="16" class="stat-icon file" />
                            <div class="stat-info">
                              <span class="stat-num">{{ w.fileCount }}</span>
                              <span class="stat-name">File Operations</span>
                            </div>
                          </div>
                          <div class="detail-stat">
                            <Network :size="16" class="stat-icon net" />
                            <div class="stat-info">
                              <span class="stat-num">{{ w.connectCount }}</span>
                              <span class="stat-name">Network Connections</span>
                            </div>
                          </div>
                        </div>
                      </div>

                      <div class="detail-section">
                        <h4>Security & Timeline</h4>
                        <div class="security-timeline">
                          <div v-if="w.blockedCount > 0" class="security-status blocked">
                            <Ban :size="18" />
                            <span class="security-label">{{ w.blockedCount }} Threat{{ w.blockedCount > 1 ? 's' : '' }}
                              Blocked</span>
                          </div>
                          <div v-if="w.alertCount - w.blockedCount > 0" class="security-status alert">
                            <AlertTriangle :size="18" />
                            <span class="security-label">{{ w.alertCount - w.blockedCount }} Alert{{ w.alertCount -
                              w.blockedCount > 1 ? 's' : '' }}</span>
                          </div>
                          <div v-if="w.alertCount === 0" class="security-status ok">
                            <ShieldCheck :size="18" />
                            <span class="security-label">No Security Issues</span>
                          </div>
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
                    </div>

                    <div class="details-actions">
                      <button v-if="w.alertCount > 0" class="action-btn secondary" @click.stop="navigateToAlerts(w.id)">
                        <Radar :size="14" />
                        Security Events
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
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-header {
  display: flex;
  align-items: center;
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
  color: var(--accent-primary);
}

.page-subtitle {
  font-size: 13px;
  color: var(--text-muted);
}

.refresh-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 18px;
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  font-size: 13px;
  color: var(--text-secondary);
  transition: all var(--transition-fast);
}

.refresh-btn:hover:not(:disabled) {
  background: var(--bg-overlay);
  color: var(--text-primary);
  border-color: var(--border-default);
}

.refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.refresh-btn .spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}

/* Stats Row */
.stats-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 18px 20px;
  background: var(--bg-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
}

.stat-card.alert {
  border-color: rgba(251, 191, 36, 0.4);
  background: linear-gradient(135deg, rgba(251, 191, 36, 0.05), var(--bg-surface) 50%);
}

.stat-card.blocked {
  border-color: rgba(239, 68, 68, 0.5);
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.06), var(--bg-surface) 50%);
}

.stat-icon-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md);
}

.stat-icon-wrap.workloads {
  background: rgba(99, 102, 241, 0.15);
  color: #6366f1;
}

.stat-icon-wrap.events {
  background: rgba(96, 165, 250, 0.15);
  color: var(--chart-exec);
}

.stat-icon-wrap.safe {
  background: var(--status-safe-dim);
  color: var(--status-safe);
}

.stat-icon-wrap.alert {
  background: rgba(251, 191, 36, 0.1);
  color: rgba(251, 191, 36, 0.9);
}

.stat-icon-wrap.blocked {
  background: rgba(239, 68, 68, 0.1);
  color: rgba(239, 68, 68, 0.9);
}

.stat-content {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 26px;
  font-weight: 700;
  font-family: var(--font-mono);
  color: var(--text-primary);
}

.stat-label {
  font-size: 12px;
  color: var(--text-muted);
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
  padding: 14px 16px;
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

.workload-row.status-alert {
  border-left: 3px solid rgba(251, 191, 36, 0.6);
}

.workload-row.status-blocked {
  border-left: 3px solid rgba(239, 68, 68, 0.7);
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
  gap: 8px;
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

/* Status Badge */
.status-badges {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: var(--radius-md);
  font-size: 12px;
  font-weight: 500;
}

.status-badge.ok {
  background: var(--status-safe-dim);
  color: var(--status-safe);
}

.status-badge.alert {
  background: rgba(251, 191, 36, 0.08);
  color: rgba(251, 191, 36, 0.9);
  border: 1px solid rgba(251, 191, 36, 0.2);
}

.status-badge.blocked {
  background: rgba(239, 68, 68, 0.08);
  color: rgba(239, 68, 68, 0.9);
  border: 1px solid rgba(239, 68, 68, 0.25);
  font-weight: 600;
}

@keyframes pulse {

  0%,
  100% {
    opacity: 1;
  }

  50% {
    opacity: 0.5;
  }
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
  margin: 0 0 14px 0;
}

.detail-stats {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.detail-stat {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  background: var(--bg-surface);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
}

.stat-icon {
  color: var(--text-muted);
}

.stat-icon.exec {
  color: var(--chart-exec);
}

.stat-icon.file {
  color: var(--chart-file);
}

.stat-icon.net {
  color: var(--chart-network);
}

.stat-info {
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

/* Security Timeline */
.security-timeline {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.security-status {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  border-radius: var(--radius-md);
}

.security-status.ok {
  background: var(--status-safe-dim);
  color: var(--status-safe);
}

.security-status.alert {
  background: rgba(251, 191, 36, 0.06);
  color: rgba(251, 191, 36, 0.85);
  border: 1px solid rgba(251, 191, 36, 0.15);
  backdrop-filter: blur(4px);
}

.security-status.blocked {
  background: rgba(239, 68, 68, 0.06);
  color: rgba(239, 68, 68, 0.85);
  border: 1px solid rgba(239, 68, 68, 0.18);
  backdrop-filter: blur(4px);
}

.security-label {
  font-size: 13px;
  font-weight: 500;
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
}

.action-btn.secondary {
  background: var(--bg-surface);
  color: var(--text-secondary);
  border: 1px solid var(--border-default);
}

.action-btn.secondary:hover {
  background: var(--bg-overlay);
  color: var(--text-primary);
  border-color: var(--accent-primary);
  transform: translateY(-1px);
}

.action-btn.security {
  background: var(--accent-primary);
  color: #fff;
  border: 1px solid var(--accent-primary);
}

.action-btn.security:hover {
  background: var(--accent-hover);
  border-color: var(--accent-hover);
  transform: translateY(-1px);
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

@media (max-width: 900px) {
  .stats-row {
    grid-template-columns: 1fr;
  }

  .details-grid {
    grid-template-columns: 1fr;
  }
}
</style>
