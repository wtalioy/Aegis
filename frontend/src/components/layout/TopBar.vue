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
      <!-- Left side empty for now -->
    </div>

    <div class="topbar-center">
      <!-- AI Omnibox Trigger - Prominent and Centered -->
      <button v-if="aiStatus?.status === 'ready'" class="omnibox-trigger prominent" @click="toggleOmnibox"
        title="Open AI Omnibox (Cmd/Ctrl+K)">
        <Sparkles :size="18" />
        <span>Ask Aegis anything...</span>
        <kbd>Cmd/Ctrl+K</kbd>
      </button>
    </div>

    <div class="topbar-right">
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
  gap: 16px;
}

.topbar-center {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  max-width: 600px;
  margin: 0 auto;
}

.diagnose-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.diagnose-btn:hover:not(:disabled) {
  background: var(--bg-hover);
  color: var(--text-primary);
  border-color: var(--border-default);
}

.diagnose-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.omnibox-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.omnibox-trigger.prominent {
  width: 100%;
  max-width: 500px;
  padding: 10px 16px;
  background: var(--bg-surface);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-lg);
  font-size: 14px;
  justify-content: flex-start;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.omnibox-trigger:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.omnibox-trigger.prominent:hover {
  background: var(--bg-elevated);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.omnibox-trigger.prominent span {
  flex: 1;
  text-align: left;
  color: var(--text-muted);
}

.omnibox-trigger.prominent:hover span {
  color: var(--text-primary);
}

.omnibox-trigger kbd {
  padding: 2px 6px;
  background: var(--bg-void);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-sm);
  font-family: monospace;
  font-size: 10px;
  color: var(--text-muted);
}
</style>
