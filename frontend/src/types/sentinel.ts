// Sentinel Types - Phase 4

export type InsightType = 
  | 'shadow_promotion' // legacy name for testing promotion
  | 'testing_promotion' // new preferred name
  | 'anomaly'
  | 'optimization'
  | 'daily_report'

export type Severity = 
  | 'low'
  | 'medium'
  | 'high'
  | 'critical'

export interface Action {
  label: string
  action_id: string
  params: Record<string, any>
}

export interface Insight {
  id: string
  type: InsightType
  title: string
  summary: string
  confidence: number
  severity: Severity
  data: Record<string, any>
  actions: Action[]
  created_at: string
}

export interface ShadowPromotionInsight extends Insight {
  type: 'shadow_promotion'
  data: {
    rule_name: string
    hits: number
    observation_hours: number
    false_positives?: number
  }
}

export interface TestingPromotionInsight extends Insight {
  type: 'testing_promotion'
  data: {
    rule_name: string
    hits: number
    observation_hours: number
    false_positives?: number
  }
}

export interface AnomalyInsight extends Insight {
  type: 'anomaly'
  data: {
    process_id?: number
    process_name?: string
    anomaly_type: string
    deviation: number
  }
}

export interface OptimizationInsight extends Insight {
  type: 'optimization'
  data: {
    rule_names: string[]
    suggestion: string
  }
}

export interface DailyReportInsight extends Insight {
  type: 'daily_report'
  data: {
    date: string
    summary: string
  }
}

