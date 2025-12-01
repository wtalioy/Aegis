<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Zap } from 'lucide-vue-next'
import { getAIStatus, type AIStatus } from '../../lib/api'
import DiagnosisModal from '../ai/DiagnosisModal.vue'

const probeStatus = ref<'active' | 'error' | 'starting'>('starting')

// AI state
const aiStatus = ref<AIStatus | null>(null)
const showDiagnosisModal = ref(false)

onMounted(async () => {
  setTimeout(() => {
    if (probeStatus.value === 'starting') {
      probeStatus.value = 'active'
    }
  }, 2000)
  
  // Fetch AI status
  try {
    aiStatus.value = await getAIStatus()
  } catch (e) {
    console.error('Failed to fetch AI status:', e)
  }
})
</script>

<template>
  <header class="topbar">
    <div class="topbar-left">
      <div class="status-indicator" :class="probeStatus">
        <span class="pulse-ring"></span>
        <span class="status-dot"></span>
        <span class="status-text">
          {{ probeStatus === 'active' ? 'Probes Active' : probeStatus === 'error' ? 'Probe Error' : 'Starting...' }}
        </span>
      </div>
    </div>

    <div class="topbar-center">
      <!-- Rate display removed -->
    </div>

    <div class="topbar-right">
      <!-- Quick Diagnose Button -->
      <button 
        v-if="aiStatus?.enabled"
        class="diagnose-btn" 
        @click="showDiagnosisModal = true"
        :disabled="aiStatus.status !== 'ready'"
        title="Quick one-click system diagnosis"
      >
        <Zap :size="15" />
        <span>Quick Diagnose</span>
      </button>
    </div>
  </header>
  
  <!-- Quick Diagnosis Modal -->
  <DiagnosisModal 
    :visible="showDiagnosisModal" 
    @close="showDiagnosisModal = false"
  />
</template>

<style scoped>
.topbar {
  height: var(--topbar-height);
  background: var(--bg-surface);
  border-bottom: 1px solid var(--border-subtle);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
}

.topbar-left,
.topbar-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.topbar-center {
  display: flex;
  align-items: center;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 10px;
  position: relative;
}

.pulse-ring {
  position: absolute;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  opacity: 0;
}

.status-indicator.active .pulse-ring {
  background: var(--status-safe);
  animation: pulse-ring 2s cubic-bezier(0.215, 0.61, 0.355, 1) infinite;
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--text-muted);
  position: relative;
  z-index: 1;
}

.status-indicator.active .status-dot {
  background: var(--status-safe);
  box-shadow: var(--glow-safe);
}

.status-indicator.error .status-dot {
  background: var(--status-critical);
  box-shadow: var(--glow-critical);
}

.status-indicator.starting .status-dot {
  background: var(--status-warning);
  animation: blink 1s ease-in-out infinite;
}

.status-text {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
}

.status-indicator.active .status-text {
  color: var(--status-safe);
}

.diagnose-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.diagnose-btn:hover:not(:disabled) {
  background: var(--bg-hover);
  color: var(--text-primary);
  border-color: var(--border-default);
}

.diagnose-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@keyframes pulse-ring {
  0% {
    transform: scale(0.5);
    opacity: 0.8;
  }

  100% {
    transform: scale(2);
    opacity: 0;
  }
}

@keyframes blink {

  0%,
  100% {
    opacity: 1;
  }

  50% {
    opacity: 0.4;
  }
}
</style>
