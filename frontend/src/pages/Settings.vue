<!-- Settings Page - Phase 4 -->
<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { getSettings, updateSettings, type Settings } from '../lib/api'
import Select from '../components/common/Select.vue'
import { Save, Loader2 } from 'lucide-vue-next'

const settings = ref<Settings | null>(null)
const loading = ref(true)
const saving = ref(false)
const saveStatus = ref<'idle' | 'success' | 'error'>('idle')
const errorMessage = ref('')

// Form state
const aiProvider = ref<'ollama' | 'openai'>('ollama')
const ollamaEndpoint = ref('')
const ollamaModel = ref('')
const ollamaTimeout = ref(60)
const openaiEndpoint = ref('')
const openaiApiKey = ref('')
const openaiModel = ref('')
const openaiTimeout = ref(30)

const providerOptions = [
  { value: 'ollama', label: 'Ollama' },
  { value: 'openai', label: 'OpenAI' }
]

const showOpenAIFields = computed(() => aiProvider.value === 'openai')
const showOllamaFields = computed(() => aiProvider.value === 'ollama')

// Load settings on mount
onMounted(async () => {
  try {
    const data = await getSettings()
    settings.value = data
    
    // Populate form
    aiProvider.value = data.ai.mode
    ollamaEndpoint.value = data.ai.ollama.endpoint
    ollamaModel.value = data.ai.ollama.model
    ollamaTimeout.value = data.ai.ollama.timeout
    openaiEndpoint.value = data.ai.openai.endpoint
    openaiApiKey.value = data.ai.openai.apiKey
    openaiModel.value = data.ai.openai.model
    openaiTimeout.value = data.ai.openai.timeout
  } catch (err) {
    console.error('Failed to load settings:', err)
    errorMessage.value = err instanceof Error ? err.message : 'Failed to load settings'
  } finally {
    loading.value = false
  }
})

// Save settings
const saveSettings = async () => {
  if (!settings.value) return
  
  saving.value = true
  saveStatus.value = 'idle'
  errorMessage.value = ''
  
  try {
    const updated: Settings = {
      ai: {
        mode: aiProvider.value,
        ollama: {
          endpoint: ollamaEndpoint.value,
          model: ollamaModel.value,
          timeout: ollamaTimeout.value
        },
        openai: {
          endpoint: openaiEndpoint.value,
          apiKey: openaiApiKey.value,
          model: openaiModel.value,
          timeout: openaiTimeout.value
        }
      }
    }
    
    await updateSettings(updated)
    settings.value = updated
    saveStatus.value = 'success'
    
    // Clear success message after 3 seconds
    setTimeout(() => {
      saveStatus.value = 'idle'
    }, 3000)
  } catch (err) {
    console.error('Failed to save settings:', err)
    errorMessage.value = err instanceof Error ? err.message : 'Failed to save settings'
    saveStatus.value = 'error'
  } finally {
    saving.value = false
  }
}

// Auto-save on provider change (optional - you can remove this if you prefer manual save only)
watch(aiProvider, () => {
  // Optionally auto-save when provider changes
  // saveSettings()
})
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
            v-model="aiProvider"
            :options="providerOptions"
            placeholder="Select AI provider"
          />
        </div>

        <!-- Ollama Settings -->
        <div v-if="showOllamaFields" class="provider-settings">
          <div class="setting-item">
            <label>Ollama Endpoint</label>
            <input
              v-model="ollamaEndpoint"
              type="text"
              placeholder="http://localhost:11434"
            />
            <p class="setting-hint">URL where Ollama API is running</p>
          </div>
          
          <div class="setting-item">
            <label>Model</label>
            <input
              v-model="ollamaModel"
              type="text"
              placeholder="qwen2.5-coder:1.5b"
            />
            <p class="setting-hint">Model name to use (e.g., llama3, qwen2.5-coder:1.5b)</p>
          </div>
          
          <div class="setting-item">
            <label>Timeout (seconds)</label>
            <input
              v-model.number="ollamaTimeout"
              type="number"
              min="10"
              max="300"
            />
            <p class="setting-hint">Request timeout in seconds</p>
          </div>
        </div>

        <!-- OpenAI Settings -->
        <div v-if="showOpenAIFields" class="provider-settings">
          <div class="setting-item">
            <label>Base URL</label>
            <input
              v-model="openaiEndpoint"
              type="text"
              placeholder="https://api.deepseek.com"
            />
            <p class="setting-hint">API endpoint URL (e.g., https://api.openai.com/v1 or https://api.deepseek.com)</p>
          </div>
          
          <div class="setting-item">
            <label>API Key</label>
            <input
              v-model="openaiApiKey"
              type="password"
              placeholder="sk-..."
            />
            <p class="setting-hint">Your API key for authentication</p>
          </div>
          
          <div class="setting-item">
            <label>Model</label>
            <input
              v-model="openaiModel"
              type="text"
              placeholder="deepseek-chat"
            />
            <p class="setting-hint">Model name to use (e.g., gpt-4, deepseek-chat)</p>
          </div>
          
          <div class="setting-item">
            <label>Timeout (seconds)</label>
            <input
              v-model.number="openaiTimeout"
              type="number"
              min="10"
              max="300"
            />
            <p class="setting-hint">Request timeout in seconds</p>
          </div>
        </div>
      </div>

      <div v-if="errorMessage" class="error-message">
        {{ errorMessage }}
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