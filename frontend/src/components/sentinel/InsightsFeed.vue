<!-- Insights Feed Component - Phase 4 -->
<script setup lang="ts">
import { ref, computed } from 'vue'
import { useSentinel } from '../../composables/useSentinel'
import type { Insight, InsightType } from '../../types/sentinel'
import InsightCard from './InsightCard.vue'
import DeepAskAI from './DeepAskAI.vue'

const isAskAIModalOpen = ref(false)
const selectedInsightForAI = ref<Insight | null>(null)

const { insights, loading, error, executeAction } = useSentinel()

const groupedInsights = computed(() => {
  const groups: Record<InsightType, Insight[]> = {
    testingPromotion: [],
    anomaly: [],
    optimization: [],
    dailyReport: []
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
  selectedInsightForAI.value = insight
  isAskAIModalOpen.value = true
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
          <h3 class="group-title">{{ type.replace(/([A-Z])/g, ' $1').replace(/\b\w/g, l => l.toUpperCase()) }}</h3>
          <div class="group-items">
            <InsightCard
              v-for="insight in group"
              :key="insight.id"
              :insight="insight"
              @action="(actionId) => handleAction(insight, actionId)"
              @askAI="handleAskAI(insight)"
            />
          </div>
        </div>
      </template>
    </div>

    <DeepAskAI
      v-if="isAskAIModalOpen && selectedInsightForAI"
      :insight="selectedInsightForAI"
      @close="isAskAIModalOpen = false"
    />
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
