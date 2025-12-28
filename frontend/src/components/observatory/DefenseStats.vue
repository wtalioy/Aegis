<!-- Defense Stats Component - Phase 4 -->
<script setup lang="ts">
import { computed } from 'vue'
import { Lock, ShieldCheck, ShieldOff, TrendingUp, Ban, Bell, Zap, Eye } from 'lucide-vue-next'

const props = defineProps<{
  stats: {
    totalBlocks: number
    totalAlerts: number
    activeRules: number
    testingRules: number
    defenseRate?: number
  }
  trend?: 'up' | 'down' | 'stable'
}>()

const defenseRate = computed(() => props.stats.defenseRate || 0)
const defenseColor = computed(() => {
  if (defenseRate.value >= 90) return 'var(--status-safe)'
  if (defenseRate.value >= 70) return 'var(--status-warning)'
  return 'var(--status-critical)'
})
</script>

<template>
  <div class="defense-stats">
    <div class="stats-header">
      <Lock :size="20" />
      <h3>Defense Statistics</h3>
    </div>

    <div class="defense-rate">
      <div class="rate-display">
        <div class="rate-value" :style="{ color: defenseColor }">
          {{ Math.round(defenseRate) }}%
        </div>
        <div class="rate-label">Defense Rate</div>
      </div>
      <div v-if="trend" class="trend-indicator">
        <TrendingUp v-if="trend === 'up'" :size="20" :color="defenseColor" />
        <ShieldCheck v-else-if="trend === 'stable'" :size="20" :color="defenseColor" />
        <ShieldOff v-else :size="20" :color="defenseColor" />
      </div>
    </div>

    <div class="stats-grid">
      <div class="stat-item">
        <div class="stat-icon blocks-icon">
          <Ban :size="20" />
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.totalBlocks }}</div>
          <div class="stat-label">Blocks</div>
        </div>
      </div>

      <div class="stat-item">
        <div class="stat-icon alerts-icon">
          <Bell :size="20" />
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.totalAlerts }}</div>
          <div class="stat-label">Alerts</div>
        </div>
      </div>

      <div class="stat-item">
        <div class="stat-icon active-icon">
          <Zap :size="20" />
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.activeRules }}</div>
          <div class="stat-label">Active Rules</div>
        </div>
      </div>

      <div class="stat-item">
        <div class="stat-icon testing-icon">
          <Eye :size="20" />
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.testingRules }}</div>
          <div class="stat-label">Testing Rules</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.defense-stats {
  padding: 20px;
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  display: flex;
  flex-direction: column;
  height: 100%;
}

.stats-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.stats-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.defense-rate {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
  padding: 20px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
}

.rate-display {
  display: flex;
  flex-direction: column;
}

.rate-value {
  font-size: 36px;
  font-weight: 700;
  line-height: 1;
  margin-bottom: 4px;
}

.rate-label {
  font-size: 13px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.trend-indicator {
  display: flex;
  align-items: center;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
}

.stat-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  background: var(--bg-void);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  transition: all 0.2s ease;
}

.stat-icon.blocks-icon {
  background: rgba(239, 68, 68, 0.15);
  color: rgb(239, 68, 68);
}

.stat-icon.alerts-icon {
  background: rgba(251, 191, 36, 0.15);
  color: rgb(251, 191, 36);
}

.stat-icon.active-icon {
  background: rgba(34, 197, 94, 0.15);
  color: rgb(34, 197, 94);
}

.stat-icon.testing-icon {
  background: rgba(139, 92, 246, 0.15);
  color: rgb(139, 92, 246);
}

.stat-item:hover .stat-icon {
  transform: scale(1.1);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
</style>

