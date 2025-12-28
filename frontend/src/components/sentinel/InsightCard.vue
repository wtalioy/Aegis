<script setup lang="ts">
import { computed } from 'vue'
import { marked } from 'marked'
import { Sparkles, CheckCircle2, AlertTriangle, TrendingUp, Lightbulb } from 'lucide-vue-next'
import type { Insight } from '../../composables/useSentinel'
import AIConfidenceBadge from '../ai/AIConfidenceBadge.vue'

// Configure marked
marked.setOptions({
  breaks: true,
  gfm: true
})

const props = defineProps<{
  insight: Insight
}>()

const emit = defineEmits<{
  action: [actionId: string]
  askAI: []
}>()

const severityColor = computed(() => {
  switch (props.insight.severity) {
    case 'critical':
      return 'var(--status-critical)'
    case 'high':
      return 'var(--status-high)'
    case 'medium':
      return 'var(--status-warning)'
    default:
      return 'var(--status-safe)'
  }
})

const typeIcon = computed(() => {
  switch (props.insight.type) {
    case 'testing_promotion':
      return CheckCircle2
    case 'anomaly':
      return AlertTriangle
    case 'optimization':
      return TrendingUp
    case 'daily_report':
      return Lightbulb
    default:
      return Sparkles
  }
})

const displayType = computed(() => String(props.insight.type).replace('_', ' '))

