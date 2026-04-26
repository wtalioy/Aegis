import { requestJSON } from '../http'
import type { Rule, TestingRule } from '../../types/rules'

const API_BASE = '/api/v1/policies'

export async function getRules(): Promise<Rule[]> {
  return requestJSON<Rule[]>(API_BASE)
}

export async function createRule(rule: Rule): Promise<Rule> {
  const data = await requestJSON<{ success: boolean, rule: Rule }>(API_BASE, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ rule })
  })
  return data.rule
}

export async function updateRule(ruleName: string, rule: Rule): Promise<Rule> {
  const encodedName = encodeURIComponent(ruleName)
  const data = await requestJSON<{ success: boolean, rule: Rule }>(`${API_BASE}/${encodedName}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ rule })
  })
  return data.rule
}

export async function deleteRule(ruleName: string): Promise<void> {
  const encodedName = encodeURIComponent(ruleName)
  await requestJSON<{ success: boolean }>(`${API_BASE}/${encodedName}`, {
    method: 'DELETE'
  })
}

export async function getRuleValidation(ruleId: string): Promise<TestingRule> {
  return requestJSON<TestingRule>(`${API_BASE}/validation/${encodeURIComponent(ruleId)}`)
}

export async function getTestingRules(): Promise<TestingRule[]> {
  return requestJSON<TestingRule[]>(`${API_BASE}/testing`)
}

export async function promoteRule(ruleId: string): Promise<{ success: boolean }> {
  return requestJSON<{ success: boolean }>(`${API_BASE}/${encodeURIComponent(ruleId)}/promote`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' }
  })
}
