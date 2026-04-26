import type { Rule } from './rules'
import type { Insight } from './sentinel'

export interface RuleGenRequest {
  description: string
  context?: {
    currentPage?: string
    selectedItem?: string
    recentActions?: string[]
  }
  examples?: Rule[]
}

export interface RuleGenResponse {
  rule: Rule
  yaml: string
  reasoning: string
  confidence: number
  warnings: string[]
}

export interface ExplainRequest {
  eventId: string
  question?: string
}

export interface ActionParams {
  ruleName?: string
  insightId?: string
  page?: string
  eventId?: string
  contextType?: string
}

export interface Action {
  label: string
  actionId: string
  params: ActionParams
}

export interface RelatedEvent {
  type: string
  timestamp: number
  pid?: number
  ppid?: number
  cgroupId?: string
  processName?: string
  filename?: string
  port?: number
  blocked: boolean
}

export interface ExplainResponse {
  explanation: string
  rootCause: string
  matchedRule?: Rule
  relatedEvents?: RelatedEvent[]
  suggestedActions?: Action[]
}

export interface AnalyzeRequest {
  type: 'process' | 'workload' | 'rule'
  id: string
}

export interface Anomaly {
  type: string
  description: string
  severity: string
  confidence: number
  evidence: string[]
}

export interface Recommendation {
  type: string
  description: string
  priority: string
  action: Action
}

export interface RelatedInsight {
  type: string
  title: string
  summary: string
}

export interface AnalyzeResponse {
  summary: string
  anomalies: Anomaly[]
  baselineStatus: string
  recommendations: Recommendation[]
  relatedInsights: RelatedInsight[]
}

export interface AskInsightRequest {
  insight: Insight
  question: string
}

export interface AskInsightResponse {
  answer: string
  confidence: number
}
