<!-- Anomaly Card Component - Phase 4 -->
<script setup lang="ts">
import { computed } from 'vue'
import { AlertTriangle, TrendingUp, Activity } from 'lucide-vue-next'
import type { AnomalyInsight } from '../../types/sentinel'
import AIConfidenceBadge from '../ai/AIConfidenceBadge.vue'

const props = defineProps<{
  insight: AnomalyInsight
}>()

const emit = defineEmits<{
  investigate: []
  dismiss: []
  askAI: []
}>()

const severityColor = computed(() => {
  switch (props.insight.severity) {
    case 'critical': return 'var(--status-critical)'
    case 'high': return 'var(--status-high)'
    case 'medium': return 'var(--status-warning)'
    default: return 'var(--status-safe)'
  }
})

const deviationText = computed(() => {
  const dev = props.insight.data.deviation
  if (dev > 3) return `${dev.toFixed(1)}σ - Extreme anomaly`
  if (dev > 2) return `${dev.toFixed(1)}σ - Significant anomaly`
  return `${dev.toFixed(1)}σ - Moderate anomaly`
})
</script>

<template>
  <div class="anomaly-card" :style="{ borderLeftColor: severityColor }">
    <div class="card-header">
      <AlertTriangle :size="20" :style="{ color: severityColor }" />
      <div class="header-content">
        <h3 class="card-title">{{ insight.title }}</h3>
        <AIConfidenceBadge :confidence="insight.confidence" size="sm" />
      </div>
    </div>

    <div class="card-body">
      <p class="summary">{{ insight.summary }}</p>
      
      <div class="anomaly-details">
        <div class="detail-item">
          <Activity :size="16" />
          <div class="detail-content">
            <div class="detail-label">Anomaly Type</div>
            <div class="detail-value">{{ insight.data.anomaly_type }}</div>
          </div>
        </div>
        <div class="detail-item">
          <TrendingUp :size="16" />
          <div class="detail-content">
            <div class="detail-label">Deviation</div>
            <div class="detail-value">{{ deviationText }}</div>
          </div>
        </div>
        <div v-if="insight.data.process_name" class="detail-item">
          <div class="detail-content">
            <div class="detail-label">Process</div>
            <div class="detail-value">{{ insight.data.process_name }} (PID: {{ insight.data.process_id }})</div>
          </div>
        </div>
      </div>
    </div>

    <div class="card-actions">
      <button class="action-btn primary" @click="$emit('investigate')">
        <AlertTriangle :size="16" />
        <span>Investigate</span>
      </button>
      <button class="action-btn secondary" @click="$emit('dismiss')">
        Dismiss
      </button>
      <button class="action-btn ask-ai" @click="$emit('askAI')">
        Ask AI
      </button>
    </div>
  </div>
</template>

<style scoped>
.anomaly-card {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-left: 4px solid;
  border-radius: var(--radius-lg);
  padding: 20px;
  transition: all 0.2s;
}

.anomaly-card:hover {
  border-color: var(--border-default);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.header-content {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.card-body {
  margin-bottom: 16px;
}

.summary {
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-secondary);
  margin-bottom: 16px;
}

.anomaly-details {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
}

.detail-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.detail-content {
  flex: 1;
}

.detail-label {
  font-size: 11px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 4px;
}

.detail-value {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
}

.card-actions {
  display: flex;
  gap: 8px;
  padding-top: 16px;
  border-top: 1px solid var(--border-subtle);
  flex-wrap: wrap;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-elevated);
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.action-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.action-btn.primary {
  background: var(--status-critical);
  color: white;
  border-color: var(--status-critical);
}

.action-btn.primary:hover {
  opacity: 0.9;
}

.action-btn.ask-ai {
  background: transparent;
  border: 1px dashed var(--border-subtle);
}
</style>

