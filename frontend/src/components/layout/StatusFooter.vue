<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { Boxes, Activity, Box } from 'lucide-vue-next'
import { getSystemStats } from '../../lib/api'

interface SystemStats {
  processCount: number
  workloadCount: number
  eventsPerSec: number
  alertCount: number
  probeStatus: string
  probeError?: string
}

const stats = ref<SystemStats>({
  processCount: 0,
  workloadCount: 0,
  eventsPerSec: 0,
  alertCount: 0,
  probeStatus: 'starting'
})

let pollInterval: number | null = null

const probeStatusLabel = computed(() => {
  switch (stats.value.probeStatus) {
    case 'active':
      return 'Active'
    case 'error':
      return 'Error'
    case 'stopped':
      return 'Stopped'
    case 'starting':
      return 'Starting'
    default:
      return stats.value.probeStatus || 'Unknown'
  }
})

const fetchStats = async () => {
  try {
    const result = await getSystemStats()
    stats.value = { ...stats.value, ...result }
  } catch (e) {
    console.error('Failed to fetch stats:', e)
  }
}

onMounted(() => {
  fetchStats()
  pollInterval = window.setInterval(fetchStats, 2000)
})

onUnmounted(() => {
  if (pollInterval) {
    clearInterval(pollInterval)
  }
})
</script>

<template>
  <footer class="status-footer">
    <div class="footer-left">
      <div class="footer-item" :title="stats.probeError || ''">
        <Activity :size="14" class="footer-icon" :class="stats.probeStatus" />
        <span class="footer-label">eBPF:</span>
        <span class="footer-value" :class="stats.probeStatus">
          {{ probeStatusLabel }}
        </span>
      </div>
      <div class="footer-divider"></div>
      <div class="footer-item">
        <Box :size="14" class="footer-icon" />
        <span class="footer-label">Processes:</span>
        <span class="footer-value">{{ stats.processCount }}</span>
      </div>
      <div class="footer-divider"></div>
      <div class="footer-item">
        <Boxes :size="14" class="footer-icon" />
        <span class="footer-label">Workloads:</span>
        <span class="footer-value">{{ stats.workloadCount }}</span>
      </div>
    </div>

  </footer>
</template>

<style scoped>
.status-footer {
  height: var(--footer-height);
  background: var(--bg-surface);
  border-top: 1px solid var(--border-subtle);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px; /* Consistent padding */
  font-size: 12px;
  color: var(--text-muted);
}

.footer-left {
  display: flex;
  align-items: center;
  gap: 16px; /* Increased gap for more breathing room */
}

.footer-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.footer-icon {
  color: var(--text-muted);
}

.footer-icon.active {
  color: var(--status-safe);
}

.footer-icon.error {
  color: var(--status-critical);
}

.footer-icon.stopped {
  color: var(--text-muted);
}

.footer-label {
  color: var(--text-muted);
}

.footer-value {
  color: var(--text-primary); /* More prominent value color */
  font-family: var(--font-mono);
  font-weight: 500;
}

.footer-value.active {
  color: var(--status-safe);
}

.footer-value.error {
  color: var(--status-critical);
}

.footer-value.stopped {
  color: var(--text-muted);
}

.footer-divider {
  width: 1px;
  height: 16px;
  background: var(--border-subtle);
}
</style>
