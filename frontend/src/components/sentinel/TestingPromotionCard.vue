<!-- Testing Promotion Card Component -->
<script setup lang="ts">
import { computed } from 'vue'
import { Target, TrendingUp, CheckCircle2, AlertTriangle } from 'lucide-vue-next'
import type { TestingPromotionInsight } from '../../types/sentinel'
import AIConfidenceBadge from '../ai/AIConfidenceBadge.vue'

const props = defineProps<{
  insight: TestingPromotionInsight
}>()

const emit = defineEmits<{
  promote: []
  dismiss: []
  askAI: []
}>()

const promotionReady = computed(() => {
  const data = props.insight.data
  return data.hits >= 10 && data.observation_hours >= 24 && (data.false_positives || 0) < data.hits * 0.1
})

const statsSummary = computed(() => {
  const data = props.insight.data
  return `${data.hits} hits over ${data.observation_hours}h, ${data.false_positives || 0} false positives`
})
</script>

<template>
  <div class="testing-promotion-card">
    <div class="card-header">
      <Target :size="20" class="icon" />
      <div class="header-content">
        <h3 class="card-title">{{ insight.title }}</h3>
        <AIConfidenceBadge :confidence="insight.confidence" size="sm" />
      </div>
    </div>

    <div class="card-body">
      <p class="summary">{{ insight.summary }}</p>
      
      <div class="stats-section">
        <div class="stat-item">
          <TrendingUp :size="16" />
          <div class="stat-content">
            <div class="stat-label">Total Hits</div>
            <div class="stat-value">{{ insight.data.hits }}</div>
          </div>
        </div>
        <div class="stat-item">
          <Target :size="16" />
          <div class="stat-content">
            <div class="stat-label">Observation Hours</div>
            <div class="stat-value">{{ insight.data.observation_hours }}h</div>
          </div>
        </div>
        <div class="stat-item">
          <AlertTriangle :size="16" />
          <div class="stat-content">
            <div class="stat-label">False Positives</div>
            <div class="stat-value">{{ insight.data.false_positives || 0 }}</div>
          </div>
        </div>
      </div>

      <div class="rule-info">
        <div class="rule-name">Rule: {{ insight.data.rule_name }}</div>
        <div class="stats-text">{{ statsSummary }}</div>
      </div>

      <div v-if="promotionReady" class="ready-badge">
        <CheckCircle2 :size="14" />
        <span>Ready for promotion</span>
      </div>
    </div>

    <div class="card-actions">
      <button class="action-btn primary" @click="$emit('promote')">
        <CheckCircle2 :size="16" />
        <span>Promote to Production</span>
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
.testing-promotion-card {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-left: 4px solid rgb(59, 130, 246);
  border-radius: var(--radius-lg);
  padding: 20px;
  transition: all 0.2s;
}

.testing-promotion-card:hover {
  border-color: var(--border-default);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.icon {
  color: rgb(59, 130, 246);
  flex-shrink: 0;
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

.stats-section {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 16px;
  padding: 16px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.stat-content {
  display: flex;
  flex-direction: column;
}

.stat-label {
  font-size: 11px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat-value {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.rule-info {
  padding: 12px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
  margin-bottom: 12px;
}

.rule-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.stats-text {
  font-size: 12px;
  color: var(--text-muted);
}

.ready-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  background: rgba(34, 197, 94, 0.1);
  border: 1px solid rgba(34, 197, 94, 0.2);
  border-radius: var(--radius-md);
  color: rgb(34, 197, 94);
  font-size: 13px;
  font-weight: 500;
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
  background: var(--accent-primary);
  color: white;
  border-color: var(--accent-primary);
}

.action-btn.primary:hover {
  opacity: 0.9;
}

.action-btn.ask-ai {
  background: transparent;
  border: 1px dashed var(--border-subtle);
}
</style>

