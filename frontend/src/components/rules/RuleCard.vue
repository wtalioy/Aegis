<script setup lang="ts">
import { ref, computed } from 'vue'
import { ChevronDown, ChevronRight, Terminal, Globe, FileText, Copy, Check, ShieldOff, AlertTriangle, ShieldCheck } from 'lucide-vue-next'
import type { Rule } from '../../lib/api'

const props = defineProps<{
  rule: Rule
}>()

const isExpanded = ref(false)
const copied = ref(false)

const typeIcon = computed(() => {
  switch (props.rule.type) {
    case 'exec': return Terminal
    case 'connect': return Globe
    case 'file': return FileText
    default: return Terminal
  }
})

const actionIcon = computed(() => {
  switch (props.rule.action) {
    case 'block': return ShieldOff
    case 'alert': return AlertTriangle
    case 'allow': return ShieldCheck
    default: return AlertTriangle
  }
})

const matchEntries = computed(() => {
  if (!props.rule.match) return []
  return Object.entries(props.rule.match).filter(([_, value]) => value)
})

const copyYaml = async () => {
  await navigator.clipboard.writeText(props.rule.yaml)
  copied.value = true
  setTimeout(() => { copied.value = false }, 2000)
}
</script>

<template>
  <div class="rule-card" :class="[`action-${rule.action}`, `severity-${rule.severity}`]">
    <!-- Header (always visible) -->
    <div class="rule-header" @click="isExpanded = !isExpanded">
      <button class="expand-btn">
        <ChevronDown v-if="isExpanded" :size="16" />
        <ChevronRight v-else :size="16" />
      </button>

      <div class="rule-type" :class="`type-${rule.type}`">
        <component :is="typeIcon" :size="14" />
      </div>

      <div class="rule-info">
        <span class="rule-name">{{ rule.name }}</span>
        <span class="rule-description">{{ rule.description }}</span>
      </div>

      <div class="rule-badges">
        <span class="severity-badge" :class="rule.severity">
          {{ rule.severity.toUpperCase() }}
        </span>
        <span class="action-badge" :class="rule.action">
          <component :is="actionIcon" :size="10" />
          {{ rule.action.toUpperCase() }}
        </span>
      </div>
    </div>

    <!-- Expanded Content -->
    <Transition name="expand">
      <div v-if="isExpanded" class="rule-details">
        <!-- Match Conditions -->
        <div class="detail-section">
          <h4 class="section-title">Match Conditions</h4>
          <div class="match-grid">
            <div v-for="[key, value] in matchEntries" :key="key" class="match-item">
              <span class="match-key">{{ key.replace(/_/g, ' ') }}</span>
              <span class="match-value font-mono">{{ value }}</span>
            </div>
            <div v-if="matchEntries.length === 0" class="match-empty">
              No specific match conditions
            </div>
          </div>
        </div>

        <!-- YAML Preview -->
        <div class="detail-section">
          <div class="yaml-header">
            <h4 class="section-title">YAML Definition</h4>
            <button class="copy-btn" @click.stop="copyYaml">
              <Check v-if="copied" :size="14" class="copied" />
              <Copy v-else :size="14" />
              {{ copied ? 'Copied!' : 'Copy' }}
            </button>
          </div>
          <pre class="yaml-preview"><code>{{ rule.yaml }}</code></pre>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.rule-card {
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
  overflow: hidden;
  transition: all var(--transition-fast);
}

.rule-card:hover {
  border-color: var(--border-default);
}

/* Action-based left border accent - removed per user preference */

/* Header */
.rule-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  cursor: pointer;
  transition: background var(--transition-fast);
}

.rule-header:hover {
  background: var(--bg-overlay);
}

.expand-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  color: var(--text-muted);
  flex-shrink: 0;
}

.rule-type {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--radius-md);
  flex-shrink: 0;
}

.rule-type.type-exec {
  background: rgba(96, 165, 250, 0.15);
  color: var(--chart-exec);
}

.rule-type.type-connect {
  background: rgba(245, 158, 11, 0.15);
  color: var(--chart-network);
}

.rule-type.type-file {
  background: rgba(16, 185, 129, 0.15);
  color: var(--chart-file);
}

.rule-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.rule-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rule-description {
  font-size: 12px;
  color: var(--text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rule-badges {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.severity-badge,
.action-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: var(--radius-sm);
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.3px;
}

/* Severity badges - all styled consistently */
.severity-badge.critical {
  background: var(--status-blocked-dim);
  color: var(--status-blocked);
}

.severity-badge.high {
  background: var(--status-critical-dim);
  color: var(--status-critical);
}

.severity-badge.warning {
  background: var(--status-warning-dim);
  color: var(--status-warning);
}

.severity-badge.info {
  background: var(--status-info-dim);
  color: var(--status-info);
}

/* Action badges - all styled consistently */
.action-badge.block {
  background: var(--status-blocked);
  color: #fff;
}

.action-badge.alert {
  background: var(--status-warning-dim);
  color: var(--status-warning);
}

.action-badge.allow {
  background: var(--status-safe-dim);
  color: var(--status-safe);
}

.action-badge.log {
  background: var(--bg-overlay);
  color: var(--text-secondary);
}

/* Expanded Details */
.rule-details {
  border-top: 1px solid var(--border-subtle);
  padding: 16px;
  background: var(--bg-surface);
}

.expand-enter-active,
.expand-leave-active {
  transition: all 0.2s ease;
}

.expand-enter-from,
.expand-leave-to {
  opacity: 0;
  max-height: 0;
  padding-top: 0;
  padding-bottom: 0;
}

.detail-section {
  margin-bottom: 16px;
}

.detail-section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin: 0 0 10px 0;
}

.match-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 8px;
}

.match-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: var(--bg-elevated);
  border-radius: var(--radius-sm);
  gap: 12px;
}

.match-key {
  font-size: 12px;
  color: var(--text-muted);
  text-transform: capitalize;
}

.match-value {
  font-size: 12px;
  color: var(--text-primary);
}

.match-empty {
  font-size: 12px;
  color: var(--text-muted);
  font-style: italic;
}

.yaml-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.yaml-header .section-title {
  margin: 0;
}

.copy-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: var(--radius-sm);
  font-size: 11px;
  color: var(--text-secondary);
  transition: all 0.2s ease;
  cursor: pointer;
  border: none;
  background: transparent;
}

.copy-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
  transform: translateY(-1px);
}

.copy-btn:active {
  transform: translateY(0);
}

.copy-btn .copied {
  color: var(--status-safe);
}

.yaml-preview {
  background: var(--bg-void);
  border-radius: var(--radius-md);
  padding: 12px 16px;
  overflow-x: auto;
  margin: 0;
}

.yaml-preview code {
  font-family: var(--font-mono);
  font-size: 12px;
  line-height: 1.5;
  color: var(--text-secondary);
  white-space: pre;
}
</style>
