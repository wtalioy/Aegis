<!-- Policy Studio Page - Redesigned with Rules.vue inspiration -->
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  Sparkles, FileCode, Search, Filter, RefreshCw,
  ShieldOff, AlertTriangle, ShieldCheck, X
} from 'lucide-vue-next'
import { getRules, createRule, updateRule, deleteRule } from '../lib/api'
import type { Rule } from '../types/rules'
import RuleCard from '../components/rules/RuleCard.vue'
import AIRuleCreator from '../components/policy/AIRuleCreator.vue'
import ManualRuleCreator from '../components/policy/ManualRuleCreator.vue'
import DeployConfirm from '../components/policy/DeployConfirm.vue'
import AIRulePreview from '../components/policy/AIRulePreview.vue'
import Select from '../components/common/Select.vue'

const router = useRouter()

const rules = ref<Rule[]>([])
const selectedRule = ref<Rule | null>(null)
const generatedRule = ref<any>(null)
const showDeployConfirm = ref(false)
const createMode = ref<'manual' | 'ai'>('manual')
const loading = ref(true)
const searchQuery = ref('')
const filterAction = ref<string>('all')

const normalizeRuleState = (state: any): 'draft' | 'testing' | 'production' => {
  // Handle null, undefined, or empty string
  if (!state || state === '') return 'draft'

  // Convert to string and normalize case
  const stateStr = String(state).toLowerCase().trim()

  // Map valid states
  if (stateStr === 'testing') return 'testing'
  if (stateStr === 'production') return 'production'
  if (stateStr === 'draft') return 'draft'

  // Default to draft for unknown values
  return 'draft'
}

const fetchRules = async () => {
  loading.value = true
  try {
    const apiRules = await getRules()
    rules.value = apiRules.map(rule => {
      const rawState = (rule as any).state
      return {
        ...rule,
        state: normalizeRuleState(rawState),
        action: (rule.action === 'alert' ? 'monitor' : rule.action) as 'block' | 'monitor' | 'allow',
        severity: rule.severity as 'critical' | 'high' | 'warning' | 'info',
        match: rule.match || {}
      }
    })
  } catch (e) {
    console.error('Failed to fetch rules:', e)
    rules.value = []
  } finally {
    loading.value = false
  }
}

const filteredRules = computed(() => {
  let result = rules.value

  if (filterAction.value !== 'all') {
    result = result.filter(r => r.action === filterAction.value)
  }

  if (searchQuery.value.trim()) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(r =>
      r.name.toLowerCase().includes(query) ||
      r.description.toLowerCase().includes(query)
    )
  }

  return result
})

const stats = computed(() => ({
  total: rules.value.length,
  block: rules.value.filter(r => r.action === 'block').length,
  alert: rules.value.filter(r => r.action === 'monitor').length,
  allow: rules.value.filter(r => r.action === 'allow').length,
  testing: rules.value.filter(r => (r as any).state === 'testing').length,
  production: rules.value.filter(r => (r as any).state === 'production').length,
}))

const handleRuleSelect = (rule: Rule) => {
  selectedRule.value = rule
  // Switch to manual mode for editing
  createMode.value = 'manual'
  // Clear any generated rule preview
  generatedRule.value = null
}


const handleManualRuleCreated = (rule: Partial<Rule>) => {
  // Convert manual rule to generated rule format for preview
  generatedRule.value = {
    rule: rule as Rule,
    yaml: rule.yaml || '',
    reasoning: 'Manually created rule',
    confidence: 1.0,
    warnings: []
  }
  // Switch to AI preview to show the created rule
  createMode.value = 'ai'
}

const handleManualRuleUpdated = async (rule: Partial<Rule>) => {
  if (!selectedRule.value) return

  try {
    const updatedRule = await updateRule(selectedRule.value.name, rule as Rule)
    // Refresh rules list
    await fetchRules()
    // Update selected rule
    selectedRule.value = updatedRule as Rule
    // Clear selection to show create mode
    selectedRule.value = null
  } catch (error) {
    console.error('Failed to update rule:', error)
    alert(`Failed to update rule: ${error instanceof Error ? error.message : 'Unknown error'}`)
  }
}

