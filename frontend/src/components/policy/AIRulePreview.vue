<!-- AI Rule Preview Component - Phase 4 -->
<script setup lang="ts">
import { computed } from 'vue'
import { CheckCircle2, AlertTriangle, Target } from 'lucide-vue-next'
import AIConfidenceBadge from '../ai/AIConfidenceBadge.vue'

const props = defineProps<{
  rule: {
    rule: any
    yaml: string
    reasoning: string
    confidence: number
    warnings?: string[]
    simulation?: any
  }
}>()

const emit = defineEmits<{
  deployToTesting: []
}>()

const hasWarnings = computed(() => props.rule.warnings && props.rule.warnings.length > 0)
</script>

<template>
  <div class="rule-preview">
    <div class="preview-header">
      <div class="header-left">
        <CheckCircle2 v-if="rule.confidence > 0.8" :size="20" class="icon-success" />
        <AlertTriangle v-else :size="20" class="icon-warning" />
        <span class="preview-title">Generated Rule</span>
        <AIConfidenceBadge :confidence="rule.confidence" />
      </div>
    </div>

    <div v-if="hasWarnings" class="warnings">
      <div v-for="(warning, idx) in rule.warnings" :key="idx" class="warning-item">
        <AlertTriangle :size="14" />
        <span>{{ warning }}</span>
      </div>
    </div>

    <div class="reasoning">
      <div class="reasoning-title">AI Reasoning:</div>
      <div class="reasoning-text">{{ rule.reasoning }}</div>
    </div>

    <div class="yaml-preview">
      <div class="yaml-header">YAML:</div>
      <pre class="yaml-content">{{ rule.yaml }}</pre>
    </div>

    <div class="preview-actions">
      <button class="action-btn deploy-testing" @click="$emit('deployToTesting')">
        <Target :size="16" />
        <span>Deploy to Testing</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.rule-preview {
  margin-top: 24px;
  padding: 20px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
}

.preview-header {
  margin-bottom: 16px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.preview-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.icon-success {
  color: rgb(34, 197, 94);
}

.icon-warning {
  color: rgb(251, 191, 36);
}

.warnings {
  margin-bottom: 16px;
  padding: 12px;
  background: rgba(251, 191, 36, 0.1);
  border: 1px solid rgba(251, 191, 36, 0.2);
  border-radius: var(--radius-md);
}

.warning-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-secondary);
  margin-top: 6px;
}

.warning-item:first-child {
  margin-top: 0;
}

.reasoning {
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.reasoning-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.reasoning-text {
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-secondary);
}

.yaml-preview {
  margin-bottom: 16px;
}

.yaml-header {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.yaml-content {
  padding: 12px;
  background: var(--bg-void);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  font-size: 12px;
  font-family: 'Monaco', 'Menlo', monospace;
  color: var(--text-primary);
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
}

.simulation-preview {
  margin-bottom: 16px;
  padding: 12px;
  background: var(--bg-void);
  border-radius: var(--radius-md);
}

.simulation-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.simulation-stats {
  display: flex;
  gap: 24px;
}

.stat {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-label {
  font-size: 12px;
  color: var(--text-muted);
}

.stat-value {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}

.preview-actions {
  display: flex;
  gap: 12px;
  padding-top: 16px;
  border-top: 1px solid var(--border-subtle);
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  border-radius: var(--radius-md);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
  border: none;
}

.action-btn.simulate {
  background: rgba(168, 85, 247, 0.15);
  color: rgb(168, 85, 247);
  border: 1px solid rgba(168, 85, 247, 0.3);
}

.action-btn.simulate:hover {
  background: rgba(168, 85, 247, 0.25);
  border-color: rgba(168, 85, 247, 0.5);
}


.action-btn.validate {
  background: rgba(59, 130, 246, 0.15);
  color: rgb(59, 130, 246);
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.action-btn.validate:hover {
  background: rgba(59, 130, 246, 0.25);
  border-color: rgba(59, 130, 246, 0.5);
}

.action-btn.enforce {
  background: rgba(34, 197, 94, 0.15);
  color: rgb(34, 197, 94);
  border: 1px solid rgba(34, 197, 94, 0.3);
}

.action-btn.enforce:hover {
  background: rgba(34, 197, 94, 0.25);
  border-color: rgba(34, 197, 94, 0.5);
}

.action-btn.deploy-testing {
  background: rgba(59, 130, 246, 0.15);
  color: rgb(59, 130, 246);
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.action-btn.deploy-testing:hover {
  background: rgba(59, 130, 246, 0.25);
  border-color: rgba(59, 130, 246, 0.5);
}
</style>

