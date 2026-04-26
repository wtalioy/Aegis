export type RuleState = 'draft' | 'testing' | 'production' | 'archived'
export type RuleAction = 'block' | 'alert' | 'allow'
export type RuleSeverity = 'critical' | 'high' | 'warning' | 'info'
export type RuleType = 'exec' | 'file' | 'connect'
export type MatchType = 'exact' | 'contains' | 'prefix'

export interface RuleMatch {
  processName?: string
  processNameType?: MatchType
  parentName?: string
  parentNameType?: MatchType
  pid?: number
  ppid?: number
  filename?: string
  destPort?: number
  destIp?: string
  cgroupId?: string
  uid?: number
}

export interface Rule {
  name: string
  description: string
  state: RuleState
  action: RuleAction
  severity: RuleSeverity
  type: RuleType
  match: RuleMatch
  yaml: string
  createdAt?: string
  deployedAt?: string
  promotedAt?: string
}

export interface RuleValidationStats {
  hits: number
  observationMinutes: number
}

export interface RuleValidation {
  isReady: boolean
  score?: number
}

export interface TestingRule extends Rule {
  state: 'testing'
  stats: RuleValidationStats
  validation: RuleValidation
}
