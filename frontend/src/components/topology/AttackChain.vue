<script setup lang="ts">
import { computed } from 'vue'
import ProcessNode from './ProcessNode.vue'
import type { ProcessInfo, Alert } from '../../lib/api'

const props = defineProps<{
  ancestors: ProcessInfo[]
  alert: Alert | null
  loading?: boolean
}>()

// Prepare process data for display
const processChain = computed(() => {
  if (!props.ancestors.length) return []
  
  return props.ancestors.map((proc, index) => {
    const isLast = index === props.ancestors.length - 1
    const isAlertSource = isLast && !!props.alert
    return {
      ...proc,
      isAlertSource,
      severity: isAlertSource ? (props.alert?.severity as 'high' | 'warning' | 'info') : undefined
    }
  })
})
</script>

<template>
  <div class="attack-chain">
    <!-- Toolbar -->
    <div class="chain-toolbar">
      <div class="toolbar-title">
        <span class="title-text">Attack Chain</span>
        <span v-if="ancestors.length" class="title-count">{{ ancestors.length }} processes</span>
      </div>
    </div>

    <!-- Chain Canvas -->
    <div class="chain-canvas">
      <!-- Loading State -->
      <div v-if="loading" class="chain-loading">
        <div class="loading-spinner"></div>
        <span>Loading ancestry chain...</span>
      </div>

      <!-- Empty State -->
      <div v-else-if="!alert" class="chain-empty">
        <div class="empty-icon">🔍</div>
        <div class="empty-title">Select an Alert</div>
        <div class="empty-description">
          Click on an alert to view its process ancestry chain
        </div>
      </div>

      <!-- No Ancestors -->
      <div v-else-if="!ancestors.length" class="chain-empty">
        <div class="empty-icon">🌲</div>
        <div class="empty-title">No Ancestry Data</div>
        <div class="empty-description">
          Process ancestry information is not available for this alert
        </div>
      </div>

      <!-- Process Chain -->
      <div v-else class="chain-content">
        <div class="chain-wrapper">
          <div 
            v-for="(proc, index) in processChain" 
            :key="proc.pid"
            class="chain-node-wrapper"
          >
            <!-- Connector line from previous node -->
            <div v-if="index > 0" class="connector">
              <div class="connector-line" :class="{ 'is-alert': index === processChain.length - 1 }"></div>
              <div class="connector-arrow" :class="{ 'is-alert': index === processChain.length - 1 }"></div>
            </div>
            
            <!-- Process Node -->
            <ProcessNode 
              :data="{
                pid: proc.pid,
                ppid: proc.ppid,
                comm: proc.comm,
                timestamp: proc.timestamp,
                cgroupId: proc.cgroupId,
                isAlertSource: proc.isAlertSource,
                severity: proc.severity
              }"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- Alert Details Footer -->
    <div v-if="alert" class="chain-footer">
      <div class="footer-label">Alert:</div>
      <div class="footer-value">{{ alert.ruleName }}</div>
      <div class="footer-separator">|</div>
      <div class="footer-label">Severity:</div>
      <div class="footer-value" :class="`severity-${alert.severity}`">
        {{ alert.severity.toUpperCase() }}
      </div>
    </div>
  </div>
</template>

<style scoped>
.attack-chain {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--bg-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  overflow: hidden;
}

.chain-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-subtle);
  background: var(--bg-elevated);
  flex-shrink: 0;
}

.toolbar-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.title-text {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.title-count {
  font-size: 12px;
  color: var(--text-muted);
  font-family: var(--font-mono);
}

.chain-canvas {
  flex: 1;
  position: relative;
  overflow: auto;
  background: var(--bg-void);
}

/* Loading State */
.chain-loading {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  color: var(--text-muted);
}

.loading-spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border-subtle);
  border-top-color: var(--accent-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

/* Empty State */
.chain-empty {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 40px;
  text-align: center;
}

.empty-icon {
  font-size: 48px;
  opacity: 0.5;
}

.empty-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.empty-description {
  font-size: 14px;
  color: var(--text-muted);
  max-width: 280px;
}

/* Chain Content */
.chain-content {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  min-height: 100%;
  padding: 32px 24px;
}

.chain-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.chain-node-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
}

/* Connector between nodes */
.connector {
  display: flex;
  flex-direction: column;
  align-items: center;
  height: 40px;
}

.connector-line {
  width: 2px;
  flex: 1;
  background: var(--border-default);
}

.connector-line.is-alert {
  background: var(--status-warning);
}

.connector-arrow {
  width: 0;
  height: 0;
  border-left: 6px solid transparent;
  border-right: 6px solid transparent;
  border-top: 8px solid var(--border-default);
}

.connector-arrow.is-alert {
  border-top-color: var(--status-warning);
}

/* Footer */
.chain-footer {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border-top: 1px solid var(--border-subtle);
  background: var(--bg-elevated);
  font-size: 12px;
  flex-shrink: 0;
}

.footer-label {
  color: var(--text-muted);
}

.footer-value {
  font-family: var(--font-mono);
  color: var(--text-secondary);
}

.footer-value.severity-high { color: var(--status-critical); }
.footer-value.severity-warning { color: var(--status-warning); }
.footer-value.severity-info { color: var(--status-info); }

.footer-separator {
  color: var(--border-default);
  margin: 0 4px;
}
</style>
