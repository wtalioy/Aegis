// AI API Wrapper
// Separate AI-specific API functions from general api.ts

export interface RuleGenRequest {
  description: string
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

const API_BASE = '/api/ai'

export async function generateRule(req: RuleGenRequest): Promise<RuleGenResponse | null> {
  try {
    const response = await fetch(`${API_BASE}/generate-rule`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req)
    })
    if (!response.ok) return null
    return await response.json()
  } catch {
    return null
  }
}

export async function explainEvent(req: ExplainRequest): Promise<ExplainResponse | null> {
  try {
    const response = await fetch(`${API_BASE}/explain`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req)
    })
    if (!response.ok) return null
    return await response.json()
  } catch {
    return null
  }
}

export async function analyzeContext(req: AnalyzeRequest): Promise<AnalyzeResponse | null> {
  try {
    const response = await fetch(`${API_BASE}/analyze`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req)
    })
    if (!response.ok) return null
    return await response.json()
  } catch {
    return null
  }
}

export async function getSentinelInsights(limit = 50): Promise<any[]> {
  try {
    const response = await fetch(`${API_BASE}/sentinel/insights?limit=${limit}`)
    if (!response.ok) return []
    const data = await response.json()
    return data.insights || []
  } catch {
    return []
  }
}

