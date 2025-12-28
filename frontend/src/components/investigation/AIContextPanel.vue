<!-- AI Context Panel - user-driven mode (no auto analysis) -->
<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { MessageSquare, Loader2 } from 'lucide-vue-next'
import { useAI } from '../../composables/useAI'
import AIExplanation from '../ai/AIExplanation.vue'

const props = defineProps<{ event?: any; processId?: number }>()

const { explainEvent, loading } = useAI()
const explanation = ref<any>(null)
const expError = ref<string | null>(null)
const expSeq = ref(0)

// Compute details text based on event type (matching EventList logic)
const eventDetails = computed(() => {
  if (!props.event) return '—'
  const event = props.event as any
  
  if (event.type === 'exec') {
    return event.commandLine || event.filename || event.header?.comm || '—'
  } else if (event.type === 'file') {
    return event.filename || '—'
  } else if (event.type === 'connect') {
    if (event.addr && event.port) {
      return `${event.addr}:${event.port}`
    } else if (event.addr) {
      return event.addr
    }
    return '—'
  }
  return '—'
})

// Clear prior output when selection changes; do NOT auto analyze
watch(() => props.event, () => {
  explanation.value = null
  expError.value = null
})

const askDetail = async () => {
  if (!props.event) return
  const my = ++expSeq.value
  explanation.value = null
  expError.value = null
  try {
    const res = await explainEvent({
      eventId: props.event.id,
      eventData: props.event,
      question: 'Explain this event in detail: what happened, why it was flagged, and key evidence. Use structured markdown with headings and bullet points.'
    })
    if (my !== expSeq.value) return
    explanation.value = res
  } catch (e: any) {
    if (my !== expSeq.value) return
    expError.value = e?.message || 'Failed to get AI explanation'
  }
}

const askAction = async () => {
  if (!props.event) return
  const my = ++expSeq.value
  explanation.value = null
  expError.value = null
  try {
    const res = await explainEvent({
      eventId: props.event.id,
      eventData: props.event,
      question: 'What should I do? Provide concrete, prioritized containment and remediation steps, plus follow-up investigation tasks. Return a concise, ordered markdown list.'
    })
    if (my !== expSeq.value) return
    explanation.value = res
  } catch (e: any) {
    if (my !== expSeq.value) return
    expError.value = e?.message || 'Failed to get AI recommendation'
  }
}
</script>

<template>
  <div class="context-panel">
    <div class="panel-header">
      <MessageSquare :size="18" />
      <h3>AI Context</h3>
    </div>

    <div v-if="loading" class="loading">
      <Loader2 :size="16" class="spin" />
      <span>Working…</span>
    </div>

    <div v-else class="panel-content">
      <div v-if="!event" class="empty-state">
        <div class="empty-card">
          <div class="title">No event selected</div>
          <div class="hint">Pick an event from the table on the left to view AI context.</div>
          <ul class="tips">
            <li>Click a row to preview details.</li>
            <li>Use the buttons to ask for an explanation or next steps.</li>
          </ul>
        </div>
      </div>

      <template v-else>
        <!-- Always-visible compact event detail -->
        <div class="event-mini">
          <div class="row"><span class="label">Type</span><span class="value type">{{ event.type?.toUpperCase?.() || event.type }}</span></div>
          <div class="row"><span class="label">Process</span><span class="value process">{{ event.header?.comm || 'Unknown' }}</span></div>
          <div class="row"><span class="label">Details</span><span class="value details">{{ eventDetails }}</span></div>
          <div class="row"><span class="label">PID</span><span class="value mono">{{ event.header?.pid ?? '—' }}</span></div>
          <div class="row"><span class="label">Time</span><span class="value mono">{{ event.header?.timestamp ? new Date(event.header.timestamp).toLocaleTimeString('en-US',{hour12:false,hour:'2-digit',minute:'2-digit',second:'2-digit'}) : 'Unknown' }}</span></div>
        </div>

        <!-- Actions -->
        <div class="quick-actions">
          <button class="quick-ask-btn" @click="askDetail" title="Explain this in detail"><span>Explain this in detail</span></button>
          <button class="quick-ask-btn" @click="askAction" title="What should I do?"><span>What should I do?</span></button>
        </div>

        <!-- Error -->
        <div v-if="expError" class="error-box">
          <div class="error-text">{{ expError }}</div>
          <button class="retry-btn" @click="askDetail">Retry</button>
        </div>

        <!-- Explanation fills remaining space -->
        <div v-if="explanation && !expError" class="explanation-area">
          <AIExplanation :explanation="explanation.explanation" :root-cause="explanation.rootCause" :visible="true" expandable fit />
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.context-panel {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: 20px;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
  box-shadow: var(--shadow-md);
  backdrop-filter: blur(10px);
}

