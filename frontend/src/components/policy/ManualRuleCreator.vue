<!-- Manual Rule Creator Component -->
<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { FileCode, Trash2 } from 'lucide-vue-next'
import Select from '../common/Select.vue'
import type { Rule } from '../../types/rules'

const props = defineProps<{
  rule?: Rule | null
}>()

const emit = defineEmits<{
  'rule-created': [rule: Partial<Rule>]
  'rule-updated': [rule: Partial<Rule>]
  'rule-deleted': [ruleName: string]
  'cancel': []
}>()

const ruleName = ref('')
const description = ref('')
const action = ref<'block' | 'monitor' | 'allow'>('monitor')
const severity = ref<'critical' | 'high' | 'warning' | 'info'>('warning')
const state = ref<'production' | 'testing' | 'draft'>('draft')

// Match conditions
const matchType = ref<'exec' | 'file' | 'connect'>('exec')
const processName = ref('')
const filename = ref('')
const destPort = ref<number | ''>('')
const destIp = ref('')
const cgroup = ref('')
const uid = ref<number | ''>('')

const canCreate = computed(() => {
  if (!ruleName.value.trim() || !description.value.trim()) return false

  // At least one match condition must be set
  if (matchType.value === 'exec' && !processName.value.trim()) return false
  if (matchType.value === 'file' && !filename.value.trim()) return false
  if (matchType.value === 'connect' && !destPort.value && !destIp.value.trim()) return false

  return true
})

// Initialize form from rule prop if editing
watch(() => props.rule, (rule) => {
  if (rule) {
    ruleName.value = rule.name || ''
    description.value = rule.description || ''
    // Handle 'alert' action (legacy) by mapping to 'monitor'
    const ruleAction = (rule as any).action
    action.value = (ruleAction === 'alert' ? 'monitor' : ruleAction) as 'block' | 'monitor' | 'allow'
    severity.value = rule.severity as 'critical' | 'high' | 'warning' | 'info'
    const rawState = (rule as any).state
    if (rawState === 'testing' || rawState === 'production' || rawState === 'draft') {
      state.value = rawState as 'production' | 'testing' | 'draft'
    } else {
      state.value = 'draft'
    }

    // Determine match type from match object
    const match = rule.match || {}
    if (match.process_name || match.process) {
      matchType.value = 'exec'
      processName.value = match.process_name || match.process || ''
    } else if (match.filename) {
      matchType.value = 'file'
      filename.value = match.filename || ''
    } else if (match.dest_port || match.dest_ip) {
      matchType.value = 'connect'
      destPort.value = match.dest_port ? Number(match.dest_port) : ''
      destIp.value = match.dest_ip || ''
    }

    cgroup.value = match.cgroup_id || match.cgroup || ''
    uid.value = match.uid ? Number(uid.value) : ''
  }
}, { immediate: true })

const generateYaml = (): string => {
  const match: Record<string, any> = {}

  if (matchType.value === 'exec' && processName.value.trim()) {
    match.process_name = processName.value.trim()
  }
  if (matchType.value === 'file' && filename.value.trim()) {
    match.filename = filename.value.trim()
  }
  if (matchType.value === 'connect') {
    if (destPort.value) match.dest_port = Number(destPort.value)
    if (destIp.value.trim()) match.dest_ip = destIp.value.trim()
  }
  if (cgroup.value.trim()) match.cgroup_id = cgroup.value.trim()
  if (uid.value) match.uid = Number(uid.value)

  const rule: any = {
    name: ruleName.value.trim(),
    description: description.value.trim(),
    action: action.value,
    severity: severity.value,
    match
  }

  // Convert to YAML format
  let yaml = `name: ${rule.name}\n`
  yaml += `description: ${rule.description}\n`
  yaml += `action: ${rule.action}\n`
  yaml += `severity: ${rule.severity}\n`
  yaml += `match:\n`
  Object.entries(match).forEach(([key, value]) => {
    yaml += `  ${key}: ${value}\n`
  })

  return yaml
}

const createRule = () => {
  if (!canCreate.value) return

  const match: Record<string, any> = {}

  if (matchType.value === 'exec' && processName.value.trim()) {
    match.process_name = processName.value.trim()
  }
  if (matchType.value === 'file' && filename.value.trim()) {
    match.filename = filename.value.trim()
  }
  if (matchType.value === 'connect') {
    if (destPort.value) match.dest_port = Number(destPort.value)
    if (destIp.value.trim()) match.dest_ip = destIp.value.trim()
  }
  if (cgroup.value.trim()) match.cgroup_id = cgroup.value.trim()
  if (uid.value) match.uid = Number(uid.value)

  const rule: any = {
    name: ruleName.value.trim(),
    description: description.value.trim(),
    action: action.value,
    severity: severity.value,
    state: state.value,
    match,
    yaml: generateYaml()
  }

  if (props.rule) {
    emit('rule-updated', rule)
  } else {
    emit('rule-created', rule)
  }
}


const deleteRule = () => {
  if (!props.rule?.name) return
  if (confirm(`Are you sure you want to delete the rule "${props.rule.name}"? This action cannot be undone.`)) {
    emit('rule-deleted', props.rule.name)
  }
}
</script>

