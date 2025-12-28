<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { CheckCircle2, AlertTriangle } from 'lucide-vue-next'

const props = defineProps<{ validationData: any }>()

// Default values (from config defaults)
const promotionMinObservationMinutes = ref(1440) // 24 hours in minutes
const promotionMinHits = ref(100)

// Fetch config from API
onMounted(async () => {
  try {
    const response = await fetch('/api/settings')
    if (response.ok) {
      const settings = await response.json()
      if (settings.promotion?.minObservationMinutes) {
        promotionMinObservationMinutes.value = settings.promotion.minObservationMinutes
      }
      if (settings.promotion?.minHits) {
        promotionMinHits.value = settings.promotion.minHits
      }
    }
  } catch (e) {
    // Use default if API fails
    console.warn('Failed to fetch promotion config, using default:', e)
  }
})

const observationMinutes = computed(() => props.validationData?.stats?.observationMinutes ?? props.validationData?.stats?.ObservationMinutes ?? 0)
const hitCount = computed(() => props.validationData?.stats?.hits ?? props.validationData?.stats?.matchCount ?? props.validationData?.stats?.matches ?? 0)

// Progress based on promotion minimum observation time
const obsPercent = computed(() => Math.min(100, (observationMinutes.value / promotionMinObservationMinutes.value) * 100))
const hitPercent = computed(() => Math.min(100, (hitCount.value / promotionMinHits.value) * 100))

// Promotion readiness uses configurable thresholds
const isReady = computed(() => (
  observationMinutes.value >= promotionMinObservationMinutes.value &&
  hitCount.value >= promotionMinHits.value
))

const unmetCriteria = computed(() => {
  const items: string[] = []
  if (observationMinutes.value < promotionMinObservationMinutes.value) {
    const minutesNeeded = promotionMinObservationMinutes.value - observationMinutes.value
    const hoursNeeded = (minutesNeeded / 60).toFixed(1)
    items.push(`${hoursNeeded}h more observation`)
  }
  if (hitCount.value < promotionMinHits.value) {
    items.push(`${promotionMinHits.value - hitCount.value} more hits`)
  }
  return items
})

// Format observation time display
const observationDisplay = computed(() => {
  if (observationMinutes.value < 60) {
    return `${observationMinutes.value}min`
  }
  const hours = (observationMinutes.value / 60).toFixed(1)
  return `${hours}h`
})

// Format observation time for minimum threshold display
// Always show in minutes for consistency
const formatObservationTime = (minutes: number) => {
  return `${minutes}min`
}
</script>

<template>
  <div class="readiness">
    <div class="header">
      <h3>Readiness Checklist</h3>
      <div class="badge" :class="{ ready: isReady }">
        <span>{{ isReady ? 'Ready' : 'Not Ready' }}</span>
      </div>
    </div>

    <div class="item" :class="{ met: observationMinutes >= promotionMinObservationMinutes }">
      <div class="icon">
        <CheckCircle2 v-if="observationMinutes >= promotionMinObservationMinutes" :size="16" />
        <AlertTriangle v-else :size="16" />
      </div>
      <div class="content">
        <div class="label">Observation Time</div>
        <div class="progress">
          <div class="fill" :style="{ width: obsPercent + '%' }" />
        </div>
        <div class="detail">{{ observationDisplay }} / {{ formatObservationTime(promotionMinObservationMinutes) }}
          minimum</div>
      </div>
    </div>

    <div class="item" :class="{ met: hitCount >= promotionMinHits }">
      <div class="icon">
        <CheckCircle2 v-if="hitCount >= promotionMinHits" :size="16" />
        <AlertTriangle v-else :size="16" />
      </div>
      <div class="content">
        <div class="label">Hit Count</div>
        <div class="progress">
          <div class="fill" :style="{ width: hitPercent + '%' }" />
        </div>
        <div class="detail">{{ hitCount }} / {{ promotionMinHits }} minimum</div>
      </div>
    </div>

    <div class="footer" :class="{ ready: isReady }">
      <span v-if="isReady">All criteria met. Safe to promote.</span>
      <span v-else>Needs: {{ unmetCriteria.join(', ') }}</span>
    </div>
  </div>
</template>

<style scoped>
.readiness {
  padding: 16px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
}

.badge {
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  background: rgba(251, 191, 36, 0.1);
  color: rgb(251, 191, 36);
  font-size: 12px;
  font-weight: 600;
}

.badge.ready {
  background: rgba(34, 197, 94, 0.1);
  color: rgb(34, 197, 94);
}

.item {
  display: flex;
  gap: 10px;
  padding: 10px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-surface);
}

.item.met {
  border-color: rgba(34, 197, 94, 0.3);
  background: rgba(34, 197, 94, 0.05);
}

.icon {
  color: currentColor;
  display: flex;
  align-items: center;
}

.content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.progress {
  height: 6px;
  background: var(--bg-elevated);
  border-radius: 3px;
  overflow: hidden;
}

.fill {
  height: 100%;
  background: rgb(34, 197, 94);
  transition: width 0.3s ease;
}

.detail {
  font-size: 12px;
  color: var(--text-secondary);
}

.footer {
  padding-top: 4px;
  font-size: 12px;
  color: var(--text-secondary);
}

.footer.ready {
  color: rgb(34, 197, 94);
}
</style>