const handleEditCancel = () => {
  selectedRule.value = null
  generatedRule.value = null
  createMode.value = 'manual'
}

const handleRuleDeleted = async (ruleName: string) => {
  try {
    await deleteRule(ruleName)
    // Refresh rules list
    await fetchRules()
    // Clear selection
    selectedRule.value = null
    generatedRule.value = null
    createMode.value = 'manual'
  } catch (error) {
    console.error('Failed to delete rule:', error)
    alert(`Failed to delete rule: ${error instanceof Error ? error.message : 'Unknown error'}`)
  }
}

const handleConfirmDeploy = async () => {
  const ruleToDeploy = generatedRule.value?.rule || selectedRule.value
  if (!ruleToDeploy) {
    console.error('No rule to deploy')
    return
  }

  try {
    // Set the rule state to testing (backend expects 'state' field)
    const ruleWithState = {
      ...ruleToDeploy,
      state: (ruleToDeploy as any).state || 'testing'
    }

    // Create the rule via API
    await createRule(ruleWithState, 'testing')

    // Close modal and refresh rules
    showDeployConfirm.value = false
    await fetchRules()

    // Clear the generated rule and reset to create mode
    generatedRule.value = null
    createMode.value = 'manual'

    // Navigate to rule validation page to see the deployed rule
    router.push({
      path: '/rule-validation',
      query: {
        rule: ruleToDeploy.name,
        from: 'deploy'
      }
    })
  } catch (error) {
    console.error('Failed to deploy rule:', error)
    alert(`Failed to deploy rule: ${error instanceof Error ? error.message : 'Unknown error'}`)
  }
}



const route = useRoute()

onMounted(async () => {
  await fetchRules()
  const qp = (route as any)?.query || {}
  const ruleParam = typeof qp.rule === 'string' ? qp.rule : null
  if (ruleParam) {
    const match = rules.value.find(r => r.name === ruleParam)
    if (match) {
      selectedRule.value = match
    }
  }
})
</script>

