<!-- Observatory Page - Phase 4: AI-Driven Situational Awareness -->
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getSystemStats, getAlerts, getRules, type SystemStats, type Alert, type DetectionRule } from '../lib/api'
import AIHealthScore from '../components/observatory/AIHealthScore.vue'
import AIThreatSummary from '../components/observatory/AIThreatSummary.vue'
import DefenseStats from '../components/observatory/DefenseStats.vue'
import SentinelPreview from '../components/observatory/SentinelPreview.vue'
import QuickAsk from '../components/ai/QuickAsk.vue'

const stats = ref<SystemStats | null>(null)
const alerts = ref<Alert[]>([])
const rules = ref<DetectionRule[]>([])
const loading = ref(true)

// Calculate health score from real data
const healthScore = computed(() => {
  if (!stats.value) return 100

  const alertCount = stats.value.alertCount
  const processCount = stats.value.processCount
  const eventsPerSec = stats.value.eventsPerSec

  // Base score starts at 100
  let score = 100

  // Deduct points for alerts (more alerts = lower score)
  // Each alert reduces score by 2 points, max deduction of 40
  score -= Math.min(alertCount * 2, 40)

  // Deduct points for high event rate (potential anomaly)
  // Events/sec > 1000 reduces score
  if (eventsPerSec > 1000) {
    score -= Math.min((eventsPerSec - 1000) / 100, 20)
  }

  // Ensure score is between 0 and 100
  return Math.max(0, Math.min(100, Math.round(score)))
})

// Generate AI assessment from real data
const aiAssessment = computed(() => {
  if (!stats.value || alerts.value.length === 0) {
    return 'System is operating normally with no active threats detected.'
  }

  const criticalAlerts = alerts.value.filter(a => a.severity === 'critical').length
  const highAlerts = alerts.value.filter(a => a.severity === 'high').length
  const blockedCount = alerts.value.filter(a => a.blocked).length

  if (criticalAlerts > 0) {
    return `System has ${criticalAlerts} critical alert${criticalAlerts > 1 ? 's' : ''} requiring immediate attention. ${blockedCount > 0 ? `${blockedCount} threat${blockedCount > 1 ? 's' : ''} have been blocked.` : ''}`
  }

  if (highAlerts > 0) {
    return `System has ${highAlerts} high-severity alert${highAlerts > 1 ? 's' : ''}. ${blockedCount > 0 ? `${blockedCount} threat${blockedCount > 1 ? 's' : ''} have been blocked.` : 'Monitoring recommended.'}`
  }

  if (stats.value.alertCount > 0) {
    return `System has ${stats.value.alertCount} alert${stats.value.alertCount > 1 ? 's' : ''} of low to medium severity. ${blockedCount > 0 ? `${blockedCount} threat${blockedCount > 1 ? 's' : ''} have been blocked.` : 'System is functioning normally.'}`
  }

  return 'System is operating normally with no active threats detected.'
})

// Calculate threats from real alerts
const threats = computed(() => {
  if (alerts.value.length === 0) {
    return []
  }

  // Group alerts by type/severity
  const threatMap = new Map<string, { count: number; severity: string; description: string; ruleName: string }>()

  alerts.value.forEach(alert => {
    const key = `${alert.severity}-${alert.ruleName}`
    if (!threatMap.has(key)) {
      threatMap.set(key, {
        count: 0,
        severity: alert.severity,
        description: alert.description || `Alert from rule: ${alert.ruleName}`,
        ruleName: alert.ruleName
      })
    }
    threatMap.get(key)!.count++
  })

  // Convert to array and sort by severity
  const threatArray = Array.from(threatMap.entries()).map(([key, data]) => ({
    type: data.severity === 'critical' ? 'Critical Threat' :
      data.severity === 'high' ? 'High Severity Alert' :
        data.severity === 'warning' ? 'Warning' : 'Info',
    severity: (data.severity === 'critical' ? 'critical' :
      data.severity === 'high' ? 'high' :
        data.severity === 'warning' ? 'medium' : 'low') as 'critical' | 'high' | 'medium' | 'low',
    count: data.count,
    description: data.description,
    ruleName: data.ruleName
  }))

  // Sort by severity (critical > high > medium > low)
  const severityOrder = { critical: 0, high: 1, medium: 2, low: 3 }
  threatArray.sort((a, b) => severityOrder[a.severity] - severityOrder[b.severity])

  return threatArray.slice(0, 5) // Return top 5 threats
})

// Calculate defense stats from real data
import { useRouter } from 'vue-router'

