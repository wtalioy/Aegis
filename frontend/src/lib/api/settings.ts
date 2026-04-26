import { requestJSON } from '../http'

const API_BASE = '/api/v1/system/settings'

export interface AISettings {
  mode: 'ollama' | 'openai' | 'gemini'
  ollama: {
    endpoint: string
    model: string
    timeout: number
  }
  openai: {
    endpoint: string
    apiKey: string
    model: string
    timeout: number
  }
  gemini: {
    endpoint: string
    apiKey: string
    model: string
    timeout: number
  }
}

export interface Settings {
  server: {
    port: number
  }
  kernel: {
    bpf_path: string
    ring_buffer_size: number
  }
  telemetry: {
    process_tree_max_age: string
    process_tree_max_size: number
    process_tree_max_chain_length: number
    recent_events_capacity: number
    event_index_size: number
  }
  policy: {
    rules_path: string
    promotion_min_observation_minutes: number
    promotion_min_hits: number
  }
  analysis: {
    mode: 'ollama' | 'openai' | 'gemini' | 'disabled'
    ollama: {
      endpoint: string
      model: string
      timeout: number
    }
    openai: {
      endpoint: string
      api_key: string
      model: string
      timeout: number
    }
    gemini: {
      endpoint: string
      api_key: string
      model: string
      timeout: number
    }
  }
  sentinel: {
    testing_promotion: string
    anomaly: string
    rule_optimization: string
    daily_report: string
  }
}

export interface UpdateSettingsResult {
  updated: boolean
  hot_reloaded_fields: string[]
  restart_required: boolean
  restart_required_fields: string[]
  config: Settings
}

export async function getSettings(): Promise<Settings> {
  return requestJSON<Settings>(API_BASE)
}

export async function updateSettings(settings: Settings): Promise<UpdateSettingsResult> {
  return requestJSON<UpdateSettingsResult>(API_BASE, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(settings)
  })
}
