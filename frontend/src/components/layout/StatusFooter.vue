<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { Cpu, HardDrive, Activity, Box } from 'lucide-vue-next'
import { getSystemStats } from '../../lib/api'

interface SystemStats {
  processCount: number
  containerCount: number
  eventsPerSec: number
  alertCount: number
  probeStatus: string
  cpuPercent: number
  memoryMB: number
}

const stats = ref<SystemStats>({
  processCount: 0,
  containerCount: 0,
  eventsPerSec: 0,
  alertCount: 0,
  probeStatus: 'starting',
  cpuPercent: 0,
  memoryMB: 0
})

let pollInterval: number | null = null

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
      <div class="footer-item">
        <Activity :size="14" class="footer-icon active" />
        <span class="footer-label">eBPF:</span>
        <span class="footer-value" :class="stats.probeStatus">
          {{ stats.probeStatus === 'active' ? 'Active' : stats.probeStatus }}
        </span>
      </div>
      <div class="footer-divider"></div>
      <div class="footer-item">
        <Box :size="14" class="footer-icon" />
        <span class="footer-label">Processes:</span>
        <span class="footer-value">{{ stats.processCount.toLocaleString() }}</span>
      </div>
      <div class="footer-divider"></div>
      <div class="footer-item">
        <HardDrive :size="14" class="footer-icon" />
        <span class="footer-label">Containers:</span>
        <span class="footer-value">{{ stats.containerCount }}</span>
      </div>
    </div>

    <div class="footer-right">
      <div class="footer-item">
        <Cpu :size="14" class="footer-icon" />
        <span class="footer-label">CPU:</span>
        <span class="footer-value">{{ stats.cpuPercent.toFixed(1) }}%</span>
      </div>
      <div class="footer-divider"></div>
      <div class="footer-item">
        <span class="footer-label">Mem:</span>
        <span class="footer-value">{{ stats.memoryMB.toFixed(0) }}MB</span>
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
  padding: 0 16px;
  font-size: 12px;
}

.footer-left,
.footer-right {
  display: flex;
  align-items: center;
  gap: 12px;
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

.footer-label {
  color: var(--text-muted);
}

.footer-value {
  color: var(--text-secondary);
  font-family: var(--font-mono);
}

.footer-value.active {
  color: var(--status-safe);
}

.footer-value.error {
  color: var(--status-critical);
}

.footer-divider {
  width: 1px;
  height: 16px;
  background: var(--border-subtle);
}
</style>

