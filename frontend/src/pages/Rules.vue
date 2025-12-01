<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { 
  FileCode, Terminal, Globe, FileText, RefreshCw, Search, Filter, 
  ShieldOff, AlertTriangle, ShieldCheck, Zap
} from 'lucide-vue-next'
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
  const blockRules: Rule[] = []
  const alertRules: Rule[] = []
  const allowRules: Rule[] = []
  
  filteredRules.value.forEach(rule => {
    if (rule.action === 'block') {
      blockRules.push(rule)
    } else if (rule.action === 'allow') {
      allowRules.push(rule)
    } else {
      alertRules.push(rule)
    }
  })

  return { block: blockRules, alert: alertRules, allow: allowRules }
})

const stats = computed(() => ({
  total: rules.value.length,
  block: rules.value.filter(r => r.action === 'block').length,
  alert: rules.value.filter(r => r.action === 'alert' || r.action === 'log').length,
  allow: rules.value.filter(r => r.action === 'allow').length,
}))

const typeStats = computed(() => ({
  exec: filteredRules.value.filter(r => deriveRuleType(r) === 'exec').length,
  file: filteredRules.value.filter(r => deriveRuleType(r) === 'file').length,
  connect: filteredRules.value.filter(r => deriveRuleType(r) === 'connect').length,
}))

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
          Security Rules
        </h1>
        <span class="page-subtitle">LSM enforcement and detection rules</span>
      </div>
      
      <!-- Action Stats -->
      <div class="action-stats">
        <button 
          class="action-stat block" 
          :class="{ active: filterAction === 'block' }"
          @click="filterAction = filterAction === 'block' ? 'all' : 'block'"
        >
          <ShieldOff :size="16" />
          <span class="stat-value">{{ stats.block }}</span>
          <span class="stat-label">Block</span>
        </button>
        <button 
          class="action-stat alert"
          :class="{ active: filterAction === 'alert' }"
          @click="filterAction = filterAction === 'alert' ? 'all' : 'alert'"
        >
          <AlertTriangle :size="16" />
          <span class="stat-value">{{ stats.alert }}</span>
          <span class="stat-label">Alert</span>
        </button>
        <button 
          class="action-stat allow"
          :class="{ active: filterAction === 'allow' }"
          @click="filterAction = filterAction === 'allow' ? 'all' : 'allow'"
        >
          <ShieldCheck :size="16" />
          <span class="stat-value">{{ stats.allow }}</span>
          <span class="stat-label">Allow</span>
        </button>
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
        
        <select v-model="filterType" class="filter-select">
          <option value="all">All Types</option>
          <option value="exec">Exec ({{ typeStats.exec }})</option>
          <option value="file">File ({{ typeStats.file }})</option>
          <option value="connect">Network ({{ typeStats.connect }})</option>
        </select>

        <select v-model="filterSeverity" class="filter-select">
          <option value="all">All Severity</option>
          <option value="critical">Critical</option>
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

    <!-- Type Legend -->
    <div class="type-legend">
      <div class="legend-item">
        <Terminal :size="14" class="exec" />
        <span>Process Exec</span>
      </div>
      <div class="legend-item">
        <FileText :size="14" class="file" />
        <span>File Access</span>
      </div>
      <div class="legend-item">
        <Globe :size="14" class="network" />
        <span>Network</span>
      </div>
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
        <FileCode :size="48" class="empty-icon" />
        <div class="empty-title">No Rules Loaded</div>
        <div class="empty-description">
          Add rules to your rules.yaml file and restart the application.
        </div>
      </div>

      <!-- No Matches -->
      <div v-else-if="filteredRules.length === 0" class="empty-state">
        <Search :size="48" class="empty-icon" />
        <div class="empty-title">No Matching Rules</div>
        <div class="empty-description">
          Try adjusting your search or filters to find rules.
        </div>
      </div>

      <!-- Rule Sections -->
      <template v-else>
        <!-- Block Rules -->
        <div v-if="groupedRules.block.length > 0" class="rules-section">
          <div class="section-header block">
            <ShieldOff :size="18" />
            <h2>Block Rules</h2>
            <span class="section-desc">Active defense - operations will be denied</span>
            <span class="section-count">{{ groupedRules.block.length }}</span>
          </div>
          <div class="rules-list">
            <RuleCard v-for="rule in groupedRules.block" :key="rule.name" :rule="rule" />
          </div>
        </div>

        <!-- Alert Rules -->
        <div v-if="groupedRules.alert.length > 0" class="rules-section">
          <div class="section-header alert">
            <AlertTriangle :size="18" />
            <h2>Alert Rules</h2>
            <span class="section-desc">Passive monitoring - events will be logged</span>
            <span class="section-count">{{ groupedRules.alert.length }}</span>
          </div>
          <div class="rules-list">
            <RuleCard v-for="rule in groupedRules.alert" :key="rule.name" :rule="rule" />
          </div>
        </div>

        <!-- Allow Rules -->
        <div v-if="groupedRules.allow.length > 0" class="rules-section">
          <div class="section-header allow">
            <ShieldCheck :size="18" />
            <h2>Allow Rules</h2>
            <span class="section-desc">Whitelist - bypass monitoring</span>
            <span class="section-count">{{ groupedRules.allow.length }}</span>
          </div>
          <div class="rules-list">
            <RuleCard v-for="rule in groupedRules.allow" :key="rule.name" :rule="rule" />
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
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.title-icon {
  color: var(--accent-primary);
}

