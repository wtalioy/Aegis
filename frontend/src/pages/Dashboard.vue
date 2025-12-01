<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import {
  Shield, ShieldCheck, ShieldOff, AlertTriangle,
  Activity, Terminal, FileText, Globe,
  Box, Boxes, ArrowRight, Zap
} from 'lucide-vue-next'
import Card from '../components/common/Card.vue'
import EventsChart from '../components/charts/EventsChart.vue'
import { useEvents } from '../composables/useEvents'
import { useAlerts } from '../composables/useAlerts'
import { getSystemStats, type SystemStats } from '../lib/api'

const { eventRate } = useEvents()
const { alerts, getAlertsBySeverity, getAlertsByAction } = useAlerts()

const stats = ref<SystemStats>({
  processCount: 0,
  workloadCount: 0,
  eventsPerSec: 0,
  alertCount: 0,
  probeStatus: 'starting'
})

const severityCounts = computed(() => getAlertsBySeverity())
const actionCounts = computed(() => getAlertsByAction())
const recentEvents = computed(() => alerts.value.slice(0, 5))

const fetchStats = async () => {
  try {
    const result = await getSystemStats()
    stats.value = result
  } catch (e) {
    console.error('Failed to fetch stats:', e)
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

const totalEvents = computed(() => eventRate.value.exec + eventRate.value.network + eventRate.value.file)
const totalThreats = computed(() => actionCounts.value.blocked + actionCounts.value.alerted)

const defenseRate = computed(() => {
  if (totalThreats.value === 0) return 100
  return Math.round((actionCounts.value.blocked / totalThreats.value) * 100)
})

// Donut chart values
const donutStroke = computed(() => {
  if (totalThreats.value === 0) return { blocked: 0, alerted: 0 }
  const circumference = 2 * Math.PI * 45
  const blockedPct = actionCounts.value.blocked / totalThreats.value
  return {
    blocked: blockedPct * circumference,
    alerted: (1 - blockedPct) * circumference,
    circumference
  }
})

onMounted(() => {
  fetchStats()
  setInterval(fetchStats, 3000)
})
</script>

<template>
  <div class="dashboard">
    <!-- Main Grid -->
    <div class="main-grid">
      <!-- Left Column: Defense Overview -->
      <div class="defense-section">
        <div class="section-header">
          <ShieldCheck :size="18" />
          <span>Active Defense</span>
        </div>

        <div class="defense-donut">
          <svg viewBox="0 0 100 100" class="donut-svg">
            <!-- Background circle -->
            <circle cx="50" cy="50" r="45" class="donut-bg" />
            <!-- Blocked arc -->
            <circle cx="50" cy="50" r="45" class="donut-blocked"
              :stroke-dasharray="`${donutStroke.blocked} ${donutStroke.circumference}`" />
            <!-- Alerted arc -->
            <circle cx="50" cy="50" r="45" class="donut-alerted"
              :stroke-dasharray="`${donutStroke.alerted} ${donutStroke.circumference}`"
              :stroke-dashoffset="`-${donutStroke.blocked}`" />
          </svg>
          <div class="donut-center">
            <span class="donut-value">{{ defenseRate }}%</span>
            <span class="donut-label">Protected</span>
          </div>
        </div>

        <div class="defense-stats">
          <div class="defense-stat blocked">
            <ShieldOff :size="20" />
            <div class="stat-content">
              <span class="stat-value">{{ actionCounts.blocked }}</span>
              <span class="stat-label">Blocked</span>
            </div>
          </div>
          <div class="defense-stat alerted">
            <AlertTriangle :size="20" />
            <div class="stat-content">
              <span class="stat-value">{{ actionCounts.alerted }}</span>
              <span class="stat-label">Alerted</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Center Column: Hook Monitors -->
      <div class="monitors-section">
        <div class="section-header">
          <Activity :size="18" />
          <span>LSM Hook Monitors</span>
        </div>

        <div class="monitors-grid">
          <div class="monitor-card exec">
            <div class="monitor-icon">
              <Terminal :size="24" />
            </div>
            <div class="monitor-info">
              <span class="monitor-name">Process Exec</span>
              <span class="monitor-hook">bprm_check_security</span>
            </div>
            <div class="monitor-rate">
              <span class="rate-value">{{ eventRate.exec }}</span>
              <span class="rate-unit">/s</span>
            </div>
          </div>

          <div class="monitor-card file">
            <div class="monitor-icon">
              <FileText :size="24" />
            </div>
            <div class="monitor-info">
              <span class="monitor-name">File Access</span>
              <span class="monitor-hook">file_open</span>
            </div>
            <div class="monitor-rate">
              <span class="rate-value">{{ eventRate.file }}</span>
              <span class="rate-unit">/s</span>
            </div>
          </div>

          <div class="monitor-card network">
            <div class="monitor-icon">
              <Globe :size="24" />
            </div>
            <div class="monitor-info">
              <span class="monitor-name">Network</span>
              <span class="monitor-hook">socket_connect</span>
            </div>
            <div class="monitor-rate">
              <span class="rate-value">{{ eventRate.network }}</span>
              <span class="rate-unit">/s</span>
            </div>
          </div>
        </div>

        <!-- System Stats -->
        <div class="system-stats">
          <div class="sys-stat">
            <Box :size="16" />
            <span class="sys-value">{{ stats.processCount }}</span>
            <span class="sys-label">Processes</span>
          </div>
          <div class="sys-stat">
            <Boxes :size="16" />
            <span class="sys-value">{{ stats.workloadCount }}</span>
            <span class="sys-label">Workloads</span>
          </div>
          <div class="sys-stat">
            <Activity :size="16" />
            <span class="sys-value">{{ totalEvents }}</span>
            <span class="sys-label">Events/s</span>
          </div>
        </div>
      </div>

      <!-- Right Column: Severity -->
      <div class="severity-section">
        <div class="section-header">
          <AlertTriangle :size="18" />
          <span>Threat Severity</span>
        </div>

        <div class="severity-list">
          <div class="severity-item critical">
            <div class="severity-info">
              <span class="severity-dot"></span>
              <span class="severity-name">Critical</span>
            </div>
            <span class="severity-count">{{ severityCounts.critical }}</span>
          </div>
          <div class="severity-item high">
            <div class="severity-info">
              <span class="severity-dot"></span>
              <span class="severity-name">High</span>
            </div>
            <span class="severity-count">{{ severityCounts.high }}</span>
          </div>
          <div class="severity-item warning">
            <div class="severity-info">
              <span class="severity-dot"></span>
              <span class="severity-name">Warning</span>
            </div>
            <span class="severity-count">{{ severityCounts.warning }}</span>
          </div>
          <div class="severity-item info">
            <div class="severity-info">
              <span class="severity-dot"></span>
              <span class="severity-name">Info</span>
            </div>
            <span class="severity-count">{{ severityCounts.info }}</span>
          </div>
        </div>

        <div class="severity-total">
          <span class="total-label">Total Threats</span>
          <span class="total-value">{{ totalThreats }}</span>
        </div>
      </div>
    </div>

    <!-- Event Timeline -->
    <Card title="Event Timeline" class="timeline-card">
      <EventsChart />
    </Card>

    <!-- Recent Events -->
    <div class="events-section">
      <div class="events-header">
        <div class="events-title">
          <Shield :size="18" />
          <span>Recent Security Events</span>
        </div>
        <router-link to="/alerts" class="view-all">
          View All
          <ArrowRight :size="16" />
        </router-link>
      </div>

      <div class="events-table" v-if="recentEvents.length > 0">
        <div v-for="event in recentEvents" :key="event.id" class="event-row" :class="{ blocked: event.blocked }">
          <div class="event-action">
            <span class="action-badge" :class="event.blocked ? 'blocked' : 'alerted'">
              <component :is="event.blocked ? ShieldOff : AlertTriangle" :size="12" />
              {{ event.blocked ? 'BLOCKED' : 'ALERT' }}
            </span>
          </div>
          <div class="event-time">{{ formatTime(event.timestamp) }}</div>
          <div class="event-rule">{{ event.ruleName }}</div>
          <div class="event-process">{{ event.processName }}</div>
          <div class="event-severity">
            <span class="severity-tag" :class="event.severity">
              {{ event.severity }}
            </span>
          </div>
        </div>
      </div>

      <div v-else class="events-empty">
        <ShieldCheck :size="48" />
        <span class="empty-title">All Clear</span>
        <span class="empty-text">No security events detected</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* Main Grid */
.main-grid {
  display: grid;
  grid-template-columns: 260px 1fr 220px;
  gap: 20px;
}

/* Section Headers */
.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

/* Defense Section */
.defense-section {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: 20px;
}

.defense-donut {
  position: relative;
  width: 180px;
  height: 180px;
  margin: 0 auto 20px;
}

.donut-svg {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.donut-bg {
  fill: none;
  stroke: var(--bg-overlay);
  stroke-width: 8;
}

.donut-blocked {
  fill: none;
  stroke: var(--status-blocked);
  stroke-width: 8;
  stroke-linecap: round;
  transition: stroke-dasharray 0.5s ease;
}

.donut-alerted {
  fill: none;
  stroke: var(--status-warning);
  stroke-width: 8;
  stroke-linecap: round;
  transition: all 0.5s ease;
}

.donut-center {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
}

.donut-value {
  display: block;
  font-size: 32px;
  font-weight: 700;
  font-family: var(--font-mono);
  color: var(--status-safe);
}

.donut-label {
  display: block;
  font-size: 11px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 1px;
}

.defense-stats {
  display: flex;
  gap: 12px;
}

.defense-stat {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
}

.defense-stat.blocked {
  color: var(--status-blocked);
}

.defense-stat.alerted {
  color: var(--status-warning);
}

.stat-content {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 20px;
  font-weight: 700;
  font-family: var(--font-mono);
  color: inherit;
}

.stat-label {
  font-size: 10px;
  color: var(--text-muted);
  text-transform: uppercase;
}

/* Monitors Section */
.monitors-section {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: 20px;
}

.monitors-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 20px;
}

.monitor-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 20px 16px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
  transition: all var(--transition-fast);
}

.monitor-card:hover {
  border-color: var(--border-default);
  transform: translateY(-2px);
}

.monitor-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
}

