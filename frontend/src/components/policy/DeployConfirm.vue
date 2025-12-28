<!-- Deploy Confirm Component - Phase 4 -->
<script setup lang="ts">
import { computed } from 'vue'
import { AlertTriangle, CheckCircle2, Moon, Shield } from 'lucide-vue-next'
import type { Rule } from '../../types/rules'

const props = defineProps<{
  rule: Rule
  mode: 'testing' | 'production'
  visible: boolean

}>()

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()

const modeIcon = computed(() => props.mode === 'testing' ? Moon : Shield)
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="visible" class="modal-overlay" @click.self="$emit('cancel')">
        <div class="modal-container">
          <div class="modal-header">
            <h3>{{ mode === 'testing' ? 'Deploy to Testing' : 'Deploy to Production' }}</h3>
            <button class="close-btn" @click="$emit('cancel')">×</button>
          </div>

          <div class="modal-body">
            <div class="rule-info">
              <div class="rule-name">{{ rule.name }}</div>
              <div class="rule-description">{{ rule.description }}</div>
            </div>

            <div class="deploy-mode">
              <component :is="modeIcon" :size="20" />
              <div class="mode-content">
                <div class="mode-label">{{ mode === 'testing' ? 'Testing Mode' : 'Production Mode' }}</div>
                <div class="mode-description">
                  <template v-if="mode === 'testing'">
                    The rule will be saved and start monitoring real traffic. It will match events and record them, but will <strong>not block</strong> anything. You can review the results and promote it to production later.
                  </template>
                  <template v-else>
                    The rule will be saved and <strong>immediately start blocking</strong> matching events. Only use this if you've already tested the rule in testing mode.
                  </template>
                </div>
              </div>
            </div>

            <div v-if="mode === 'testing'" class="info-section">
              <CheckCircle2 :size="16" />
              <div class="info-text">
                <strong>What happens next:</strong> After deployment, you can view this rule's activity in the <strong>Rule Validation</strong> page. Once it has enough data (24+ hours, 10+ matches, low false positives), you can promote it to production.
              </div>
            </div>

            <div class="warning-section">
              <AlertTriangle :size="16" />
              <div class="warning-text">
                <strong>Note:</strong> This will save the rule and {{ mode === 'production' ? 'immediately start blocking' : 'start monitoring' }} matching events.
                {{ mode === 'production' ? 'Make sure you have tested this rule in testing mode first.' : '' }}
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button class="btn-secondary" @click="$emit('cancel')">Cancel</button>
            <button class="btn-primary" @click="$emit('confirm')">
              <CheckCircle2 :size="16" />
              <span>{{ mode === 'testing' ? 'Deploy to Testing' : 'Deploy to Production' }}</span>
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  z-index: 10000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.modal-container {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  width: 100%;
  max-width: 600px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px;
  border-bottom: 1px solid var(--border-subtle);
}

.modal-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.close-btn {
  padding: 4px 12px;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: 24px;
  cursor: pointer;
  border-radius: var(--radius-sm);
  transition: all 0.15s;
}

.close-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.modal-body {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.rule-info {
  padding-bottom: 20px;
  border-bottom: 1px solid var(--border-subtle);
}

.rule-name {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.rule-description {
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-secondary);
}

.deploy-mode {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 16px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
}

.mode-content {
  flex: 1;
}

.mode-label {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.mode-description {
  font-size: 13px;
  line-height: 1.5;
  color: var(--text-secondary);
}

.simulation-preview {
  padding: 16px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
}

.preview-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 12px;
}

.preview-stats {
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

.info-section {
  display: flex;
  gap: 12px;
  padding: 16px;
  background: rgba(59, 130, 246, 0.1);
  border: 1px solid rgba(59, 130, 246, 0.2);
  border-radius: var(--radius-md);
}

.info-text {
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-secondary);
}

.warning-section {
  display: flex;
  gap: 12px;
  padding: 16px;
  background: rgba(251, 191, 36, 0.1);
  border: 1px solid rgba(251, 191, 36, 0.2);
  border-radius: var(--radius-md);
}

.warning-text {
  font-size: 13px;
  line-height: 1.6;
  color: var(--text-secondary);
}

.modal-footer {
  display: flex;
  gap: 12px;
  padding: 20px;
  border-top: 1px solid var(--border-subtle);
  justify-content: flex-end;
}

.btn-secondary,
.btn-primary {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  border-radius: var(--radius-md);
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
}

.btn-secondary {
  background: var(--bg-elevated);
  color: var(--text-secondary);
  border: 1px solid var(--border-subtle);
}

.btn-secondary:hover:not(:disabled) {
  background: var(--bg-hover);
  color: var(--text-primary);
  transform: translateY(-1px);
}

.btn-secondary:active:not(:disabled) {
  transform: translateY(0);
}

.btn-primary {
  background: var(--accent-primary);
  color: white;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.25);
}

.btn-primary:hover:not(:disabled) {
  background: var(--accent-hover);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.35);
  transform: translateY(-1px);
}

.btn-primary:active:not(:disabled) {
  transform: translateY(0);
  box-shadow: 0 2px 6px rgba(59, 130, 246, 0.3);
}

.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>

