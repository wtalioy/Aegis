<script setup lang="ts">
import { ref, computed } from 'vue'
import { Check, X, ChevronDown, ChevronRight, Search, Filter, Terminal, FileText, Globe } from 'lucide-vue-next'
import type { Rule } from '../../lib/api'

const props = defineProps<{
  rules: Rule[]
}>()

defineEmits<{
  apply: [selectedIndices: number[]]
  cancel: []
}>()

const selectedRules = ref<Set<number>>(new Set(
  props.rules.map((_, i) => i).filter(i => props.rules[i].selected)
))
const expandedRules = ref<Set<number>>(new Set())
const searchQuery = ref('')
const filterType = ref<string>('all')

const getRuleType = (rule: Rule): string => {
  if (rule.type) return rule.type
  const name = rule.name.toLowerCase()
  if (name.includes('file') || name.includes('access')) return 'file'
  if (name.includes('port') || name.includes('connect')) return 'connect'
  return 'exec'
}

const filteredRulesWithIndex = computed(() => {
  return props.rules
    .map((rule, index) => ({ rule, index }))
    .filter(({ rule }) => {
      if (searchQuery.value.trim()) {
        const query = searchQuery.value.toLowerCase()
        if (!rule.name.toLowerCase().includes(query) &&
            !rule.description.toLowerCase().includes(query)) {
          return false
        }
      }
      if (filterType.value !== 'all') {
        const type = getRuleType(rule)
        if (type !== filterType.value) return false
      }
      return true
    })
    .sort((a, b) => {
      const typeOrder = { exec: 0, file: 1, connect: 2 }
      const typeA = getRuleType(a.rule)
      const typeB = getRuleType(b.rule)
      return (typeOrder[typeA as keyof typeof typeOrder] ?? 3) - (typeOrder[typeB as keyof typeof typeOrder] ?? 3)
    })
})

const getTypeIcon = (rule: Rule) => {
  const type = getRuleType(rule)
  switch (type) {
    case 'exec': return Terminal
    case 'file': return FileText
    case 'connect': return Globe
    default: return Terminal
  }
}

const toggleRule = (index: number) => {
  if (selectedRules.value.has(index)) {
    selectedRules.value.delete(index)
  } else {
    selectedRules.value.add(index)
  }
  selectedRules.value = new Set(selectedRules.value)
}

const toggleExpand = (index: number) => {
  if (expandedRules.value.has(index)) {
    expandedRules.value.delete(index)
  } else {
    expandedRules.value.add(index)
  }
  expandedRules.value = new Set(expandedRules.value)
}

const selectAll = () => {
  filteredRulesWithIndex.value.forEach(({ index }) => {
    selectedRules.value.add(index)
  })
  selectedRules.value = new Set(selectedRules.value)
}

const deselectAll = () => {
  filteredRulesWithIndex.value.forEach(({ index }) => {
    selectedRules.value.delete(index)
  })
  selectedRules.value = new Set(selectedRules.value)
}

const selectedCount = computed(() => selectedRules.value.size)
</script>

<template>
  <div class="rule-review">
    <div class="review-header">
      <div class="header-title">
        <h2 class="title">Generated Rules</h2>
        <span class="rule-count">{{ rules.length }} patterns found</span>
      </div>
      <div class="header-actions">
        <button class="select-btn" @click="selectAll">Select All</button>
        <button class="select-btn" @click="deselectAll">Deselect All</button>
      </div>
    </div>

    <!-- Filters -->
    <div class="review-filters">
      <div class="filter-search">
        <Search :size="14" />
        <input 
          v-model="searchQuery"
          type="text"
          placeholder="Search rules..."
        />
      </div>
      <div class="filter-select">
        <Filter :size="14" />
        <select v-model="filterType">
          <option value="all">All Types</option>
          <option value="exec">Exec</option>
          <option value="file">File</option>
          <option value="connect">Network</option>
        </select>
      </div>
    </div>

    <!-- Rule List -->
    <div class="rule-list">
      <div 
        v-for="{ rule, index } in filteredRulesWithIndex"
        :key="index"
        class="rule-item"
        :class="{ selected: selectedRules.has(index) }"
      >
        <div class="rule-header" @click="toggleRule(index)">
          <div class="rule-checkbox" :class="{ checked: selectedRules.has(index) }">
            <Check v-if="selectedRules.has(index)" :size="12" />
          </div>
          <div class="rule-type-icon" :class="getRuleType(rule)">
            <component :is="getTypeIcon(rule)" :size="14" />
          </div>
          <div class="rule-info">
            <span class="rule-name">{{ rule.name }}</span>
            <span class="rule-desc">{{ rule.description }}</span>
          </div>
          <span class="rule-badge" :class="getRuleType(rule)">{{ getRuleType(rule).toUpperCase() }}</span>
          <button class="expand-btn" @click.stop="toggleExpand(index)">
            <ChevronDown v-if="expandedRules.has(index)" :size="16" />
            <ChevronRight v-else :size="16" />
          </button>
        </div>
        
        <Transition name="expand">
          <div v-if="expandedRules.has(index)" class="rule-yaml">
            <pre><code>{{ rule.yaml }}</code></pre>
          </div>
        </Transition>
      </div>

      <div v-if="filteredRulesWithIndex.length === 0" class="no-rules">
        No matching rules found
      </div>
    </div>

    <!-- Footer Actions -->
    <div class="review-footer">
      <div class="selection-info">
        Selected: <strong>{{ selectedCount }}</strong> / {{ rules.length }}
      </div>
      <div class="footer-actions">
        <button class="cancel-btn" @click="$emit('cancel')">
          <X :size="16" />
          Cancel
        </button>
        <button 
          class="apply-btn"
          :disabled="selectedCount === 0"
          @click="$emit('apply', Array.from(selectedRules))"
        >
          <Check :size="16" />
          Apply to Whitelist
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.rule-review {
  display: flex;
  flex-direction: column;
  background: var(--bg-surface);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-subtle);
  overflow: hidden;
}

