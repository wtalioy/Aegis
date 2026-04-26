<!-- Settings Page - Phase 4 -->
<script setup lang="ts">
import { computed } from 'vue'
import { type ProviderField, useSettingsPage } from '../composables/useSettingsPage'
import Select from '../components/common/Select.vue'
import { CheckCircle2, Save, Loader2, RefreshCcw, AlertTriangle } from 'lucide-vue-next'

const {
  loading,
  saving,
  saveStatus,
  errorMessage,
  updateResult,
  form,
  providerOptions,
  activeProviderFields,
  saveSettings,
  getFieldValue,
  setFieldValue
} = useSettingsPage()

const activeProvider = computed(() => {
  if (form.mode === 'disabled') {
    return null
  }
  return form.mode
})

const updateProviderField = (field: ProviderField, value: string) => {
  if (!activeProvider.value) {
    return
  }
  setFieldValue(activeProvider.value, field.key, value)
}
</script>

<template>
  <div class="settings-page">
    <div class="page-header">
      <h1>Settings</h1>
      <p class="page-description">Configure AI provider and system settings</p>
    </div>

    <div v-if="loading" class="loading-state">
      <Loader2 :size="24" class="spinner" />
      <span>Loading settings...</span>
    </div>

    <div v-else class="settings-sections">
      <div class="settings-section">
        <h2>AI Configuration</h2>
        
        <div class="setting-item">
          <label>AI Provider</label>
          <Select
            v-model="form.mode"
            :options="providerOptions"
            placeholder="Select AI provider"
          />
        </div>

        <div v-if="activeProvider" class="provider-settings">
          <div v-for="field in activeProviderFields" :key="field.key" class="setting-item">
            <label>{{ field.label }}</label>
            <input
              :type="field.type"
              :min="field.min"
              :max="field.max"
              :placeholder="field.placeholder"
              :value="getFieldValue(activeProvider, field.key)"
              @input="updateProviderField(field, ($event.target as HTMLInputElement).value)"
            />
            <p class="setting-hint">{{ field.hint }}</p>
          </div>
        </div>
      </div>

      <div v-if="errorMessage" class="error-message">
        {{ errorMessage }}
      </div>

      <div v-if="updateResult && saveStatus === 'success'" class="update-summary">
        <div v-if="updateResult.hot_reloaded_fields.length > 0" class="update-summary-card hot-reload-card">
          <div class="summary-header">
            <RefreshCcw :size="16" />
            <span>Applied live</span>
          </div>
          <p>These settings took effect immediately:</p>
          <ul>
            <li v-for="field in updateResult.hot_reloaded_fields" :key="field">{{ field }}</li>
          </ul>
        </div>

        <div v-if="updateResult.restart_required" class="update-summary-card restart-card">
          <div class="summary-header">
            <AlertTriangle :size="16" />
            <span>Restart required</span>
          </div>
          <p>These settings were saved but need a restart:</p>
          <ul>
            <li v-for="field in updateResult.restart_required_fields" :key="field">{{ field }}</li>
          </ul>
        </div>
      </div>

      <div class="settings-actions">
        <button
          @click="saveSettings"
          :disabled="saving || loading"
          class="save-button"
          :class="{ 'is-saving': saving, 'is-success': saveStatus === 'success' }"
        >
          <Save v-if="!saving && saveStatus !== 'success'" :size="16" />
          <Loader2 v-if="saving" :size="16" class="spinner" />
          <CheckCircle2 v-if="saveStatus === 'success'" :size="16" />
          <span>{{ saving ? 'Saving...' : saveStatus === 'success' ? 'Saved!' : 'Save Settings' }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.settings-page {
  padding: 24px;
  max-width: 800px; /* Adjusted for better readability */
  margin: 0 auto;
}

.page-header h1 {
  font-size: 28px; /* Softened */
  font-weight: 600; /* Softened */
  color: var(--text-primary);
  margin: 0 0 8px 0;
}

.page-description {
  font-size: 14px;
  color: var(--text-secondary);
  margin: 0 0 32px 0;
}

.loading-state {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 48px;
  justify-content: center;
  color: var(--text-secondary);
}

.spinner {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.settings-sections {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.settings-section {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: 24px;
}

.settings-section h2 {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 24px 0; /* Increased margin */
}

.setting-item {
  margin-bottom: 24px; /* Increased margin */
}

.setting-item:last-child {
  margin-bottom: 0;
}

.setting-item label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.setting-item input[type="text"],
.setting-item input[type="number"],
.setting-item input[type="password"] {
  width: 100%;
  padding: 10px 12px;
  background: var(--bg-overlay);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  font-size: 14px;
  color: var(--text-primary);
  font-family: inherit;
  transition: all 0.2s;
}

.setting-item input:focus {
  outline: none;
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 3px var(--accent-glow);
}

.setting-hint {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 6px; /* Increased margin */
  margin-bottom: 0;
}

.provider-settings {
  margin-top: 24px; /* Increased margin */
  padding-top: 24px; /* Increased margin */
  border-top: 1px solid var(--border-subtle);
}

.error-message {
  padding: 12px 16px;
  background: var(--status-critical-dim);
  border: 1px solid var(--status-critical);
  border-radius: var(--radius-md);
  color: var(--status-critical);
  font-size: 14px;
}

.update-summary {
  display: grid;
  gap: 16px;
}

.update-summary-card {
  padding: 16px 18px;
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
}

.summary-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-size: 14px;
  font-weight: 600;
}

.hot-reload-card .summary-header {
  color: var(--status-safe);
}

.restart-card .summary-header {
  color: var(--status-warning);
}

.update-summary-card p {
  margin: 0 0 8px 0;
  color: var(--text-secondary);
  font-size: 14px;
}

.update-summary-card ul {
  margin: 0;
  padding-left: 18px;
}

.update-summary-card li {
  margin: 4px 0;
  color: var(--text-primary);
  font-size: 14px;
}

.settings-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 16px;
}

.save-button {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  background: var(--accent-primary);
  color: white;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.save-button:hover:not(:disabled) {
  background: var(--accent-primary-hover);
}

.save-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.save-button.is-success {
  background: var(--status-safe);
}

.save-button.is-success:hover:not(:disabled) {
  background: var(--status-safe);
  filter: brightness(1.1);
}
</style>
