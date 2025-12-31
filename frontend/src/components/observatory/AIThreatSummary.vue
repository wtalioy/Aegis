<!-- AI Threat Summary Component - Updated to use global card styles -->
<script setup lang="ts">
import { computed } from 'vue'
import { AlertTriangle, Skull, CheckCircle2 } from 'lucide-vue-next'

const props = defineProps<{
  threats: Array<{
    type: string
    severity: 'low' | 'medium' | 'high' | 'critical'
    count: number
    description: string
    ruleName: string
  }>
  aiSummary?: string
}>()

const criticalThreats = computed(() =>
  props.threats.filter(t => t.severity === 'critical' || t.severity === 'high')
)

const totalThreats = computed(() =>
  props.threats.reduce((sum, t) => sum + t.count, 0)
)

const severityColor = (severity: string) => {
  switch (severity) {
    case 'critical': return 'var(--status-critical)'
    case 'high': return 'var(--status-warning)' // Mapped high to warning for color
    case 'medium': return 'var(--status-warning)'
    default: return 'var(--status-info)'
  }
}
</script>

<template>
  <div class="card-base threat-summary">
    <div class="card-content-base">
      <div class="summary-header">
        <AlertTriangle :size="20" />
        <h3>Threat Summary</h3>
        <div class="threat-count">{{ totalThreats }} threats</div>
      </div>

      <div v-if="criticalThreats.length > 0" class="critical-section">
        <div class="section-title">
          <Skull :size="16" />
          <span>Critical Threats</span>
        </div>
        <div class="threats-list">
          <div v-for="(threat, idx) in criticalThreats" :key="idx" class="threat-item"
            :style="{ borderLeftColor: severityColor(threat.severity) }">
            <div class="threat-header">
              <span class="threat-type">{{ threat.type }}</span>
              <span class="threat-count-badge"
                :style="{ backgroundColor: severityColor(threat.severity) + '20', color: severityColor(threat.severity) }">
                {{ threat.count }}
              </span>
            </div>
            <div class="threat-description">{{ threat.description }}</div>
          </div>
        </div>
      </div>

      <div v-if="props.threats.length > 0" class="all-threats">
        <div class="section-title">All Threats</div>
        <div class="threats-grid">
          <div v-for="(threat, idx) in props.threats" :key="idx" class="threat-card"
            @click="$emit('threat-click', threat)" style="cursor: pointer;">
            <div class="card-header">
              <span class="threat-type">{{ threat.type }}</span>
              <span class="severity-badge" :style="{ color: severityColor(threat.severity) }">
                {{ threat.severity }}
              </span>
            </div>
            <div class="card-count">{{ threat.count }} occurrences</div>
          </div>
        </div>
      </div>

      <div v-if="props.threats.length === 0" class="no-threats">
        <CheckCircle2 :size="32" />
        <p>No active threats detected</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.threat-summary {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.card-content-base {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.summary-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.summary-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
  flex: 1;
}

.threat-count {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-secondary);
  padding: 4px 12px;
  background: var(--bg-overlay);
  border-radius: var(--radius-md);
}

.critical-section {
  margin-bottom: 24px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 12px;
}

.threats-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.threat-item {
  padding: 16px;
  background: var(--bg-overlay);
  border-left: 4px solid;
  border-radius: var(--radius-md);
}

.threat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.threat-type {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.threat-count-badge {
  padding: 4px 10px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 600;
}

.threat-description {
  font-size: 13px;
  line-height: 1.5;
  color: var(--text-secondary);
}

.all-threats {
  margin-top: auto;
  /* Pushes to the bottom */
  padding-top: 24px;
  border-top: 1px solid var(--border-subtle);
}

.threats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
}

.threat-card {
  padding: 12px;
  background: var(--bg-overlay);
  border-radius: var(--radius-md);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.severity-badge {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.card-count {
  font-size: 12px;
  color: var(--text-muted);
}

.no-threats {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  text-align: center;
  color: var(--text-muted);
}

.no-threats p {
  margin-top: 16px;
  font-size: 14px;
}
</style>