.review-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  background: var(--bg-elevated);
  border-bottom: 1px solid var(--border-subtle);
}

.header-title {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.rule-count {
  font-size: 12px;
  color: var(--text-muted);
}

.header-actions {
  display: flex;
  gap: 8px;
}

.select-btn {
  padding: 6px 12px;
  border-radius: var(--radius-md);
  font-size: 12px;
  color: var(--text-secondary);
  background: var(--bg-surface);
  transition: all var(--transition-fast);
}

.select-btn:hover {
  background: var(--bg-overlay);
  color: var(--text-primary);
}

.review-filters {
  display: flex;
  gap: 12px;
  padding: 12px 20px;
  background: var(--bg-surface);
  border-bottom: 1px solid var(--border-subtle);
}

.filter-search {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  padding: 6px 10px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
  color: var(--text-muted);
}

.filter-search input {
  flex: 1;
  background: transparent;
  border: none;
  color: var(--text-primary);
  font-size: 12px;
  outline: none;
}

.filter-search input::placeholder {
  color: var(--text-muted);
}

.filter-select {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--text-muted);
}

.filter-select select {
  padding: 6px 10px;
  background: var(--bg-elevated);
  border: none;
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-size: 12px;
  cursor: pointer;
}

.rule-list {
  flex: 1;
  overflow-y: auto;
  max-height: 400px;
  padding: 12px;
}

.rule-item {
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
  margin-bottom: 8px;
  border: 1px solid transparent;
  transition: all var(--transition-fast);
}

.rule-item:last-child {
  margin-bottom: 0;
}

.rule-item.selected {
  border-color: var(--status-safe);
}

.rule-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  cursor: pointer;
}

.rule-checkbox {
  width: 18px;
  height: 18px;
  border-radius: var(--radius-sm);
  border: 2px solid var(--border-default);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-fast);
  flex-shrink: 0;
}

.rule-checkbox.checked {
  background: var(--status-safe);
  border-color: var(--status-safe);
  color: white;
}

.rule-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.rule-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rule-desc {
  font-size: 11px;
  color: var(--text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rule-type-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: var(--radius-sm);
  flex-shrink: 0;
}

.rule-type-icon.exec {
  background: var(--status-info-dim);
  color: var(--status-info);
}

.rule-type-icon.file {
  background: var(--status-safe-dim);
  color: var(--status-safe);
}

.rule-type-icon.connect {
  background: var(--status-warning-dim);
  color: var(--status-warning);
}

.rule-badge {
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  font-size: 10px;
  font-weight: 600;
  flex-shrink: 0;
}

.rule-badge.exec {
  background: var(--status-info-dim);
  color: var(--status-info);
}

.rule-badge.file {
  background: var(--status-safe-dim);
  color: var(--status-safe);
}

.rule-badge.connect {
  background: var(--status-warning-dim);
  color: var(--status-warning);
}

.expand-btn {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  border-radius: var(--radius-sm);
  flex-shrink: 0;
}

.expand-btn:hover {
  background: var(--bg-overlay);
  color: var(--text-primary);
}

.rule-yaml {
  padding: 0 12px 12px;
}

.rule-yaml pre {
  margin: 0;
  padding: 12px;
  background: var(--bg-void);
  border-radius: var(--radius-md);
  overflow-x: auto;
}

.rule-yaml code {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-secondary);
  white-space: pre;
}

.expand-enter-active,
.expand-leave-active {
  transition: all 0.2s ease;
}

.expand-enter-from,
.expand-leave-to {
  opacity: 0;
  max-height: 0;
}

.no-rules {
  text-align: center;
  padding: 32px;
  color: var(--text-muted);
  font-size: 13px;
}

.review-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  background: var(--bg-elevated);
  border-top: 1px solid var(--border-subtle);
}

.selection-info {
  font-size: 13px;
  color: var(--text-secondary);
}

.selection-info strong {
  color: var(--text-primary);
}

.footer-actions {
  display: flex;
  gap: 12px;
}

.cancel-btn,
.apply-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 500;
  transition: all var(--transition-fast);
}

.cancel-btn {
  background: var(--bg-surface);
  color: var(--text-secondary);
}

.cancel-btn:hover {
  background: var(--bg-overlay);
  color: var(--text-primary);
}

.apply-btn {
  background: var(--status-safe);
  color: white;
}

.apply-btn:hover:not(:disabled) {
  filter: brightness(1.1);
}

.apply-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>

