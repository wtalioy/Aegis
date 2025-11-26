<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { Brain, Play, CheckCircle } from 'lucide-vue-next'
import LearningTimer from '../components/profiler/LearningTimer.vue'
import LiveCapture from '../components/profiler/LiveCapture.vue'
import RuleReview from '../components/profiler/RuleReview.vue'
import { 
  getLearningStatus, 
  startLearning, 
  stopLearning, 
  applyWhitelistRules,
  type LearningStatus,
  type Rule
} from '../lib/api'

const router = useRouter()

type ProfilerState = 'idle' | 'learning' | 'reviewing'

const state = ref<ProfilerState>('idle')
const learningDuration = ref(300)
const learningStatus = ref<LearningStatus>({
  active: false,
  startTime: 0,
  duration: 0,
  patternCount: 0,
  execCount: 0,
  fileCount: 0,
  connectCount: 0,
  remainingSeconds: 0
})
const generatedRules = ref<Rule[]>([])
const applying = ref(false)
const error = ref<string | null>(null)
const success = ref<string | null>(null)

let statusPollInterval: number | null = null

const durationOptions = [
  { value: 30, label: '30 seconds' },
  { value: 60, label: '1 minute' },
  { value: 180, label: '3 minutes' },
  { value: 300, label: '5 minutes' },
  { value: 600, label: '10 minutes' },
]

const handleStartLearning = async () => {
  error.value = null
  try {
    await startLearning(learningDuration.value)
    state.value = 'learning'
    startStatusPolling()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to start learning'
  }
}

const handleStopLearning = async () => {
  error.value = null
  try {
    stopStatusPolling()
    const rules = await stopLearning()
    generatedRules.value = rules
    state.value = 'reviewing'
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to stop learning'
    state.value = 'idle'
  }
}

const handleApplyRules = async (selectedIndices: number[]) => {
  error.value = null
  success.value = null
  applying.value = true
  try {
    await applyWhitelistRules(selectedIndices)
    success.value = `Successfully merged ${selectedIndices.length} rules into rules.yaml`
    state.value = 'idle'
    generatedRules.value = []
    // Navigate to Rules page after a short delay so user sees the success message
    setTimeout(() => {
      router.push('/rules')
    }, 1500)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to apply rules'
  } finally {
    applying.value = false
  }
}

const handleCancel = () => {
  state.value = 'idle'
  generatedRules.value = []
}

const pollStatus = async () => {
  try {
    const status = await getLearningStatus()
    learningStatus.value = status
    
    if (!status.active && state.value === 'learning') {
      handleStopLearning()
    }
  } catch (e) {
    console.error('Failed to poll status:', e)
  }
}

const startStatusPolling = () => {
  pollStatus()
  statusPollInterval = window.setInterval(pollStatus, 1000)
}

const stopStatusPolling = () => {
  if (statusPollInterval !== null) {
    clearInterval(statusPollInterval)
    statusPollInterval = null
  }
}

onMounted(async () => {
  try {
    const status = await getLearningStatus()
    if (status.active) {
      learningStatus.value = status
      state.value = 'learning'
      startStatusPolling()
    }
  } catch (e) {
  }
})

onUnmounted(() => {
  stopStatusPolling()
})
</script>

<template>
  <div class="profiler-page">
    <!-- Header -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">
          <Brain :size="24" class="title-icon" />
          Behavior Profiler
        </h1>
        <span class="page-subtitle">Learn normal behavior and generate whitelist rules</span>
      </div>
      
      <!-- State Indicator -->
      <div class="state-indicator" :class="state">
        <div class="state-dot"></div>
        <span class="state-label">
          {{ state === 'idle' ? 'Ready' : state === 'learning' ? 'Learning' : 'Review' }}
        </span>
      </div>
    </div>

    <!-- Error Banner -->
    <div v-if="error" class="error-banner">
      {{ error }}
      <button @click="error = null">&times;</button>
    </div>

    <!-- Success Banner -->
    <div v-if="success" class="success-banner">
      <CheckCircle :size="16" />
      {{ success }}
      <button @click="success = null">&times;</button>
    </div>

    <!-- IDLE State -->
    <div v-if="state === 'idle'" class="idle-state">
      <div class="idle-card">
        <div class="idle-icon">
          <Play :size="48" />
        </div>
        <h2 class="idle-title">Start Learning Mode</h2>
        <p class="idle-description">
          EulerGuard will observe normal system behavior and generate whitelist rules 
          based on the patterns detected during the learning period.
        </p>

        <div class="duration-selector">
          <label class="duration-label">Learning Duration</label>
          <div class="duration-options">
            <button
              v-for="opt in durationOptions"
              :key="opt.value"
              class="duration-btn"
              :class="{ active: learningDuration === opt.value }"
              @click="learningDuration = opt.value"
            >
              {{ opt.label }}
            </button>
          </div>
        </div>

        <button class="start-btn" @click="handleStartLearning">
          <Play :size="18" />
          Start Learning
        </button>

        <div class="idle-tips">
          <h4>Tips for effective learning:</h4>
          <ul>
            <li>Run your typical workloads during the learning period</li>
            <li>Include common operations like builds, tests, and deployments</li>
            <li>Avoid running unusual or one-time tasks</li>
            <li>Longer learning periods capture more patterns</li>
          </ul>
        </div>
      </div>
    </div>

    <!-- LEARNING State -->
    <div v-if="state === 'learning'" class="learning-state">
      <div class="learning-main">
        <LearningTimer
          :remaining-seconds="learningStatus.remainingSeconds"
          :total-duration="learningStatus.duration"
          :pattern-count="learningStatus.patternCount"
          :exec-count="learningStatus.execCount"
          :file-count="learningStatus.fileCount"
          :connect-count="learningStatus.connectCount"
          @stop="handleStopLearning"
        />
      </div>
      
      <div class="learning-sidebar">
        <LiveCapture />
      </div>
    </div>

    <!-- REVIEWING State -->
    <div v-if="state === 'reviewing'" class="reviewing-state">
      <RuleReview
        :rules="generatedRules"
        @apply="handleApplyRules"
        @cancel="handleCancel"
      />
    </div>
  </div>
