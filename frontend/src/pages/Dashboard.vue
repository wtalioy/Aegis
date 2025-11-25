<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Activity, AlertTriangle, Box, Container, ArrowRight } from 'lucide-vue-next'
import Card from '../components/common/Card.vue'
import StatCard from '../components/common/StatCard.vue'
import EventsChart from '../components/charts/EventsChart.vue'
import SeverityPie from '../components/charts/SeverityPie.vue'
import { useEvents } from '../composables/useEvents'
import { useAlerts } from '../composables/useAlerts'
import { getSystemStats, type SystemStats } from '../lib/api'

const { eventRate, totalEvents } = useEvents()
const { alerts, getAlertsBySeverity } = useAlerts()

const stats = ref<SystemStats>({
  processCount: 0,
  containerCount: 0,
  eventsPerSec: 0,
  alertCount: 0,
  probeStatus: 'starting'
})

const severityCounts = computed(() => getAlertsBySeverity())
const recentAlerts = computed(() => alerts.value.slice(0, 5))

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

const getSeverityClass = (severity: string) => {
  switch (severity) {
    case 'high': return 'severity-high'
    case 'warning': return 'severity-warning'
    default: return 'severity-info'
  }
}

onMounted(() => {
  fetchStats()
  setInterval(fetchStats, 3000)
})
</script>

<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <h1 class="page-title">Dashboard</h1>
      <span class="page-subtitle">Real-time security monitoring</span>
    </div>

    <!-- Stats Cards Row -->
    <div class="stats-grid">
      <StatCard 
        :value="eventRate.exec + eventRate.network + eventRate.file"
        label="Events / Second"
        :icon="Activity"
        color="info"
      />
      <StatCard 
        :value="stats.alertCount"
        label="Total Alerts"
        :icon="AlertTriangle"
        :color="stats.alertCount > 0 ? 'critical' : 'safe'"
      />
      <StatCard 
        :value="stats.processCount"
        label="Monitored Processes"
        :icon="Box"
      />
      <StatCard 
        :value="stats.containerCount"
        label="Containers"
        :icon="Container"
        color="info"
      />
    </div>

    <!-- Charts Row -->
    <div class="charts-grid">
      <Card title="Events Per Second" class="chart-card events-chart-card">
        <EventsChart />
      </Card>

      <Card title="Alert Severity Distribution" class="chart-card severity-card">
        <SeverityPie 
          :high="severityCounts.high" 
          :warning="severityCounts.warning" 
          :info="severityCounts.info" 
        />
        <div class="severity-legend">
          <div class="legend-item">
            <span class="legend-dot high"></span>
            <span class="legend-label">High</span>
            <span class="legend-value">{{ severityCounts.high }}</span>
          </div>
          <div class="legend-item">
            <span class="legend-dot warning"></span>
            <span class="legend-label">Warning</span>
            <span class="legend-value">{{ severityCounts.warning }}</span>
          </div>
          <div class="legend-item">
            <span class="legend-dot info"></span>
            <span class="legend-label">Info</span>
            <span class="legend-value">{{ severityCounts.info }}</span>
          </div>
        </div>
      </Card>
    </div>

    <!-- Recent Alerts -->
    <Card class="alerts-card">
      <template #default>
        <div class="alerts-header">
          <h3 class="alerts-title">Recent Alerts</h3>
          <router-link to="/alerts" class="view-all">
            View All <ArrowRight :size="16" />
          </router-link>
        </div>
        <div class="alerts-list" v-if="recentAlerts.length > 0">
          <div 
            v-for="alert in recentAlerts" 
            :key="alert.id"
            class="alert-item"
            :class="getSeverityClass(alert.severity)"
          >
            <span class="alert-indicator"></span>
            <span class="alert-time font-mono">{{ formatTime(alert.timestamp) }}</span>
            <span class="alert-rule">{{ alert.ruleName }}</span>
            <span class="alert-process font-mono">{{ alert.processName }}</span>
            <span class="alert-pid font-mono">PID {{ alert.pid }}</span>
            <span class="alert-badge" :class="alert.severity">{{ alert.severity.toUpperCase() }}</span>
          </div>
        </div>
        <div v-else class="no-alerts">
          <span class="no-alerts-icon">✓</span>
          <span class="no-alerts-text">No recent alerts</span>
        </div>
      </template>
    </Card>
  </div>
</template>

<style scoped>
.dashboard {
  max-width: 1400px;
}

.dashboard-header {
  margin-bottom: 24px;
}

.page-title {
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.page-subtitle {
  font-size: 14px;
  color: var(--text-muted);
}

/* Stats Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

/* Charts Grid */
.charts-grid {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 16px;
  margin-bottom: 24px;
}

.chart-card {
  min-height: 340px;
}

.severity-card :deep(.card-content) {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px;
}

.severity-legend {
  display: flex;
  gap: 24px;
  margin-top: 16px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.legend-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.legend-dot.high { background: var(--status-critical); }
.legend-dot.warning { background: var(--status-warning); }
.legend-dot.info { background: var(--status-info); }

.legend-label {
  font-size: 12px;
  color: var(--text-secondary);
}

.legend-value {
  font-size: 12px;
  font-weight: 600;
  font-family: var(--font-mono);
  color: var(--text-primary);
}

/* Alerts Card */
.alerts-card :deep(.card-content) {
  padding: 0;
}

.alerts-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-subtle);
}

.alerts-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
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

.alerts-list {
  display: flex;
  flex-direction: column;
}

.alert-item {
  display: grid;
  grid-template-columns: 4px 80px 1fr 120px 80px 80px;
  align-items: center;
  gap: 16px;
  padding: 12px 20px;
  border-bottom: 1px solid var(--border-subtle);
  transition: background var(--transition-fast);
}

.alert-item:last-child {
  border-bottom: none;
}

.alert-item:hover {
  background: var(--bg-hover);
}

.alert-indicator {
  width: 4px;
  height: 32px;
  border-radius: 2px;
  background: var(--text-muted);
}

.severity-high .alert-indicator { background: var(--status-critical); }
.severity-warning .alert-indicator { background: var(--status-warning); }
.severity-info .alert-indicator { background: var(--status-info); }

.alert-time {
  font-size: 12px;
  color: var(--text-muted);
}

.alert-rule {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.alert-process {
  font-size: 12px;
  color: var(--text-secondary);
}

.alert-pid {
  font-size: 11px;
  color: var(--text-muted);
}

.alert-badge {
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  font-size: 10px;
  font-weight: 600;
  text-align: center;
}

.alert-badge.high {
  background: var(--status-critical-dim);
  color: var(--status-critical);
}

.alert-badge.warning {
  background: var(--status-warning-dim);
  color: var(--status-warning);
}

.alert-badge.info {
  background: var(--status-info-dim);
  color: var(--status-info);
}

.no-alerts {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 20px;
  gap: 8px;
}

.no-alerts-icon {
  font-size: 32px;
  color: var(--status-safe);
}

.no-alerts-text {
  font-size: 14px;
  color: var(--text-muted);
}
</style>