.monitor-card.exec .monitor-icon {
  background: rgba(96, 165, 250, 0.15);
  color: var(--chart-exec);
}

.monitor-card.file .monitor-icon {
  background: rgba(16, 185, 129, 0.15);
  color: var(--chart-file);
}

.monitor-card.network .monitor-icon {
  background: rgba(245, 158, 11, 0.15);
  color: var(--chart-network);
}

.monitor-info {
  text-align: center;
}

.monitor-name {
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.monitor-hook {
  display: block;
  font-size: 10px;
  font-family: var(--font-mono);
  color: var(--text-muted);
}

.monitor-rate {
  display: flex;
  align-items: baseline;
  gap: 2px;
}

.rate-value {
  font-size: 24px;
  font-weight: 700;
  font-family: var(--font-mono);
  color: var(--text-primary);
}

.rate-unit {
  font-size: 12px;
  color: var(--text-muted);
}

.system-stats {
  display: flex;
  justify-content: space-around;
  padding-top: 16px;
  border-top: 1px solid var(--border-subtle);
}

.sys-stat {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-muted);
}

.sys-value {
  font-size: 16px;
  font-weight: 600;
  font-family: var(--font-mono);
  color: var(--text-primary);
}

.sys-label {
  font-size: 11px;
}

/* Severity Section */
.severity-section {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: 20px;
}