</template>

<style scoped>
.profiler-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* Header */
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 16px;
}

.header-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.title-icon {
  color: var(--status-learning);
}

.page-subtitle {
  font-size: 14px;
  color: var(--text-muted);
}

.state-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: var(--bg-elevated);
  border-radius: var(--radius-full);
  border: 1px solid var(--border-subtle);
}

.state-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--text-muted);
}

.state-indicator.idle .state-dot {
  background: var(--text-muted);
}

.state-indicator.learning .state-dot {
  background: var(--status-learning);
  animation: pulse-dot 1.5s ease-in-out infinite;
}

.state-indicator.reviewing .state-dot {
  background: var(--status-safe);
}

@keyframes pulse-dot {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.6; transform: scale(1.2); }
}

.state-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
}

/* Error Banner */
.error-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--status-critical-dim);
  border: 1px solid var(--status-critical);
  border-radius: var(--radius-md);
  color: var(--status-critical);
  font-size: 13px;
}

.error-banner button {
  background: none;
  border: none;
  color: var(--status-critical);
  font-size: 18px;
  cursor: pointer;
  padding: 0 4px;
}

/* Success Banner */
.success-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: var(--status-safe-dim);
  border: 1px solid var(--status-safe);
  border-radius: var(--radius-md);
  color: var(--status-safe);
  font-size: 13px;
}

.success-banner button {
  background: none;
  border: none;
  color: var(--status-safe);
  font-size: 18px;
  cursor: pointer;
  padding: 0 4px;
  margin-left: auto;
}

/* IDLE State */
.idle-state {
  display: flex;
  justify-content: center;
  padding: 40px 20px;
}

.idle-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  max-width: 500px;
  padding: 48px;
  background: var(--bg-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  text-align: center;
}

.idle-icon {
  width: 80px;
  height: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--status-learning-dim);
  border-radius: 50%;
  color: var(--status-learning);
  margin-bottom: 24px;
}

.idle-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 12px 0;
}

.idle-description {
  font-size: 14px;
  color: var(--text-muted);
  line-height: 1.6;
  margin: 0 0 32px 0;
}

.duration-selector {
  width: 100%;
  margin-bottom: 32px;
}

.duration-label {
  display: block;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 12px;
}

.duration-options {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: center;
}

.duration-btn {
  padding: 8px 16px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
  font-size: 12px;
  color: var(--text-secondary);
  transition: all var(--transition-fast);
}

.duration-btn:hover {
  background: var(--bg-overlay);
  color: var(--text-primary);
}

.duration-btn.active {
  background: var(--status-learning-dim);
  color: var(--status-learning);
  border: 1px solid var(--status-learning);
}

.start-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 32px;
  background: var(--status-learning);
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 500;
  color: white;
  transition: all var(--transition-fast);
  margin-bottom: 32px;
}

.start-btn:hover {
  filter: brightness(1.1);
}

.idle-tips {
  width: 100%;
  padding: 20px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
  text-align: left;
}

.idle-tips h4 {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  margin: 0 0 12px 0;
}

.idle-tips ul {
  margin: 0;
  padding-left: 20px;
}

.idle-tips li {
  font-size: 12px;
  color: var(--text-muted);
  line-height: 1.8;
}

/* LEARNING State */
.learning-state {
  display: grid;
  grid-template-columns: 1fr 350px;
  gap: 24px;
}

@media (max-width: 900px) {
  .learning-state {
    grid-template-columns: 1fr;
  }
}

.learning-main {
  display: flex;
  justify-content: center;
  align-items: flex-start;
}

.learning-sidebar {
  display: flex;
  flex-direction: column;
}

/* REVIEWING State */
.reviewing-state {
  flex: 1;
}
</style>