const formatTime = computed(() => {
  // Parse the timestamp - handle both ISO string and Unix timestamp
  let date: Date
  const created_at = props.insight.created_at

  if (typeof created_at === 'string') {
    // Try parsing as ISO string first
    date = new Date(created_at)
    // If invalid, try parsing as Unix timestamp (seconds)
    if (isNaN(date.getTime())) {
      const timestamp = parseInt(created_at, 10)
      if (!isNaN(timestamp)) {
        // If timestamp is in seconds (less than year 2000 in milliseconds), convert to milliseconds
        date = new Date(timestamp < 946684800000 ? timestamp * 1000 : timestamp)
      }
    }
  } else if (typeof created_at === 'number') {
    // Handle numeric timestamp
    date = new Date(created_at < 946684800000 ? created_at * 1000 : created_at)
  } else {
    // Fallback to current time if parsing fails
    date = new Date()
  }

  // Validate the date
  if (isNaN(date.getTime())) {
    date = new Date()
  }

  const now = new Date()
  const diffMs = now.getTime() - date.getTime()

  // Handle negative differences (future dates) - use current time
  if (diffMs < 0) {
    return 'Just now'
  }

  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`
  return date.toLocaleDateString()
})

const renderedSummary = computed(() => {
  try {
    return marked.parse(props.insight.summary, { breaks: true })
  } catch (err) {
    console.error('Failed to parse markdown:', err)
    return props.insight.summary
  }
})
</script>

<template>
  <div class="insight-card">
    <div class="card-header">
      <div class="header-left">
        <div class="type-icon-wrapper"
          :style="{ backgroundColor: severityColor + '20', borderColor: severityColor + '40' }">
          <component :is="typeIcon" :size="18" class="type-icon" :style="{ color: severityColor }" />
        </div>
        <div class="header-text">
          <div class="title-row">
            <h3 class="card-title">{{ insight.title }}</h3>
            <span class="severity-badge" :style="{ backgroundColor: severityColor + '20', color: severityColor }">
              {{ insight.severity }}
            </span>
          </div>
          <div class="card-meta">
            <span class="card-type">{{ displayType }}</span>
            <span class="card-separator">•</span>
            <span class="card-time">{{ formatTime }}</span>
          </div>
        </div>
      </div>
      <AIConfidenceBadge :confidence="insight.confidence" size="sm" />
    </div>

    <div class="card-body">
      <div class="card-summary markdown" v-html="renderedSummary"></div>
    </div>

    <div v-if="insight.actions && insight.actions.length > 0" class="card-actions">
      <button v-for="action in insight.actions" :key="action.action_id" class="action-btn" :class="{
        'primary': action.action_id === 'promote' || action.action_id === 'apply' || action.action_id === 'investigate',
        'secondary': action.action_id === 'dismiss'
      }" @click="$emit('action', action.action_id)">
        {{ action.label }}
      </button>
      <button class="action-btn ask-ai" @click="$emit('askAI')">
        <Sparkles :size="14" />
        Ask AI
      </button>
    </div>
  </div>
</template>

<style scoped>
.insight-card {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: 20px;
  transition: all 0.2s;
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 400px;
  max-height: 400px;
}

.insight-card:hover {
  border-color: var(--border-default);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.header-left {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.type-icon-wrapper {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  border-radius: var(--radius-md);
  border: 1px solid;
  display: flex;
  align-items: center;
  justify-content: center;
}

.type-icon {
  flex-shrink: 0;
}

.header-text {
  flex: 1;
  min-width: 0;
}

.title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
  line-height: 1.4;
  flex: 1;
}

.severity-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.card-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-muted);
}

.card-type {
  text-transform: capitalize;
}

.card-separator {
  opacity: 0.5;
}

.card-time {
  font-size: 12px;
}

.card-body {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
}

.card-summary {
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-secondary);
  margin: 0;
}

/* Custom scrollbar for card body */
.card-body::-webkit-scrollbar {
  width: 6px;
}

.card-body::-webkit-scrollbar-track {
  background: var(--bg-void);
  border-radius: 3px;
}

.card-body::-webkit-scrollbar-thumb {
  background: var(--border-subtle);
  border-radius: 3px;
}

.card-body::-webkit-scrollbar-thumb:hover {
  background: var(--border-default);
}

/* Markdown styling */
.card-summary.markdown :deep(h1),
.card-summary.markdown :deep(h2),
.card-summary.markdown :deep(h3),
.card-summary.markdown :deep(h4) {
  color: var(--text-primary);
  font-weight: 600;
  margin-top: 16px;
  margin-bottom: 8px;
}

.card-summary.markdown :deep(h1) {
  font-size: 18px;
}

.card-summary.markdown :deep(h2) {
  font-size: 16px;
}

.card-summary.markdown :deep(h3) {
  font-size: 15px;
}

.card-summary.markdown :deep(h4) {
  font-size: 14px;
}

.card-summary.markdown :deep(p) {
  margin: 0 0 12px 0;
  color: var(--text-secondary);
}

.card-summary.markdown :deep(ul),
.card-summary.markdown :deep(ol) {
  margin: 8px 0;
  padding-left: 20px;
}

.card-summary.markdown :deep(li) {
  margin: 4px 0;
  color: var(--text-secondary);
}

.card-summary.markdown :deep(code) {
  background: var(--bg-void);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  font-family: var(--font-mono);
  font-size: 12px;
  color: var(--accent-primary);
}

.card-summary.markdown :deep(pre) {
  background: var(--bg-void);
  padding: 12px;
  border-radius: var(--radius-md);
  overflow-x: auto;
  margin: 12px 0;
}

.card-summary.markdown :deep(pre code) {
  background: transparent;
  padding: 0;
  color: var(--text-primary);
}

.card-summary.markdown :deep(blockquote) {
  border-left: 3px solid var(--accent-primary);
  padding-left: 12px;
  margin: 12px 0;
  color: var(--text-muted);
  font-style: italic;
}

.card-summary.markdown :deep(strong) {
  color: var(--text-primary);
  font-weight: 600;
}

.card-summary.markdown :deep(em) {
  font-style: italic;
}

.card-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  padding-top: 12px;
  border-top: 1px solid var(--border-subtle);
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-elevated);
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
}

.action-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
  border-color: var(--border-default);
}

.action-btn.primary {
  background: var(--accent-primary);
  color: white;
  border-color: var(--accent-primary);
}

.action-btn.primary:hover {
  background: var(--accent-primary-hover);
  border-color: var(--accent-primary-hover);
}

.action-btn.secondary {
  background: transparent;
  color: var(--text-secondary);
}

.action-btn.ask-ai {
  background: transparent;
  border: 1px dashed var(--border-subtle);
  color: var(--text-secondary);
}

.action-btn.ask-ai:hover {
  border-color: var(--accent-primary);
  color: var(--accent-primary);
  background: var(--accent-glow);
}
</style>