.severity-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 20px;
}

.severity-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
}

.severity-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.severity-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.severity-item.critical .severity-dot {
  background: var(--status-blocked);
}

.severity-item.high .severity-dot {
  background: var(--status-critical);
}

.severity-item.warning .severity-dot {
  background: var(--status-warning);
}

.severity-item.info .severity-dot {
  background: var(--status-info);
}

.severity-name {
  font-size: 13px;
  color: var(--text-secondary);
}

.severity-count {
  font-size: 16px;
  font-weight: 600;
  font-family: var(--font-mono);
  color: var(--text-primary);
}

.severity-total {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 16px;
  border-top: 1px solid var(--border-subtle);
}

.total-label {
  font-size: 12px;
  color: var(--text-muted);
}

.total-value {
  font-size: 24px;
  font-weight: 700;
  font-family: var(--font-mono);
  color: var(--text-primary);
}

/* Timeline Card */
.timeline-card {
  min-height: 280px;
}

/* Events Section */
.events-section {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.events-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-subtle);
}

.events-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.view-all {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: var(--accent-primary);
}

.view-all:hover {
  color: var(--accent-primary-hover);
}

.events-table {
  display: flex;
  flex-direction: column;
}

.event-row {
  display: grid;
  grid-template-columns: 100px 80px 1fr 150px 80px;
  align-items: center;
  gap: 16px;
  padding: 14px 20px;
  border-bottom: 1px solid var(--border-subtle);
  transition: background var(--transition-fast);
}