<template>
  <div class="policy-studio-page">
    <!-- Header -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">
          Policy Studio
        </h1>
        <span class="page-subtitle">AI-assisted security rule creation and management</span>
      </div>

      <!-- Action Stats -->
      <div class="action-stats">
        <button class="action-stat block" :class="{ active: filterAction === 'block' }"
          @click="filterAction = filterAction === 'block' ? 'all' : 'block'">
          <ShieldOff :size="16" />
          <span class="stat-value">{{ stats.block }}</span>
          <span class="stat-label">Block</span>
        </button>
        <button class="action-stat alert" :class="{ active: filterAction === 'alert' }"
          @click="filterAction = filterAction === 'alert' ? 'all' : 'alert'">
          <AlertTriangle :size="16" />
          <span class="stat-value">{{ stats.alert }}</span>
          <span class="stat-label">Alert</span>
        </button>
        <button class="action-stat allow" :class="{ active: filterAction === 'allow' }"
          @click="filterAction = filterAction === 'allow' ? 'all' : 'allow'">
          <ShieldCheck :size="16" />
          <span class="stat-value">{{ stats.allow }}</span>
          <span class="stat-label">Allow</span>
        </button>
      </div>
    </div>

    <!-- Filters Bar -->
    <div class="filters-bar">
      <div class="search-box">
        <Search :size="16" class="search-icon" />
        <input v-model="searchQuery" type="text" placeholder="Search rules..." class="search-input" />
      </div>

      <div class="filter-group">
        <Filter :size="14" class="filter-icon" />
        <Select v-model="filterAction" :options="[
          { value: 'all', label: 'All Actions' },
          { value: 'block', label: 'Block' },
          { value: 'alert', label: 'Alert' },
          { value: 'allow', label: 'Allow' }
        ]" background="var(--bg-elevated)" />
      </div>

      <button class="refresh-btn" @click="fetchRules" :disabled="loading">
        <RefreshCw :size="16" :class="{ spinning: loading }" />
        Refresh
      </button>
    </div>

    <!-- Main Content Layout -->
    <div class="studio-layout">
      <!-- Left: Rules List -->
      <div class="rules-panel">
        <div class="panel-header">
          <h3>Existing Rules</h3>
        </div>

        <div class="rules-content">
          <div v-if="loading" class="loading-state">
            <div class="loading-spinner"></div>
            <span>Loading rules...</span>
          </div>

          <div v-else-if="rules.length === 0" class="empty-state">
            <FileCode :size="32" class="empty-icon" />
            <div class="empty-title">No Rules</div>
            <div class="empty-description">Create your first rule using AI</div>
          </div>

          <div v-else-if="filteredRules.length === 0" class="empty-state">
            <Search :size="32" class="empty-icon" />
            <div class="empty-title">No Matching Rules</div>
            <div class="empty-description">Try adjusting your search or filters</div>
          </div>

          <div v-else class="rules-list">
            <div v-for="rule in filteredRules" :key="rule.name" class="rule-item-wrapper"
              :class="{ selected: selectedRule?.name === rule.name }" @click="handleRuleSelect(rule)">
              <RuleCard :rule="rule" />
            </div>
          </div>
        </div>
      </div>

      <!-- Right: Workspace -->
      <div class="workspace-panel">
        <!-- Workspace Header -->
        <div class="panel-header">
          <h3>{{ selectedRule ? 'Edit Rule' : 'Create Rule' }}</h3>
          <button v-if="selectedRule" class="cancel-edit-btn" @click="handleEditCancel">
            <X :size="16" />
            <span>Cancel Edit</span>
          </button>
        </div>

        <!-- Create Rule Content -->
        <div class="workspace-content">
          <!-- Creation Mode Toggle (only show when not editing) -->
          <div v-if="!selectedRule" class="create-mode-toggle">
            <button class="mode-btn" :class="{ active: createMode === 'manual' }" @click="createMode = 'manual'">
              <FileCode :size="16" />
              <span>Manual</span>
            </button>
            <button class="mode-btn" :class="{ active: createMode === 'ai' }" @click="createMode = 'ai'">
              <Sparkles :size="16" />
              <span>AI Assistant</span>
            </button>
          </div>

          <!-- Manual Creation/Editing -->
          <div v-show="createMode === 'manual'" class="creator-section">
            <ManualRuleCreator :rule="selectedRule" @rule-created="handleManualRuleCreated"
              @rule-updated="handleManualRuleUpdated" @rule-deleted="handleRuleDeleted" />
          </div>

          <!-- AI Creation -->
          <div v-show="createMode === 'ai'" class="creator-section">
            <AIRuleCreator @rule-generated="(rule) => { generatedRule = rule }" />
          </div>

          <!-- Rule Preview (shown when rule is generated from either method) -->
          <div v-if="generatedRule" class="preview-section">
            <AIRulePreview :rule="generatedRule" @deployToTesting="() => { showDeployConfirm = true }" />
          </div>
        </div>


      </div>
    </div>

    <DeployConfirm v-if="generatedRule || selectedRule" :rule="generatedRule?.rule || selectedRule!" :mode="'testing'"
      :visible="showDeployConfirm" @confirm="handleConfirmDeploy" @cancel="() => { showDeployConfirm = false }" />
  </div>
</template>

<style scoped>
.policy-studio-page {
  padding: 24px;
  max-width: 1920px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-height: calc(100vh - 120px);
}

/* Header */
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 20px;
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
  font-weight: 600;
  /* Softened from 700 */
  color: var(--text-primary);
  margin: 0;
}

.title-icon {
  color: var(--accent-primary);
}

.page-subtitle {
  font-size: 14px;
  /* Slightly larger */
  color: var(--text-muted);
}

/* Action Stats */
.action-stats {
  display: flex;
  gap: 12px;
}