<template>
  <div class="manual-creator">
    <div class="creator-body">
      <!-- Basic Info -->
      <div class="form-section">
        <h4 class="section-title">Basic Information</h4>
        <div class="form-group">
          <label class="form-label">Rule Name *</label>
          <input v-model="ruleName" type="text" class="form-input" placeholder="e.g., Block Suspicious Process" />
        </div>
        <div class="form-group">
          <label class="form-label">Description *</label>
          <textarea v-model="description" class="form-textarea" rows="3"
            placeholder="Describe what this rule detects or blocks" />
        </div>
      </div>

      <!-- Rule Configuration -->
      <div class="form-section">
        <h4 class="section-title">Rule Configuration</h4>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Action</label>
            <Select v-model="action" :options="[
              { value: 'block', label: 'Block' },
              { value: 'monitor', label: 'Monitor' },
              { value: 'allow', label: 'Allow' }
            ]" />
          </div>
          <div class="form-group">
            <label class="form-label">Severity</label>
            <Select v-model="severity" :options="[
              { value: 'critical', label: 'Critical' },
              { value: 'high', label: 'High' },
              { value: 'warning', label: 'Warning' },
              { value: 'info', label: 'Info' }
            ]" />
          </div>
          <div class="form-group">
            <label class="form-label">State</label>
            <Select v-model="state" :options="[
              { value: 'draft', label: 'Draft' },
              { value: 'testing', label: 'Testing' },
              { value: 'production', label: 'Production' }
            ]" />
          </div>
        </div>
      </div>

      <!-- Match Conditions -->
      <div class="form-section">
        <h4 class="section-title">Match Conditions</h4>
        <div class="form-group">
          <label class="form-label">Match Type *</label>
          <Select v-model="matchType" :options="[
            { value: 'exec', label: 'Process Execution' },
            { value: 'file', label: 'File Access' },
            { value: 'connect', label: 'Network Connection' }
          ]" />
        </div>

        <!-- Exec Match -->
        <div v-if="matchType === 'exec'" class="form-group">
          <label class="form-label">Process Name *</label>
          <input v-model="processName" type="text" class="form-input" placeholder="e.g., /usr/bin/bash" />
        </div>

        <!-- File Match -->
        <div v-if="matchType === 'file'" class="form-group">
          <label class="form-label">File Path *</label>
          <input v-model="filename" type="text" class="form-input" placeholder="e.g., /tmp/suspicious.sh" />
        </div>

        <!-- Connect Match -->
        <div v-if="matchType === 'connect'" class="form-row">
          <div class="form-group">
            <label class="form-label">Destination Port</label>
            <input v-model.number="destPort" type="number" class="form-input" placeholder="e.g., 3306" />
          </div>
          <div class="form-group">
            <label class="form-label">Destination IP</label>
            <input v-model="destIp" type="text" class="form-input" placeholder="e.g., 192.168.1.100" />
          </div>
        </div>

        <!-- Optional Conditions -->
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Cgroup (optional)</label>
            <input v-model="cgroup" type="text" class="form-input" placeholder="e.g., /system.slice/nginx.service" />
          </div>
          <div class="form-group">
            <label class="form-label">UID (optional)</label>
            <input v-model.number="uid" type="number" class="form-input" placeholder="e.g., 1000" />
          </div>
        </div>
      </div>

      <!-- Actions -->
      <div class="form-actions">
        <button v-if="rule" class="btn-icon btn-danger" @click="deleteRule" title="Delete Rule">
          <Trash2 :size="18" />
        </button>
        <button class="btn-primary" @click="createRule" :disabled="!canCreate">
          <FileCode :size="16" />
          <span>{{ rule ? 'Update Rule' : 'Create Rule' }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.manual-creator {
  padding: 24px;
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
}

.creator-body {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.form-section {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 24px;
  background: var(--bg-overlay);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
}

.section-title {
  font-size: 11px;
  font-weight: 600; /* Softened */
  color: var(--text-muted);
  margin: 0;
  text-transform: uppercase;
  letter-spacing: 0.6px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.form-label {
  font-size: 13px; /* Increased size */
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 0;
}

.form-input,
.form-textarea {
  padding: 10px 14px;
  background: var(--bg-surface);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  font-size: 14px;
  font-family: inherit;
  color: var(--text-primary);
  transition: all var(--transition-fast);
}

.form-input:hover,
.form-textarea:hover {
  border-color: var(--border-default);
}

.form-input:focus,
.form-textarea:focus {
  outline: none;
  border-color: var(--accent-primary);
  background: var(--bg-surface);
}

.form-textarea {
  resize: vertical;
  min-height: 100px;
}

.form-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  padding-top: 24px;
  margin-top: 0;
  border-top: 1px solid var(--border-subtle);
}

.btn-primary,
.btn-secondary,
.btn-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px 20px;
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  border: 1px solid transparent;
}

.btn-icon {
  padding: 10px;
  width: 40px;
  height: 40px;
}

.btn-primary {
  background: var(--accent-primary);
  color: white;
  border-color: var(--accent-primary);
}

.btn-primary:hover:not(:disabled) {
  background: var(--accent-primary-hover);
  border-color: var(--accent-primary-hover);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background: var(--bg-surface);
  color: var(--text-secondary);
  border-color: var(--border-default);
}

.btn-secondary:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.btn-danger {
  background: var(--status-critical-dim);
  color: var(--status-critical);
  border-color: var(--status-critical);
}

.btn-danger:hover:not(:disabled) {
  background: var(--status-critical);
  color: white;
}
</style>