<script setup lang="ts">
import { computed } from 'vue'
import { Clock, StopCircle, Terminal, FileText, Globe } from 'lucide-vue-next'

const props = defineProps<{
  remainingSeconds: number
  totalDuration: number
  patternCount: number
  execCount?: number
  fileCount?: number
  connectCount?: number
}>()

defineEmits<{
  stop: []
}>()

const progress = computed(() => {
  if (props.totalDuration === 0) return 0
  const elapsed = props.totalDuration - props.remainingSeconds
  return Math.min((elapsed / props.totalDuration) * 100, 100)
})

const formatTime = (seconds: number) => {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`
}
</script>

<template>
  <div class="learning-timer">
    <div class="timer-display">
      <div class="timer-icon">
        <Clock :size="32" class="pulsing" />
      </div>
      <div class="timer-value font-mono">{{ formatTime(remainingSeconds) }}</div>
      <div class="timer-label">remaining</div>
    </div>

    <div class="progress-container">
      <div class="progress-bar">
        <div class="progress-fill" :style="{ width: `${progress}%` }"></div>
      </div>
      <div class="progress-label">{{ Math.round(progress) }}% complete</div>
    </div>

    <div class="pattern-counter">
      <div class="counter-value font-mono">{{ patternCount }}</div>
      <div class="counter-label">patterns captured</div>
    </div>

    <!-- Breakdown by type -->
    <div class="breakdown-grid">
      <div class="breakdown-item exec">
        <Terminal :size="14" />
        <span class="breakdown-value font-mono">{{ execCount ?? 0 }}</span>
        <span class="breakdown-label">Exec</span>
      </div>
      <div class="breakdown-item file">
        <FileText :size="14" />
        <span class="breakdown-value font-mono">{{ fileCount ?? 0 }}</span>
        <span class="breakdown-label">File</span>
      </div>
      <div class="breakdown-item connect">
        <Globe :size="14" />
        <span class="breakdown-value font-mono">{{ connectCount ?? 0 }}</span>
        <span class="breakdown-label">Network</span>
      </div>
    </div>

    <button class="stop-btn" @click="$emit('stop')">
      <StopCircle :size="18" />
      Stop Learning
    </button>
  </div>
</template>

<style scoped>
.learning-timer {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px;
  background: var(--bg-elevated);
  border-radius: var(--radius-lg);
  border: 1px solid var(--status-learning);
  box-shadow: 0 0 40px rgba(139, 92, 246, 0.1);
}

.timer-display {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 32px;
}

.timer-icon {
  color: var(--status-learning);
  margin-bottom: 16px;
}

.timer-icon .pulsing {
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.7; transform: scale(1.1); }
}

.timer-value {
  font-size: 48px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1;
}

.timer-label {
  font-size: 14px;
  color: var(--text-muted);
  margin-top: 8px;
}

.progress-container {
  width: 100%;
  max-width: 300px;
  margin-bottom: 32px;
}

.progress-bar {
  height: 8px;
  background: var(--bg-void);
  border-radius: var(--radius-full);
  overflow: hidden;
  margin-bottom: 8px;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--status-learning), var(--accent-primary));
  border-radius: var(--radius-full);
  transition: width 1s ease;
}

.progress-label {
  font-size: 12px;
  color: var(--text-muted);
  text-align: center;
}

.pattern-counter {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 32px;
  padding: 16px 32px;
  background: var(--bg-surface);
  border-radius: var(--radius-md);
}

.counter-value {
  font-size: 32px;
  font-weight: 700;
  color: var(--status-learning);
}

.counter-label {
  font-size: 12px;
  color: var(--text-muted);
}

.stop-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  background: var(--status-critical-dim);
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 500;
  color: var(--status-critical);
  transition: all var(--transition-fast);
}

.stop-btn:hover {
  background: var(--status-critical);
  color: white;
}

/* Breakdown Grid */
.breakdown-grid {
  display: flex;
  gap: 16px;
  margin-bottom: 32px;
}

.breakdown-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 12px 20px;
  background: var(--bg-surface);
  border-radius: var(--radius-md);
  min-width: 80px;
}

.breakdown-item.exec { color: var(--status-info); }
.breakdown-item.file { color: var(--status-safe); }
.breakdown-item.connect { color: var(--status-warning); }

.breakdown-value {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
}

.breakdown-label {
  font-size: 10px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
</style>

