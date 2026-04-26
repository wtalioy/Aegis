<!-- Manual Rule Creator Component -->
<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import { FileCode, Trash2 } from 'lucide-vue-next'
import Select from '../common/Select.vue'
import type { Rule, RuleAction, RuleMatch, RuleState, RuleType } from '../../types/rules'

const props = defineProps<{
  rule?: Rule | null
}>()

const emit = defineEmits<{
  'rule-created': [rule: Partial<Rule>]
  'rule-updated': [rule: Partial<Rule>]
  'rule-deleted': [ruleName: string]
  'cancel': []
}>()

const form = reactive({
  name: '',
  description: '',
  action: 'alert' as RuleAction,
  severity: 'warning' as 'critical' | 'high' | 'warning' | 'info',
  state: 'draft' as RuleState,
  matchType: 'exec' as RuleType,
  processName: '',
  filename: '',
  destPort: '' as number | '',
  destIp: '',
  cgroupId: '',
  uid: '' as number | ''
})

const matchFields: Record<RuleType, Array<keyof RuleMatch | 'processName' | 'filename' | 'destPort' | 'destIp'>> = {
  exec: ['processName'],
  file: ['filename'],
  connect: ['destPort', 'destIp']
}

const canCreate = computed(() => {
  if (!form.name.trim() || !form.description.trim()) return false

  if (form.matchType === 'exec' && !form.processName.trim()) return false
  if (form.matchType === 'file' && !form.filename.trim()) return false
  if (form.matchType === 'connect' && !form.destPort && !form.destIp.trim()) return false

  return true
})

watch(() => props.rule, (rule) => {
  if (rule) {
    form.name = rule.name || ''
    form.description = rule.description || ''
    form.action = rule.action
    form.severity = rule.severity as 'critical' | 'high' | 'warning' | 'info'
    form.state = rule.state
    const match = rule.match || {}
    if (match.processName) {
      form.matchType = 'exec'
      form.processName = match.processName
    } else if (match.filename) {
      form.matchType = 'file'
      form.filename = match.filename || ''
    } else if (match.destPort || match.destIp) {
      form.matchType = 'connect'
      form.destPort = match.destPort ? Number(match.destPort) : ''
      form.destIp = match.destIp || ''
    }
    form.cgroupId = match.cgroupId || ''
    form.uid = match.uid ? Number(match.uid) : ''
    return
  }
  form.name = ''
  form.description = ''
  form.action = 'alert'
  form.severity = 'warning'
  form.state = 'draft'
  form.matchType = 'exec'
  form.processName = ''
  form.filename = ''
  form.destPort = ''
  form.destIp = ''
  form.cgroupId = ''
  form.uid = ''
}, { immediate: true })

const buildMatch = (): RuleMatch => {
  const match: RuleMatch = {}

  if (form.matchType === 'exec' && form.processName.trim()) {
    match.processName = form.processName.trim()
  }
  if (form.matchType === 'file' && form.filename.trim()) {
    match.filename = form.filename.trim()
  }
  if (form.matchType === 'connect') {
    if (form.destPort) match.destPort = Number(form.destPort)
    if (form.destIp.trim()) match.destIp = form.destIp.trim()
  }
  if (form.cgroupId.trim()) match.cgroupId = form.cgroupId.trim()
  if (form.uid) match.uid = Number(form.uid)

  return match
}

const generateYaml = (): string => {
  const match = buildMatch()

  const rule = {
    name: form.name.trim(),
    description: form.description.trim(),
    action: form.action,
    severity: form.severity,
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

  const rule: Rule = {
    name: form.name.trim(),
    description: form.description.trim(),
    action: form.action,
    severity: form.severity,
    state: form.state,
    type: form.matchType,
    match: buildMatch(),
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
          <input v-model="form.name" type="text" class="form-input" placeholder="e.g., Block Suspicious Process" />
        </div>
        <div class="form-group">
          <label class="form-label">Description *</label>
          <textarea v-model="form.description" class="form-textarea" rows="3"
            placeholder="Describe what this rule detects or blocks" />
        </div>
      </div>

      <!-- Rule Configuration -->
      <div class="form-section">
        <h4 class="section-title">Rule Configuration</h4>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Action</label>
            <Select v-model="form.action" :options="[
              { value: 'block', label: 'Block' },
              { value: 'alert', label: 'Alert' },
              { value: 'allow', label: 'Allow' }
            ]" />
          </div>
          <div class="form-group">
            <label class="form-label">Severity</label>
            <Select v-model="form.severity" :options="[
              { value: 'critical', label: 'Critical' },
              { value: 'high', label: 'High' },
              { value: 'warning', label: 'Warning' },
              { value: 'info', label: 'Info' }
            ]" />
          </div>
          <div class="form-group">
            <label class="form-label">State</label>
            <Select v-model="form.state" :options="[
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
          <Select v-model="form.matchType" :options="[
            { value: 'exec', label: 'Process Execution' },
            { value: 'file', label: 'File Access' },
            { value: 'connect', label: 'Network Connection' }
          ]" />
        </div>

        <!-- Exec Match -->
        <div v-if="matchFields.exec && form.matchType === 'exec'" class="form-group">
          <label class="form-label">Process Name *</label>
          <input v-model="form.processName" type="text" class="form-input" placeholder="e.g., /usr/bin/bash" />
        </div>

        <!-- File Match -->
        <div v-if="matchFields.file && form.matchType === 'file'" class="form-group">
          <label class="form-label">File Path *</label>
          <input v-model="form.filename" type="text" class="form-input" placeholder="e.g., /tmp/suspicious.sh" />
        </div>

        <!-- Connect Match -->
        <div v-if="matchFields.connect && form.matchType === 'connect'" class="form-row">
          <div class="form-group">
            <label class="form-label">Destination Port</label>
            <input v-model.number="form.destPort" type="number" class="form-input" placeholder="e.g., 3306" />
          </div>
          <div class="form-group">
            <label class="form-label">Destination IP</label>
            <input v-model="form.destIp" type="text" class="form-input" placeholder="e.g., 192.168.1.100" />
          </div>
        </div>

        <!-- Optional Conditions -->
        <div class="form-row">
          <div class="form-group">
            <label class="form-label">Cgroup (optional)</label>
            <input v-model="form.cgroupId" type="text" class="form-input" placeholder="e.g., /system.slice/nginx.service" />
          </div>
          <div class="form-group">
            <label class="form-label">UID (optional)</label>
            <input v-model.number="form.uid" type="number" class="form-input" placeholder="e.g., 1000" />
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
