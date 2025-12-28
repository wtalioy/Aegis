<!-- Sentinel Page - Phase 4: AI Active Insights -->
<script setup lang="ts">
import { computed } from 'vue'
import { Sparkles, AlertTriangle } from 'lucide-vue-next'
import { useSentinel, type Insight } from '../composables/useSentinel'
import { useOmnibox } from '../composables/useOmnibox'
import InsightCard from '../components/sentinel/InsightCard.vue'

const { insights, loading, error, connected, executeAction } = useSentinel()
const { openWithQuery } = useOmnibox()

const sortedInsights = computed(() => {
  return [...insights.value].sort((a, b) => {
    // Sort by severity (critical first) then by time (newest first)
    const severityOrder = { critical: 4, high: 3, medium: 2, low: 1 }
    const severityDiff = (severityOrder[b.severity] || 0) - (severityOrder[a.severity] || 0)
    if (severityDiff !== 0) return severityDiff
    return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
  })
})

const handleAction = (insight: Insight, actionId: string) => {
  executeAction(insight, actionId)
}

const handleAskAI = (insight: Insight) => {
  openWithQuery(`Tell me more about this insight: ${insight.title}`)
}
</script>

<template>
  <div class="sentinel-page">
    <div class="page-header">
      <div class="header-content">
        <div class="header-title">
          <Sparkles :size="24" class="header-icon" />
          <div>
            <h1>AI Sentinel</h1>
            <p class="header-subtitle">Active security insights powered by AI</p>
          </div>
        </div>
      </div>
      <div class="header-status">
        <div class="status-indicator" :class="{ connected }">
          <span class="status-dot"></span>
          <span>{{ connected ? 'Connected' : 'Disconnected' }}</span>
        </div>
      </div>
    </div>

    <div v-if="loading && insights.length === 0" class="loading-state">
      <div class="spinner"></div>
      <p>Loading insights...</p>
    </div>

    <div v-else-if="error" class="error-state">
      <AlertTriangle :size="24" />
      <p>{{ error }}</p>
    </div>

    <div v-else-if="insights.length === 0" class="empty-state">
      <Sparkles :size="48" class="empty-icon" />
      <h2>No insights yet</h2>
      <p>AI Sentinel is monitoring your system. Insights will appear here as they are discovered.</p>
    </div>

    <div v-else class="insights-container">
      <div class="insights-header">
        <h2 class="insights-title">Insights ({{ insights.length }})</h2>
      </div>
      <div class="insights-grid">
        <InsightCard
          v-for="insight in sortedInsights"
          :key="insight.id"
          :insight="insight"
          @action="(actionId) => handleAction(insight, actionId)"
          @askAI="() => handleAskAI(insight)"
        />
      </div>
    </div>
  </div>
</template>


<style scoped>
.sentinel-page {
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 32px;
  padding-bottom: 24px;
  border-bottom: 1px solid var(--border-subtle);
}

.header-content {
  flex: 1;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-icon {
  color: var(--accent-primary);
  flex-shrink: 0;
}

.header-title h1 {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 4px 0;
}

.header-subtitle {
  font-size: 14px;
  color: var(--text-secondary);
  margin: 0;
}

.header-status {
  display: flex;
  align-items: center;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  font-size: 13px;
  color: var(--text-secondary);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--text-muted);
  transition: all 0.2s;
}

.status-indicator.connected .status-dot {
  background: var(--status-safe);
  box-shadow: 0 0 8px var(--status-safe);
}

.loading-state,
.error-state,
.empty-state {
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
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 16px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-icon {
  color: var(--accent-primary);
  opacity: 0.5;
  margin-bottom: 16px;
}

.empty-state h2 {
  font-size: 20px;
  color: var(--text-primary);
  margin: 0 0 8px 0;
}

.empty-state p {
  color: var(--text-secondary);
  max-width: 400px;
  margin: 0;
}

.error-state {
  color: var(--status-critical);
}

.error-state svg {
  margin-bottom: 12px;
}

.insights-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.insights-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.insights-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.insights-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: 20px;
}

@media (max-width: 1200px) {
  .insights-grid {
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  }
}

@media (max-width: 768px) {
  .insights-grid {
    grid-template-columns: 1fr;
  }
  
  .page-header {
    flex-direction: column;
    gap: 16px;
  }
  
  .header-status {
    align-self: flex-start;
  }
}
</style>
