<!-- AI Rule Preview Component - Updated to use global styles -->
<script setup lang="ts">
import { computed } from 'vue'
import { CheckCircle2, AlertTriangle, Target } from 'lucide-vue-next'
import type { RuleGenResponse } from '../../types/ai'
import AIConfidenceBadge from '../ai/AIConfidenceBadge.vue'

const props = defineProps<{
  rule: RuleGenResponse
}>()

const emit = defineEmits<{
  deployToTesting: []
}>()

const hasWarnings = computed(() => props.rule.warnings && props.rule.warnings.length > 0)
</script>

<template>
  <div class="card-base rule-preview">
    <div class="card-content-base">
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
        <button class="btn btn-primary" @click="$emit('deployToTesting')">
          <Target :size="16" />
          <span>Deploy to Testing</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.rule-preview {
  margin-top: 24px;
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
  color: var(--status-safe);
}

.icon-warning {
  color: var(--status-warning);
}

.warnings {
  margin-bottom: 16px;
  padding: 12px;
  background: var(--status-warning-dim);
  border: 1px solid var(--status-warning);
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
  background: var(--bg-overlay);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  font-size: 12px;
  font-family: var(--font-mono);
  color: var(--text-primary);
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
}

.preview-actions {
  display: flex;
  gap: 12px;
  padding-top: 16px;
  border-top: 1px solid var(--border-subtle);
}
</style>
