import { computed, onMounted, reactive, ref } from 'vue'
import { getSettings, updateSettings, type Settings, type UpdateSettingsResult } from '../lib/api'

type ProviderKey = 'ollama' | 'openai' | 'gemini'
type ProviderFieldKey = 'endpoint' | 'api_key' | 'model' | 'timeout'
type ProviderConfig = {
  endpoint: string
  model: string
  timeout: number
  api_key?: string
}

export interface ProviderField {
  key: ProviderFieldKey
  label: string
  type: 'text' | 'password' | 'number'
  placeholder: string
  hint: string
  min?: number
  max?: number
}

export function useSettingsPage() {
  const settings = ref<Settings | null>(null)
  const loading = ref(true)
  const saving = ref(false)
  const saveStatus = ref<'idle' | 'success' | 'error'>('idle')
  const errorMessage = ref('')
  const updateResult = ref<UpdateSettingsResult | null>(null)

  const form = reactive<Settings['analysis']>({
    mode: 'ollama',
    ollama: {
      endpoint: '',
      model: '',
      timeout: 60
    },
    openai: {
      endpoint: '',
      api_key: '',
      model: '',
      timeout: 30
    },
    gemini: {
      endpoint: '',
      api_key: '',
      model: '',
      timeout: 30
    }
  })

  const providerOptions = [
    { value: 'disabled', label: 'Disabled' },
    { value: 'ollama', label: 'Ollama' },
    { value: 'openai', label: 'OpenAI-Compatible' },
    { value: 'gemini', label: 'Gemini' }
  ]

  const providerFields: Record<ProviderKey, ProviderField[]> = {
    ollama: [
      { key: 'endpoint', label: 'Ollama Endpoint', type: 'text', placeholder: 'http://localhost:11434', hint: 'URL where Ollama API is running' },
      { key: 'model', label: 'Model', type: 'text', placeholder: 'qwen2.5-coder:1.5b', hint: 'Model name to use (e.g., llama3, qwen2.5-coder:1.5b)' },
      { key: 'timeout', label: 'Timeout (seconds)', type: 'number', placeholder: '', hint: 'Request timeout in seconds', min: 10, max: 300 }
    ],
    openai: [
      { key: 'endpoint', label: 'Base URL', type: 'text', placeholder: 'https://api.deepseek.com', hint: 'API endpoint URL (e.g., https://api.openai.com/v1 or https://api.deepseek.com)' },
      { key: 'api_key', label: 'API Key', type: 'password', placeholder: 'sk-...', hint: 'Your API key for authentication' },
      { key: 'model', label: 'Model', type: 'text', placeholder: 'deepseek-chat', hint: 'Model name to use (e.g., gpt-4, deepseek-chat)' },
      { key: 'timeout', label: 'Timeout (seconds)', type: 'number', placeholder: '', hint: 'Request timeout in seconds', min: 10, max: 300 }
    ],
    gemini: [
      { key: 'endpoint', label: 'Base URL', type: 'text', placeholder: 'https://generativelanguage.googleapis.com', hint: 'Gemini API base URL. Aegis appends the model route automatically.' },
      { key: 'api_key', label: 'API Key', type: 'password', placeholder: 'AIza...', hint: 'Your Google AI Studio or Gemini API key.' },
      { key: 'model', label: 'Model', type: 'text', placeholder: 'gemini-3-flash-preview', hint: 'Model name only, for example gemini-3-flash-preview.' },
      { key: 'timeout', label: 'Timeout (seconds)', type: 'number', placeholder: '', hint: 'Request timeout in seconds', min: 10, max: 300 }
    ]
  }

  const activeProviderFields = computed(() => {
    if (form.mode === 'disabled') {
      return [] as ProviderField[]
    }
    return providerFields[form.mode]
  })

  const loadForm = (analysis: Settings['analysis']) => {
    form.mode = analysis.mode
    form.ollama = { ...analysis.ollama }
    form.openai = { ...analysis.openai }
    form.gemini = { ...analysis.gemini }
  }

  const buildSettingsPayload = (): Settings | null => {
    if (!settings.value) {
      return null
    }
    return {
      ...settings.value,
      analysis: {
        mode: form.mode,
        ollama: { ...form.ollama },
        openai: { ...form.openai },
        gemini: { ...form.gemini }
      }
    }
  }

  const getProviderConfig = (provider: ProviderKey): ProviderConfig => form[provider] as ProviderConfig

  const getFieldValue = (provider: ProviderKey, key: ProviderFieldKey): string | number => getProviderConfig(provider)[key] ?? ''

  const setFieldValue = (provider: ProviderKey, key: ProviderFieldKey, value: string) => {
    const config = getProviderConfig(provider)
    switch (key) {
      case 'endpoint':
        config.endpoint = value
        break
      case 'api_key':
        config.api_key = value
        break
      case 'model':
        config.model = value
        break
      case 'timeout':
        config.timeout = Number(value)
        break
    }
  }

  const saveSettings = async () => {
    const updated = buildSettingsPayload()
    if (!updated) {
      return
    }

    saving.value = true
    saveStatus.value = 'idle'
    errorMessage.value = ''
    updateResult.value = null

    try {
      const result = await updateSettings(updated)
      updateResult.value = result
      settings.value = result.config
      loadForm(result.config.analysis)
      saveStatus.value = 'success'
      setTimeout(() => {
        saveStatus.value = 'idle'
      }, 3000)
    } catch (error) {
      console.error('Failed to save settings:', error)
      errorMessage.value = error instanceof Error ? error.message : 'Failed to save settings'
      saveStatus.value = 'error'
    } finally {
      saving.value = false
    }
  }

  onMounted(async () => {
    try {
      const data = await getSettings()
      settings.value = data
      loadForm(data.analysis)
    } catch (error) {
      console.error('Failed to load settings:', error)
      errorMessage.value = error instanceof Error ? error.message : 'Failed to load settings'
    } finally {
      loading.value = false
    }
  })

  return {
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
  }
}
