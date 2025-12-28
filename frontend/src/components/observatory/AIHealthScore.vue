<!-- AI Health Score Component - Phase 4 -->
<script setup lang="ts">
import { computed } from 'vue'
import { Heart, TrendingUp, TrendingDown, Minus } from 'lucide-vue-next'

const props = defineProps<{
  score: number // 0-100
  assessment?: string
  trend?: 'up' | 'down' | 'stable'
}>()

const scoreColor = computed(() => {
  if (props.score >= 80) return 'var(--status-safe)'
  if (props.score >= 60) return 'var(--status-warning)'
  return 'var(--status-critical)'
})

const scoreLabel = computed(() => {
  if (props.score >= 80) return 'Excellent'
  if (props.score >= 60) return 'Good'
  if (props.score >= 40) return 'Fair'
  return 'Poor'
})
</script>

<template>
  <div class="health-score">
    <div class="score-header">
      <Heart :size="24" :color="scoreColor" />
      <div class="score-display">
        <div class="score-value" :style="{ color: scoreColor }">
          {{ score }}/100
        </div>
        <div class="score-label">{{ scoreLabel }}</div>
      </div>
      <div v-if="trend" class="trend">
        <TrendingUp v-if="trend === 'up'" :size="16" color="var(--status-safe)" />
        <TrendingDown v-else-if="trend === 'down'" :size="16" color="var(--status-critical)" />
        <Minus v-else :size="16" :color="scoreColor" />
      </div>
    </div>
    <div v-if="assessment" class="assessment">
      <div class="assessment-label">AI Assessment:</div>
      <div class="assessment-text">{{ assessment }}</div>
    </div>
  </div>
</template>

<style scoped>
.health-score {
  padding: 20px;
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
}

.score-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}

.score-display {
  flex: 1;
}

.score-value {
  font-size: 32px;
  font-weight: 700;
  line-height: 1;
  margin-bottom: 4px;
}

.score-label {
  font-size: 14px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding-left: 3px;
}

.trend {
  display: flex;
  align-items: center;
}

.assessment {
  padding-top: 16px;
  border-top: 1px solid var(--border-subtle);
}

.assessment-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.assessment-text {
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-secondary);
}
</style>
