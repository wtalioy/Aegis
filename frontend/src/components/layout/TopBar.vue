<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Zap, Sparkles } from 'lucide-vue-next'
import { getAIStatus, type AIStatus } from '../../lib/api'
import DiagnosisModal from '../ai/DiagnosisModal.vue'
import AIOmnibox from '../ai/AIOmnibox.vue'
import { useOmnibox } from '../../composables/useOmnibox'

// AI state
const aiStatus = ref<AIStatus | null>(null)
const showDiagnosisModal = ref(false)

// Omnibox
const { toggle: toggleOmnibox } = useOmnibox()

onMounted(async () => {
  // Fetch AI status
  try {
    aiStatus.value = await getAIStatus()
  } catch (e) {
    console.error('Failed to fetch AI status:', e)
  }
})
</script>

<template>
  <header class="topbar">
    <div class="topbar-left">
      <!-- Logo or breadcrumbs can go here -->
    </div>

    <div class="topbar-right">
      <!-- AI Omnibox Trigger -->
      <button v-if="aiStatus?.status === 'ready'" class="omnibox-trigger" @click="toggleOmnibox"
        title="Open AI Omnibox (Cmd/Ctrl+K)">
        <Sparkles :size="16" />
        <span>Ask Aegis...</span>
        <kbd>⌘K</kbd>
      </button>

      <!-- Quick Diagnose Button -->
      <button v-if="aiStatus?.status === 'ready'" class="diagnose-btn" @click="showDiagnosisModal = true"
        :disabled="aiStatus.status !== 'ready'" title="Quick one-click system diagnosis">
        <Zap :size="15" />
        <span>Quick Diagnose</span>
      </button>
    </div>
  </header>

  <!-- Quick Diagnosis Modal -->
  <DiagnosisModal :visible="showDiagnosisModal" @close="showDiagnosisModal = false" />

  <!-- AI Omnibox -->
  <AIOmnibox />
</template>

<style scoped>
.topbar {
  height: var(--topbar-height);
  background: var(--bg-surface);
  border-bottom: 1px solid var(--border-subtle);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
}

.topbar-left,
.topbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.diagnose-btn,
.omnibox-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  background: transparent;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast) ease;
}

.diagnose-btn:hover:not(:disabled),
.omnibox-trigger:hover:not(:disabled) {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.diagnose-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.omnibox-trigger kbd {
  padding: 2px 6px;
  background: var(--bg-overlay);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--text-muted);
}
</style>