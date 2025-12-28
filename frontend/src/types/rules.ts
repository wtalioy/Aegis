// Rule Types - Phase 4

export type RuleState = 'draft' | 'testing' | 'production' | 'archived'

export type RuleAction = 'block' | 'monitor' | 'allow'

export type RuleSeverity = 'critical' | 'high' | 'warning' | 'info'

export interface Rule {
  name: string
  description: string
  state: RuleState
  action: RuleAction
  severity: RuleSeverity
  match: {
    process?: string
    filename?: string
    dest_port?: number
    cgroup?: string
    uid?: number
    uid_not?: number
    [key: string]: any
  }
  yaml: string
}

export interface RuleStats {
  hits: number
  blocks: number
  falsePositives?: number
  observationMinutes?: number
  // Legacy field for backward compatibility
  observationHours?: number
}

// Legacy alias maintained for backward compatibility
// Preferred explicit type for testing rules
export interface TestingRule extends Rule {
  state: 'testing'
  stats: RuleStats
}