.page-subtitle {
  font-size: 13px;
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
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.action-stat:hover {
  border-color: var(--border-default);
}

.action-stat.active {
  background: var(--bg-overlay);
}

.action-stat.block {
  color: var(--status-blocked);
}

.action-stat.block.active {
  border-color: var(--status-blocked);
  background: var(--status-blocked-dim);
}

.action-stat.alert {
  color: var(--status-warning);
}

.action-stat.alert.active {
  border-color: var(--status-warning);
  background: var(--status-warning-dim);
}

.action-stat.allow {
  color: var(--status-safe);
}

.action-stat.allow.active {
  border-color: var(--status-safe);
  background: var(--status-safe-dim);
}

.action-stat .stat-value {
  font-size: 18px;
  font-weight: 700;
  font-family: var(--font-mono);
}

.action-stat .stat-label {
  font-size: 12px;
  opacity: 0.8;
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
  border-color: var(--accent-primary);
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
  border-color: var(--accent-primary);
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

/* Type Legend */
.type-legend {
  display: flex;
  gap: 20px;
  padding: 10px 16px;
  background: var(--bg-surface);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-muted);
}

.legend-item .exec {
  color: var(--chart-exec);
}

.legend-item .file {
  color: var(--chart-file);
}

.legend-item .network {
  color: var(--chart-network);
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
  background: var(--bg-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
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
  color: var(--text-muted);
  opacity: 0.5;
  margin-bottom: 16px;
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

/* Rule Sections */
.rules-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 18px;
  border-radius: var(--radius-md);
}

.section-header.block {
  background: linear-gradient(135deg, var(--status-blocked-dim), transparent 60%);
  border-left: 3px solid var(--status-blocked);
  color: var(--status-blocked);
}

.section-header.alert {
  background: linear-gradient(135deg, var(--status-warning-dim), transparent 60%);
  border-left: 3px solid var(--status-warning);
  color: var(--status-warning);
}

.section-header.allow {
  background: linear-gradient(135deg, var(--status-safe-dim), transparent 60%);
  border-left: 3px solid var(--status-safe);
  color: var(--status-safe);
}

.section-header h2 {
  font-size: 14px;
  font-weight: 600;
  margin: 0;
}

.section-desc {
  flex: 1;
  font-size: 12px;
  opacity: 0.7;
  color: var(--text-secondary);
}

.section-count {
  padding: 4px 12px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-full);
  font-size: 12px;
  font-weight: 600;
  font-family: var(--font-mono);
}

.rules-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
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

  .type-legend {
    justify-content: center;
  }

  .section-desc {
    display: none;
  }
}
</style>
