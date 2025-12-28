// AI Types

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

