<!-- Rule List Component - Phase 4 -->
<script setup lang="ts">
import { ref, computed } from 'vue'
import { Search, Plus } from 'lucide-vue-next'
import type { Rule } from '../../types/rules'

const props = defineProps<{
  rules: Rule[]
  loading?: boolean
}>()

const emit = defineEmits<{
  select: [rule: Rule]
  create: []
  promote: [rule: Rule]
}>()

const searchQuery = ref('')
const filterMode = ref<'all' | 'production' | 'testing' | 'draft'>('all')
const filterAction = ref<'all' | 'block' | 'monitor' | 'allow'>('all')

const filteredRules = computed(() => {
  let result = props.rules

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(r => 
      r.name.toLowerCase().includes(query) ||
      r.description.toLowerCase().includes(query)
    )
  }

  if (filterMode.value !== 'all') {
    result = result.filter(r => (r as any).state === filterMode.value)
  }

  if (filterAction.value !== 'all') {
    result = result.filter(r => r.action === filterAction.value)
  }

  return result
})

const ruleCounts = computed(() => {
  return {
    total: props.rules.length,
    production: props.rules.filter(r => (r as any).state === 'production').length,
    testing: props.rules.filter(r => (r as any).state === 'testing').length,
    draft: props.rules.filter(r => (r as any).state === 'draft').length
  }
})
</script>

<template>
  <div class="rule-list">
    <div class="list-header">
      <h3>Rules</h3>
      <button class="create-btn" @click="$emit('create')">
        <Plus :size="16" />
        <span>Create</span>
      </button>
    </div>

    <div class="filters">
      <div class="search-box">
        <Search :size="16" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search rules..."
          class="search-input"
        />
      </div>
    </div>

    <div v-if="loading" class="loading-state">
      Loading rules...
    </div>

    <div v-else-if="filteredRules.length === 0" class="empty-state">
      No rules found
    </div>

    <div v-else class="rules-container">
      <div
        v-for="rule in filteredRules"
        :key="rule.name"
        class="rule-item"
        @click="$emit('select', rule)"
      >
        <div class="rule-header">
          <div class="rule-name">{{ rule.name }}</div>
          <div class="rule-badges">
            <span class="mode-badge" :class="(rule as any).state">{{ (rule as any).state }}</span>
            <span class="action-badge" :class="rule.action">{{ rule.action }}</span>
          </div>
        </div>
        <div class="rule-description">{{ rule.description }}</div>
        <div class="rule-footer">
          <span class="rule-severity">{{ rule.severity }}</span>
          <button
            v-if="(rule as any).state === 'testing'"
            class="promote-btn"
            @click.stop="$emit('promote', rule)"
          >
            Promote
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.rule-list {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  padding: 20px;
}

.list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.list-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.create-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  background: rgba(34, 197, 94, 0.15);
  color: rgb(34, 197, 94);
  border: 1px solid rgba(34, 197, 94, 0.3);
  border-radius: var(--radius-md);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.create-btn:hover {
  background: rgba(34, 197, 94, 0.25);
  border-color: rgba(34, 197, 94, 0.5);
}

.filters {
  margin-bottom: 20px;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
}

.search-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  font-size: 14px;
  color: var(--text-primary);
}


.rules-container {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-right: 4px;
}

.rule-item {
  padding: 14px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.15s;
}

.rule-item:hover {
  background: var(--bg-hover);
  border-color: var(--border-default);
}

.rule-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
  gap: 10px;
}

.rule-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rule-badges {
  display: flex;
  gap: 6px;
}

.mode-badge,
.action-badge {
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
}

.mode-badge.production {
  background: rgba(34, 197, 94, 0.1);
  color: rgb(34, 197, 94);
}

.mode-badge.testing {
  background: rgba(251, 191, 36, 0.1);
  color: rgb(251, 191, 36);
}

.mode-badge.draft {
  background: rgba(156, 163, 175, 0.1);
  color: rgb(156, 163, 175);
}

.action-badge.block {
  background: rgba(239, 68, 68, 0.1);
  color: rgb(239, 68, 68);
}

.action-badge.monitor {
  background: rgba(59, 130, 246, 0.1);
  color: rgb(59, 130, 246);
}

.action-badge.allow {
  background: rgba(34, 197, 94, 0.1);
  color: rgb(34, 197, 94);
}

.rule-description {
  font-size: 12px;
  line-height: 1.5;
  color: var(--text-secondary);
  margin-bottom: 10px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.rule-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.rule-severity {
  font-size: 12px;
  color: var(--text-muted);
  text-transform: capitalize;
}

.promote-btn {
  padding: 6px 12px;
  background: rgba(34, 197, 94, 0.15);
  border: 1px solid rgba(34, 197, 94, 0.3);
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 500;
  color: rgb(34, 197, 94);
  cursor: pointer;
  transition: all 0.15s;
}

.promote-btn:hover {
  background: rgba(34, 197, 94, 0.25);
  border-color: rgba(34, 197, 94, 0.5);
}

.loading-state,
.empty-state {
  padding: 40px;
  text-align: center;
  color: var(--text-muted);
}
</style>