.panel-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-default);
  color: var(--accent-primary);
}

.panel-header svg {
  filter: drop-shadow(0 0 6px var(--accent-glow));
}

.panel-header h3 {
  font-size: 16px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
  letter-spacing: -0.3px;
}

.loading {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 48px 24px;
  justify-content: center;
  color: var(--text-muted);
  font-size: 14px;
  font-weight: 500;
}

.spin {
  animation: spin 1s linear infinite;
  color: var(--accent-primary);
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.panel-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 20px;
  overflow-y: auto;
  min-height: 0;
}

/* Empty state card */
.empty-state {
  padding: 24px;
  display: flex;
  justify-content: center;
  align-items: center;
  flex: 1;
}

.empty-card {
  width: 100%;
  background: linear-gradient(135deg, var(--bg-elevated) 0%, var(--bg-overlay) 100%);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: 32px 24px;
  text-align: center;
  color: var(--text-secondary);
  box-shadow: var(--shadow-sm);
}

.empty-card .icon {
  color: var(--accent-primary);
  margin-bottom: 12px;
  display: inline-flex;
  opacity: 0.7;
}

.empty-card .title {
  font-size: 16px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 8px;
  letter-spacing: -0.2px;
}

.empty-card .hint {
  font-size: 14px;
  margin-bottom: 12px;
  line-height: 1.5;
  color: var(--text-secondary);
}

.empty-card .tips {
  display: inline-block;
  text-align: left;
  margin: 16px auto 0;
  padding-left: 20px;
  color: var(--text-muted);
  font-size: 13px;
  line-height: 1.8;
}

.empty-card .tips li {
  margin-bottom: 4px;
}

/* Event mini detail */
.event-mini {
  display: grid;
  grid-template-columns: 1fr;
  gap: 10px;
  background: linear-gradient(135deg, var(--bg-elevated) 0%, rgba(26, 26, 36, 0.8) 100%);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: 16px;
  box-shadow: var(--shadow-sm);
  transition: all var(--transition-normal);
}

.event-mini:hover {
  border-color: var(--border-default);
  box-shadow: var(--shadow-md);
  transform: translateY(-1px);
}

.event-mini .row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 4px 0;
}

.event-mini .label {
  width: 80px;
  color: var(--text-muted);
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.6px;
  font-weight: 600;
  flex-shrink: 0;
}

.event-mini .value {
  color: var(--text-primary);
  font-size: 13px;
  flex: 1;
}

.event-mini .value.type {
  font-weight: 700;
  letter-spacing: 0.5px;
  color: var(--accent-primary);
  text-transform: uppercase;
  font-size: 12px;
}

.event-mini .value.process {
  font-weight: 600;
  color: var(--text-primary);
}

.event-mini .value.details {
  font-family: var(--font-mono);
  color: var(--text-secondary);
  font-size: 12px;
  word-break: break-all;
  line-height: 1.4;
}

.event-mini .value.mono {
  font-family: var(--font-mono);
  color: var(--text-secondary);
  font-size: 12px;
}

/* Quick actions */
.quick-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.quick-ask-btn {
  display: inline-flex;
  align-items: center;
  justify-content: flex-start;
  padding: 12px 16px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all var(--transition-normal);
  height: 48px;
  box-shadow: var(--shadow-sm);
  position: relative;
  overflow: hidden;
}

.quick-ask-btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(96, 165, 250, 0.1), transparent);
  transition: left var(--transition-slow);
}

.quick-ask-btn:hover {
  background: linear-gradient(135deg, var(--bg-overlay) 0%, var(--bg-elevated) 100%);
  color: var(--text-primary);
  border-color: var(--accent-primary);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(96, 165, 250, 0.2), var(--shadow-md);
}

.quick-ask-btn:hover::before {
  left: 100%;
}

.quick-ask-btn:active {
  transform: translateY(0);
  box-shadow: var(--shadow-sm);
}

/* Error box */
.error-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 16px;
  border: 1px solid var(--status-critical);
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.12) 0%, rgba(239, 68, 68, 0.06) 100%);
  color: var(--status-critical);
  border-radius: var(--radius-md);
  box-shadow: 0 0 0 1px rgba(239, 68, 68, 0.2);
}

.error-text {
  font-size: 13px;
  font-weight: 500;
  flex: 1;
}

.retry-btn {
  padding: 6px 14px;
  font-size: 12px;
  font-weight: 600;
  background: transparent;
  color: var(--status-critical);
  border: 1px solid var(--status-critical);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.retry-btn:hover {
  background: var(--status-critical);
  color: white;
  transform: scale(1.05);
}

.explanation-area {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
  padding: 4px 0;
}
</style>
