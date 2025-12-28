// AI Core Composable
import { ref } from 'vue'
import type { Ref } from 'vue'

export interface RuleGenRequest {
  description: string
  context?: {
    currentPage?: string
    selectedItem?: string
    recentActions?: string[]
  }
  examples?: any[]
}

export interface RuleGenResponse {
  rule: any
  yaml: string
  reasoning: string
  confidence: number
  warnings: string[]
  simulation?: any
}

export interface ExplainRequest {
  eventId?: string
  eventData?: any
  question?: string
}

export interface ExplainResponse {
  explanation: string
  rootCause: string
  matchedRule?: any
  relatedEvents?: any[]
  suggestedActions?: any[]
}

export interface AnalyzeRequest {
  type: 'process' | 'workload' | 'rule'
  id: string
}

export interface AnalyzeResponse {
  summary: string
  anomalies: any[]
  baselineStatus: string
  recommendations: any[]
  relatedInsights: any[]
}

const API_BASE = '/api'

export function useAI() {
  const loading = ref(false)
  const error = ref<string | null>(null)

  const generateRule = async (req: RuleGenRequest): Promise<RuleGenResponse | null> => {
    loading.value = true
    error.value = null
    try {
      const response = await fetch(`${API_BASE}/ai/generate-rule`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(req)
      })
      const text = await response.text()
      let payload: any = null
      try { payload = text ? JSON.parse(text) : null } catch { payload = null }
      if (!response.ok) {
        error.value = (payload && (payload.error || payload.message)) || `HTTP ${response.status}`
        return null
      }
      return payload as RuleGenResponse
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to generate rule'
      return null
    } finally {
      loading.value = false
    }
  }

  const explainEvent = async (req: ExplainRequest): Promise<ExplainResponse | null> => {
    loading.value = true
    error.value = null
    try {
      const response = await fetch(`${API_BASE}/ai/explain`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(req)
      })
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`)
      }
      const data = await response.json()
      return data
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to explain event'
      return null
    } finally {
      loading.value = false
    }
  }

  const analyzeContext = async (req: AnalyzeRequest): Promise<AnalyzeResponse | null> => {
    loading.value = true
    error.value = null
    try {
      const response = await fetch(`${API_BASE}/ai/analyze`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(req)
      })
      const text = await response.text()
      let payload: any = null
      try { payload = text ? JSON.parse(text) : null } catch { payload = null }
      if (!response.ok) {
        error.value = (payload && (payload.error || payload.message)) || `HTTP ${response.status}`
        return null
      }
      return payload as AnalyzeResponse
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to analyze context'
      return null
    } finally {
      loading.value = false
    }
  }

  const getAIStatus = async (): Promise<any> => {
    try {
      const res = await fetch(`${API_BASE}/ai/status`)
      const text = await res.text()
      try { return text ? JSON.parse(text) : null } catch { return null }
    } catch {
      return null
    }
  }

  return {
    loading,
    error,
    generateRule,
    explainEvent,
    analyzeContext,
    getAIStatus
  }
}