.action-stat {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  background: var(--bg-surface);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.action-stat:hover {
  background: var(--bg-hover);
  border-color: var(--border-default);
}

.action-stat.active {
  background: var(--bg-overlay);
}

.action-stat.block.active {
  border-color: var(--status-blocked);
  background: var(--status-blocked-dim);
  color: var(--status-blocked);
}

.action-stat.alert.active {
  border-color: var(--status-warning);
  background: var(--status-warning-dim);
  color: var(--status-warning);
}

.action-stat.allow.active {
  border-color: var(--status-safe);
  background: var(--status-safe-dim);
  color: var(--status-safe);
}

.action-stat .stat-value {
  font-size: 16px;
  /* Softened */
  font-weight: 600;
  font-family: var(--font-mono);
}

.action-stat .stat-label {
  font-size: 12px;
  opacity: 0.8;
}

/* Filters Bar */
.filters-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: var(--bg-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  flex-wrap: wrap;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 14px;
  background: var(--bg-overlay);
  border-radius: var(--radius-md);
  border: 1px solid transparent;
  flex: 1;
  min-width: 200px;
  height: 40px;
  box-sizing: border-box;
  transition: all var(--transition-normal);
}

.search-box:focus-within {
  border-color: var(--accent-primary);
  background: var(--bg-surface);
  box-shadow: 0 0 0 3px var(--accent-glow);
}

.search-icon {
  color: var(--text-muted);
  flex-shrink: 0;
}

.search-input {
  flex: 1;
  background: transparent;
  border: none;
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
}

.search-input::placeholder {
  color: var(--text-muted);
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 40px;
}

.filter-icon {
  color: var(--text-muted);
}

.refresh-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 16px;
  background: transparent;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
  height: 40px;
  box-sizing: border-box;
}

.refresh-btn:hover:not(:disabled) {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.refresh-btn .spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* Main Layout */
.studio-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  flex: 1;
  min-height: 0;
}

/* Rules Panel */
.rules-panel,
.workspace-panel {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-subtle);
  flex-shrink: 0;
}

.panel-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.cancel-edit-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: transparent;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.cancel-edit-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.rules-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.rules-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.rule-item-wrapper {
  cursor: pointer;
  border-radius: var(--radius-md);
  border: 1px solid transparent;
  transition: all var(--transition-fast);
}

.rule-item-wrapper.selected {
  border-color: var(--accent-primary);
  background-color: var(--bg-hover);
}

/* Workspace Panel */
.workspace-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.create-mode-toggle {
  display: flex;
  gap: 4px;
  padding: 4px;
  background: var(--bg-overlay);
  border-radius: var(--radius-md);
}

.mode-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  flex: 1;
  padding: 8px 12px;
  border: 1px solid transparent;
  background: transparent;
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}

.mode-btn:hover:not(.active) {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.mode-btn.active {
  background: var(--bg-surface);
  color: var(--text-primary);
  box-shadow: var(--shadow-sm);
}

.creator-section,
.preview-section {
  flex-shrink: 0;
}

.preview-section {
  border-top: 1px solid var(--border-subtle);
  padding-top: 24px;
}

/* Loading & Empty States */
.loading-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
}

.loading-spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border-subtle);
  border-top-color: var(--accent-primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 16px;
}

.loading-state span {
  color: var(--text-muted);
  font-size: 14px;
}

.empty-icon {
  color: var(--text-muted);
  opacity: 0.5;
  margin-bottom: 12px;
}

.empty-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 6px;
}

.empty-description {
  font-size: 13px;
  color: var(--text-muted);
}

@media (max-width: 1400px) {
  .studio-layout {
    grid-template-columns: 1fr;
  }

  .rules-panel {
    max-height: 400px;
  }
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .action-stats {
    width: 100%;
    justify-content: space-between;
  }

  .action-stat {
    flex: 1;
    justify-content: center;
    padding: 12px;
  }

  .action-stat .stat-label {
    display: none;
  }

  .filters-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-group {
    flex-wrap: wrap;
  }
}
</style>