<!-- Optimization Card Component - Phase 4 -->
<script setup lang="ts">
import { computed } from 'vue'
import { TrendingUp, Lightbulb, Zap } from 'lucide-vue-next'
import type { OptimizationInsight } from '../../types/sentinel'
import AIConfidenceBadge from '../ai/AIConfidenceBadge.vue'

const props = defineProps<{
  insight: OptimizationInsight
}>()

const emit = defineEmits<{
  apply: []
  dismiss: []
  askAI: []
}>()

const rulesCount = computed(() => props.insight.data.rule_names?.length || 0)
</script>

<template>
  <div class="optimization-card">
    <div class="card-header">
      <Lightbulb :size="20" class="icon" />
      <div class="header-content">
        <h3 class="card-title">{{ insight.title }}</h3>
        <AIConfidenceBadge :confidence="insight.confidence" size="sm" />
      </div>
    </div>

    <div class="card-body">
      <p class="summary">{{ insight.summary }}</p>
      
      <div class="optimization-section">
        <div class="suggestion-header">
          <Zap :size="16" />
          <span>Suggestion</span>
        </div>
        <div class="suggestion-text">{{ insight.data.suggestion }}</div>
      </div>

      <div v-if="rulesCount > 0" class="rules-section">
        <div class="rules-header">Affected Rules:</div>
        <div class="rules-list">
          <span
            v-for="(ruleName, idx) in insight.data.rule_names"
            :key="idx"
            class="rule-tag"
          >
            {{ ruleName }}
          </span>
        </div>
      </div>
    </div>

    <div class="card-actions">
      <button class="action-btn primary" @click="$emit('apply')">
        <TrendingUp :size="16" />
        <span>Apply Optimization</span>
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
.optimization-card {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-left: 4px solid rgb(59, 130, 246);
  border-radius: var(--radius-lg);
  padding: 20px;
  transition: all 0.2s;
}

.optimization-card:hover {
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

.optimization-section {
  padding: 16px;
  background: rgba(59, 130, 246, 0.1);
  border: 1px solid rgba(59, 130, 246, 0.2);
  border-radius: var(--radius-md);
  margin-bottom: 16px;
}

.suggestion-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.suggestion-text {
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-primary);
}

.rules-section {
  padding-top: 16px;
  border-top: 1px solid var(--border-subtle);
}

.rules-header {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.rules-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.rule-tag {
  padding: 4px 10px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  font-size: 12px;
  color: var(--text-secondary);
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

