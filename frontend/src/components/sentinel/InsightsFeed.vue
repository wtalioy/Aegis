<!-- Insights Feed Component - Phase 4 -->
<script setup lang="ts">
import { computed } from 'vue'
import { useSentinel, type Insight } from '../../composables/useSentinel'
import InsightCard from './InsightCard.vue'
import TestingPromotionCard from './TestingPromotionCard.vue'
import AnomalyCard from './AnomalyCard.vue'
import OptimizationCard from './OptimizationCard.vue'
import DeepAskAI from './DeepAskAI.vue'

const { insights, loading, error, executeAction } = useSentinel()

const groupedInsights = computed(() => {
  const groups: Record<string, Insight[]> = {
    testing_promotion: [],
    anomaly: [],
    optimization: [],
    daily_report: []
  }

  insights.value.forEach(insight => {
    if (groups[insight.type]) {
      groups[insight.type].push(insight)
    }
  })

  return groups
})

const handleAction = (insight: Insight, actionId: string) => {
  executeAction(insight, actionId)
}

const handleAskAI = (insight: Insight) => {
  // Navigate to AI chat with context
  window.location.href = `/ai?context=insight&id=${insight.id}`
}
</script>

<template>
  <div class="insights-feed">
    <div v-if="loading" class="loading-state">
      Loading insights...
    </div>

    <div v-else-if="error" class="error-state">
      {{ error }}
    </div>

    <div v-else-if="insights.length === 0" class="empty-state">
      No insights yet. AI Sentinel is monitoring your system.
    </div>

    <div v-else class="insights-groups">
      <template v-for="(group, type) in groupedInsights" :key="type">
        <div v-if="group.length > 0" class="insight-group">
          <h3 class="group-title">{{ type.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase()) }}</h3>
          <div class="group-items">
            <TestingPromotionCard
              v-if="type === 'shadow_promotion' || type === 'testing_promotion'"
              v-for="insight in group"
              :key="insight.id"
              :insight="insight as any"
              @promote="() => handleAction(insight, 'promote')"
              @dismiss="() => handleAction(insight, 'dismiss')"
              @askAI="handleAskAI(insight)"
            />
            <AnomalyCard
              v-else-if="type === 'anomaly'"
              v-for="insight in group"
              :key="insight.id"
              :insight="insight as any"
              @investigate="() => handleAction(insight, 'investigate')"
              @dismiss="() => handleAction(insight, 'dismiss')"
              @askAI="handleAskAI(insight)"
            />
            <OptimizationCard
              v-else-if="type === 'optimization'"
              v-for="insight in group"
              :key="insight.id"
              :insight="insight as any"
              @apply="() => handleAction(insight, 'apply')"
              @dismiss="() => handleAction(insight, 'dismiss')"
              @askAI="handleAskAI(insight)"
            />
            <div v-else class="insight-with-deep-ask">
              <InsightCard
                v-for="insight in group"
                :key="insight.id"
                :insight="insight"
                @action="(actionId) => handleAction(insight, actionId)"
                @askAI="handleAskAI(insight)"
              />
              <DeepAskAI
                v-for="insight in group"
                :key="`deep-${insight.id}`"
                :insight="insight"
              />
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.insights-feed {
  padding: 20px;
}

.loading-state,
.error-state,
.empty-state {
  padding: 40px;
  text-align: center;
  color: var(--text-muted);
}

.error-state {
  color: var(--status-critical);
}

.insights-groups {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.insight-group {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.group-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
  text-transform: capitalize;
}

.group-items {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 20px;
}

.insight-with-deep-ask {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
</style>

