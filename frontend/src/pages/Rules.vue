<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { FileCode, Terminal, Globe, FileText, RefreshCw, Search, Filter, Shield, ShieldCheck } from 'lucide-vue-next'
import RuleCard from '../components/rules/RuleCard.vue'
import { getRules, subscribeToRulesReload, type Rule } from '../lib/api'

const rules = ref<Rule[]>([])
const loading = ref(true)
const searchQuery = ref('')
const filterType = ref<string>('all')
const filterSeverity = ref<string>('all')
const filterAction = ref<string>('all')

let unsubscribeReload: (() => void) | null = null

const fetchRules = async () => {
  loading.value = true
  try {
    rules.value = await getRules()
  } catch (e) {
    console.error('Failed to fetch rules:', e)
    rules.value = []
  } finally {
    loading.value = false
  }
}

const deriveRuleType = (rule: Rule): string => {
  if (rule.type) return rule.type
  if (rule.match?.filename || rule.match?.file_path) return 'file'
  if (rule.match?.dest_port || rule.match?.dest_ip) return 'connect'
  return 'exec'
}

const filteredRules = computed(() => {
  let result = rules.value

  if (filterAction.value !== 'all') {
    result = result.filter(r => r.action === filterAction.value)
  }

  if (filterType.value !== 'all') {
    result = result.filter(r => deriveRuleType(r) === filterType.value)
  }

  if (filterSeverity.value !== 'all') {
    result = result.filter(r => r.severity === filterSeverity.value)
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

const groupedRules = computed(() => {
  const alertRules: Record<string, Rule[]> = { exec: [], file: [], connect: [] }
  const allowRules: Record<string, Rule[]> = { exec: [], file: [], connect: [] }
  
  filteredRules.value.forEach(rule => {
    const target = rule.action === 'allow' ? allowRules : alertRules
    const ruleType = deriveRuleType(rule)
    if (target[ruleType]) {
      target[ruleType].push(rule)
    }
  })

  return { alert: alertRules, allow: allowRules }
})

const stats = computed(() => ({
  total: rules.value.length,
  alert: rules.value.filter(r => r.action !== 'allow').length,
  allow: rules.value.filter(r => r.action === 'allow').length,
  exec: rules.value.filter(r => r.type === 'exec').length,
  file: rules.value.filter(r => r.type === 'file').length,
  connect: rules.value.filter(r => r.type === 'connect').length,
}))

const hasAlertRules = computed(() => 
  groupedRules.value.alert.exec.length > 0 ||
  groupedRules.value.alert.file.length > 0 ||
  groupedRules.value.alert.connect.length > 0
)

const hasAllowRules = computed(() => 
  groupedRules.value.allow.exec.length > 0 ||
  groupedRules.value.allow.file.length > 0 ||
  groupedRules.value.allow.connect.length > 0
)

onMounted(() => {
  fetchRules()
  unsubscribeReload = subscribeToRulesReload(() => {
    fetchRules()
  })
})

onUnmounted(() => {
  unsubscribeReload?.()
})
</script>

<template>
  <div class="rules-page">
    <!-- Header -->
    <div class="page-header">
      <div class="header-content">
        <h1 class="page-title">
          <FileCode :size="24" class="title-icon" />
          Detection Rules
        </h1>
        <span class="page-subtitle">Manage security detection rules</span>
      </div>
      <div class="header-stats">
        <div class="stat-item alert">
          <Shield :size="14" class="stat-icon" />
          <span class="stat-value">{{ stats.alert }}</span>
          <span class="stat-label">Alert</span>
        </div>
        <div class="stat-item allow">
          <ShieldCheck :size="14" class="stat-icon" />
          <span class="stat-value">{{ stats.allow }}</span>
          <span class="stat-label">Allow</span>
        </div>
        <div class="stat-item">
          <Terminal :size="14" class="stat-icon exec" />
          <span class="stat-value">{{ stats.exec }}</span>
          <span class="stat-label">Exec</span>
        </div>
        <div class="stat-item">
          <FileText :size="14" class="stat-icon file" />
          <span class="stat-value">{{ stats.file }}</span>
          <span class="stat-label">File</span>
        </div>
        <div class="stat-item">
          <Globe :size="14" class="stat-icon connect" />
          <span class="stat-value">{{ stats.connect }}</span>
          <span class="stat-label">Network</span>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="filters-bar">
      <div class="search-box">
        <Search :size="16" class="search-icon" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search rules..."
          class="search-input"
        />
      </div>

      <div class="filter-group">
        <Filter :size="14" class="filter-icon" />
        <select v-model="filterAction" class="filter-select">
          <option value="all">All Actions</option>
          <option value="alert">Alert Rules</option>
          <option value="allow">Allow Rules</option>
        </select>

        <select v-model="filterType" class="filter-select">
          <option value="all">All Types</option>
          <option value="exec">Exec</option>
          <option value="file">File</option>
          <option value="connect">Network</option>
        </select>

        <select v-model="filterSeverity" class="filter-select">
          <option value="all">All Severity</option>
          <option value="high">High</option>
          <option value="warning">Warning</option>
          <option value="info">Info</option>
        </select>
      </div>

      <button class="refresh-btn" @click="fetchRules" :disabled="loading">
        <RefreshCw :size="16" :class="{ spinning: loading }" />
        Refresh
      </button>
    </div>

    <!-- Content -->
    <div class="rules-content">
      <!-- Loading State -->
      <div v-if="loading" class="loading-state">
        <div class="loading-spinner"></div>
        <span>Loading rules...</span>
      </div>

      <!-- Empty State -->
      <div v-else-if="rules.length === 0" class="empty-state">
        <div class="empty-icon">📝</div>
        <div class="empty-title">No Rules Loaded</div>
        <div class="empty-description">
          No detection rules have been loaded. Add rules to your rules.yaml file and restart the application.
        </div>
      </div>

      <!-- No Matches -->
      <div v-else-if="filteredRules.length === 0" class="empty-state">
        <div class="empty-icon">🔍</div>
        <div class="empty-title">No Matching Rules</div>
        <div class="empty-description">
          Try adjusting your search or filters to find rules.
        </div>
      </div>

      <!-- Rule Groups -->
      <template v-else>
        <!-- Alert Rules Section -->
        <div v-if="hasAlertRules" class="rules-section">
          <div class="section-header alert">
            <Shield :size="18" />
            <h2>Alert Rules</h2>
            <span class="section-badge">{{ stats.alert }}</span>
          </div>

          <!-- Exec Alert Rules -->
          <div v-if="groupedRules.alert.exec.length > 0" class="rule-group">
            <div class="group-header">
              <div class="group-icon exec">
                <Terminal :size="16" />
              </div>
              <h3 class="group-title">Process Execution</h3>
              <span class="group-count">{{ groupedRules.alert.exec.length }}</span>
            </div>
            <div class="group-content">
              <RuleCard
                v-for="rule in groupedRules.alert.exec"
                :key="rule.name"
                :rule="rule"
              />
            </div>
          </div>

          <!-- File Alert Rules -->
          <div v-if="groupedRules.alert.file.length > 0" class="rule-group">
            <div class="group-header">
              <div class="group-icon file">
                <FileText :size="16" />
              </div>
              <h3 class="group-title">File Access</h3>
              <span class="group-count">{{ groupedRules.alert.file.length }}</span>
            </div>
            <div class="group-content">
              <RuleCard
                v-for="rule in groupedRules.alert.file"
                :key="rule.name"
                :rule="rule"
              />
            </div>
          </div>

          <!-- Network Alert Rules -->
          <div v-if="groupedRules.alert.connect.length > 0" class="rule-group">
            <div class="group-header">
              <div class="group-icon connect">
                <Globe :size="16" />
              </div>
              <h3 class="group-title">Network Connection</h3>
              <span class="group-count">{{ groupedRules.alert.connect.length }}</span>
            </div>
            <div class="group-content">
              <RuleCard
                v-for="rule in groupedRules.alert.connect"
                :key="rule.name"
                :rule="rule"
              />
            </div>
          </div>
        </div>

        <!-- Allow Rules Section -->
        <div v-if="hasAllowRules" class="rules-section">
          <div class="section-header allow">
            <ShieldCheck :size="18" />
            <h2>Allow Rules (Whitelist)</h2>
            <span class="section-badge">{{ stats.allow }}</span>
          </div>

          <!-- Exec Allow Rules -->
          <div v-if="groupedRules.allow.exec.length > 0" class="rule-group">
            <div class="group-header">
              <div class="group-icon exec">
                <Terminal :size="16" />
              </div>
              <h3 class="group-title">Process Execution</h3>
              <span class="group-count">{{ groupedRules.allow.exec.length }}</span>
            </div>
            <div class="group-content">
              <RuleCard
                v-for="rule in groupedRules.allow.exec"
                :key="rule.name"
                :rule="rule"
              />
            </div>
          </div>

          <!-- File Allow Rules -->
          <div v-if="groupedRules.allow.file.length > 0" class="rule-group">
            <div class="group-header">
              <div class="group-icon file">
                <FileText :size="16" />
              </div>
              <h3 class="group-title">File Access</h3>
              <span class="group-count">{{ groupedRules.allow.file.length }}</span>
            </div>
            <div class="group-content">
              <RuleCard
                v-for="rule in groupedRules.allow.file"
                :key="rule.name"
                :rule="rule"
              />
            </div>
          </div>

          <!-- Network Allow Rules -->
          <div v-if="groupedRules.allow.connect.length > 0" class="rule-group">
            <div class="group-header">
              <div class="group-icon connect">
                <Globe :size="16" />
              </div>
              <h3 class="group-title">Network Connection</h3>
              <span class="group-count">{{ groupedRules.allow.connect.length }}</span>
            </div>
            <div class="group-content">
              <RuleCard
                v-for="rule in groupedRules.allow.connect"
                :key="rule.name"
                :rule="rule"
              />
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.rules-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* Header */
.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 16px;
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
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.title-icon {
  color: var(--accent-primary);
}

.page-subtitle {
  font-size: 14px;
  color: var(--text-muted);
}

.header-stats {
  display: flex;
  gap: 16px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
}

.stat-item.alert .stat-icon { color: var(--status-critical); }
.stat-item.allow .stat-icon { color: var(--status-safe); }
.stat-icon.exec { color: var(--status-info); }
.stat-icon.file { color: var(--status-safe); }
.stat-icon.connect { color: var(--status-warning); }

.stat-value {
  font-size: 16px;
  font-weight: 700;
  font-family: var(--font-mono);
  color: var(--text-primary);
}

.stat-label {
  font-size: 12px;
  color: var(--text-muted);
}

/* Filters */
.filters-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: var(--bg-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  flex-wrap: wrap;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
  flex: 1;
  min-width: 200px;
}

.search-box:focus-within {
  border-color: var(--border-focus);
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
}

.filter-icon {
  color: var(--text-muted);
}

.filter-select {
  padding: 8px 12px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-size: 12px;
  cursor: pointer;
}

.filter-select:focus {
  border-color: var(--border-focus);
  outline: none;
}

.refresh-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  transition: all var(--transition-fast);
}

.refresh-btn:hover:not(:disabled) {
  background: var(--bg-overlay);
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
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* Content */
.rules-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* Loading & Empty States */
.loading-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 40px;
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
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
  opacity: 0.5;
}

.empty-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.empty-description {
  font-size: 14px;
  color: var(--text-muted);
  max-width: 400px;
}

/* Rule Groups */
.rule-group {
  background: var(--bg-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  overflow: hidden;
}

.group-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 20px;
  background: var(--bg-elevated);
  border-bottom: 1px solid var(--border-subtle);
}

.group-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--radius-md);
}

.group-icon.exec {
  background: var(--status-info-dim);
  color: var(--status-info);
}

.group-icon.file {
  background: var(--status-safe-dim);
  color: var(--status-safe);
}

.group-icon.connect {
  background: var(--status-warning-dim);
  color: var(--status-warning);
}

.group-title {
  flex: 1;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.group-count {
  padding: 4px 12px;
  background: var(--bg-overlay);
  border-radius: var(--radius-full);
  font-size: 12px;
  font-weight: 600;
  font-family: var(--font-mono);
  color: var(--text-secondary);
}

.group-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 16px;
}

/* Rules Sections */
.rules-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border-radius: var(--radius-md);
  margin-bottom: 8px;
}

.section-header.alert {
  background: var(--status-critical-dim);
  color: var(--status-critical);
}

.section-header.allow {
  background: var(--status-safe-dim);
  color: var(--status-safe);
}

.section-header h2 {
  flex: 1;
  font-size: 14px;
  font-weight: 600;
  margin: 0;
}

.section-badge {
  padding: 4px 10px;
  background: rgba(255, 255, 255, 0.15);
  border-radius: var(--radius-full);
  font-size: 12px;
  font-weight: 600;
  font-family: var(--font-mono);
}
</style>
