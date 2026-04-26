export type InsightType =
  | 'testingPromotion'
  | 'anomaly'
  | 'optimization'
  | 'dailyReport'

export type InsightSeverity =
  | 'low'
  | 'medium'
  | 'high'
  | 'critical'

export interface InsightAction {
  label: string
  actionId: string
  params: {
    ruleName?: string
    insightId?: string
    page?: string
    eventId?: string
    contextType?: string
  }
}

export interface TestingPromotionInsightData {
  ruleName: string
  hits: number
  observationHours: number
  falsePositives?: number
}

export interface AnomalyInsightData {
  kind?: string
  eventCount?: number
  processId?: number
  processName?: string
  anomalyType: string
  deviation: number
}

export interface OptimizationInsightData {
  ruleNames: string[]
  suggestion: string
  ruleCount?: number
}

export interface DailyReportInsightData {
  kind?: string
  date: string
  summary: string
}

export type InsightData =
  | TestingPromotionInsightData
  | AnomalyInsightData
  | OptimizationInsightData
  | DailyReportInsightData
  | Record<string, string | number | boolean | string[] | undefined>

export interface Insight {
  id: string
  type: InsightType
  title: string
  summary: string
  confidence: number
  severity: InsightSeverity
  data: InsightData
  actions: InsightAction[]
  createdAt: string
}