const router = useRouter()

const handleThreatClick = (threat: any) => {
  if (threat.ruleName) {
    router.push({ name: 'investigation', query: { search: threat.ruleName } })
  }
}

// Calculate defense stats from real data
const defenseStats = computed(() => {
  const totalBlocks = alerts.value.filter(a => a.blocked).length
  const highSeverityAlerts = alerts.value.filter(a => a.severity === 'high' || a.severity === 'critical').length
  const activeRules = rules.value.filter(r => (r as any).state === 'production' || (r as any).state === 'testing').length
  const testingRules = rules.value.filter(r => (r as any).state === 'testing').length

  // Defense rate: considers blocks against high-severity threats
  const defenseRate = highSeverityAlerts > 0
    ? Math.round((totalBlocks / highSeverityAlerts) * 100)
    : 100

  return {
    totalBlocks,
    totalAlerts: alerts.value.length,
    activeRules,
    testingRules,
    defenseRate
  }
})

onMounted(async () => {
  try {
    loading.value = true
    // Fetch all data in parallel
    const [statsData, alertsData, rulesData] = await Promise.all([
      getSystemStats(),
      getAlerts(),
      getRules()
    ])

    stats.value = statsData
    alerts.value = alertsData
    rules.value = rulesData
  } catch (err) {
    console.error('Failed to load Observatory data:', err)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="observatory-page">
    <div class="page-header">
      <h1>Observatory</h1>
      <p class="page-subtitle">AI-driven situational awareness</p>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Loading system data...</p>
    </div>

    <template v-else>
      <div class="health-section">
        <AIHealthScore :score="healthScore" :assessment="aiAssessment" trend="stable" />
      </div>

      <div class="main-grid">
        <div class="threat-section">
          <AIThreatSummary :threats="threats" :ai-summary="aiAssessment" @threat-click="handleThreatClick" />
        </div>
        <div class="defense-section">
          <DefenseStats :stats="defenseStats" trend="stable" />
        </div>
      </div>

      <div class="sentinel-section">
        <SentinelPreview />
      </div>

      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-label">Processes</div>
          <div class="stat-value">{{ stats?.processCount || 0 }}</div>
          <QuickAsk question="What processes are most active?" />
        </div>
        <div class="stat-card">
          <div class="stat-label">Workloads</div>
          <div class="stat-value">{{ stats?.workloadCount || 0 }}</div>
          <QuickAsk question="Which workloads have alerts?" />
        </div>
        <div class="stat-card">
          <div class="stat-label">Events/sec</div>
          <div class="stat-value">{{ stats?.eventsPerSec?.toFixed(1) || 0 }}</div>
          <QuickAsk question="Is the current event rate normal?" />
        </div>
        <div class="stat-card">
          <div class="stat-label">Alerts</div>
          <div class="stat-value">{{ stats?.alertCount || 0 }}</div>
          <QuickAsk question="Show me the most recent alerts" />
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.observatory-page {
  padding: 32px;
  /* Increased padding for more space */
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 40px;
  /* Increased margin */
}

.page-header h1 {
  font-size: 28px;
  /* Softened font size */
  font-weight: 600;
  /* Softened font weight */
  color: var(--text-primary);
  margin: 0 0 8px 0;
}

.page-subtitle {
  font-size: 16px;
  color: var(--text-secondary);
  margin: 0;
}

.health-section {
  margin-bottom: 40px;
  /* Increased margin */
}

.main-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 32px;
  /* Increased gap */
  margin-bottom: 40px;
  /* Increased margin */
  align-items: stretch;
}

.threat-section,
.defense-section {
  display: flex;
  height: 100%;
}

.threat-section>*,
.defense-section>* {
  width: 100%;
  display: flex;
  flex-direction: column;
}

.sentinel-section {
  margin-bottom: 40px;
  /* Increased margin */
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 24px;
  /* Increased gap */
}

.stat-card {
  padding: 20px;
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  transition: background-color var(--transition-fast);
}

.stat-card:hover {
  background: var(--bg-hover);
}

.stat-label {
  font-size: 14px;
  color: var(--text-muted);
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat-value {
  font-size: 28px;
  /* Softened font size */
  font-weight: 600;
  /* Softened font weight */
  color: var(--text-primary);
  margin-bottom: 12px;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 24px;
  text-align: center;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border-subtle);
  border-top-color: var(--accent-primary);
  /* Consistent spinner color */
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 16px;
}

.loading-state p {
  color: var(--text-secondary);
  font-size: 14px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>