.event-row:last-child {
  border-bottom: none;
}

.event-row:hover {
  background: var(--bg-hover);
}

.event-row.blocked {
  background: linear-gradient(90deg, var(--status-blocked-dim), transparent 40%);
}

.action-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.action-badge.blocked {
  background: var(--status-blocked);
  color: #fff;
}

.action-badge.alerted {
  background: var(--status-warning-dim);
  color: var(--status-warning);
}

.event-time {
  font-size: 12px;
  font-family: var(--font-mono);
  color: var(--text-muted);
}

.event-rule {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.event-process {
  font-size: 12px;
  font-family: var(--font-mono);
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
}

.severity-tag {
  display: inline-block;
  padding: 3px 8px;
  border-radius: var(--radius-sm);
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
}

.severity-tag.critical {
  background: var(--status-blocked-dim);
  color: var(--status-blocked);
}

.severity-tag.high {
  background: var(--status-critical-dim);
  color: var(--status-critical);
}

.severity-tag.warning {
  background: var(--status-warning-dim);
  color: var(--status-warning);
}

.severity-tag.info {
  background: var(--status-info-dim);
  color: var(--status-info);
}

.events-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 20px;
  gap: 12px;
  color: var(--status-safe);
}

.empty-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}

.empty-text {
  font-size: 13px;
  color: var(--text-muted);
}

/* Responsive */
@media (max-width: 1200px) {
  .main-grid {
    grid-template-columns: 1fr 1fr;
  }

  .defense-section {
    grid-column: 1 / -1;
  }

  .defense-donut {
    width: 140px;
    height: 140px;
  }

  .monitors-section {
    grid-column: 1 / 2;
  }

  .severity-section {
    grid-column: 2 / 3;
  }
}

@media (max-width: 900px) {
  .main-grid {
    grid-template-columns: 1fr;
  }

  .monitors-section,
  .severity-section {
    grid-column: 1;
  }

  .monitors-grid {
    grid-template-columns: 1fr;
  }

  .event-row {
    grid-template-columns: 90px 1fr 70px;
  }

  .event-time,
  .event-process {
    display: none;
  }
}

@media (max-width: 600px) {
  .dashboard-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .header-right {
    width: 100%;
    justify-content: flex-start;
  }
}
</style>
