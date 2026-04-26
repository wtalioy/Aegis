import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { createRule, deleteRule, getRules, updateRule } from '../lib/api'
import type { Rule } from '../types/rules'
import type { RuleGenResponse } from '../types/ai'

export function usePolicyStudio() {
  const router = useRouter()
  const route = useRoute()
  const routeQuery = route.query as Record<string, string | string[] | undefined>

  const rules = ref<Rule[]>([])
  const selectedRule = ref<Rule | null>(null)
  const generatedRule = ref<RuleGenResponse | null>(null)
  const showDeployConfirm = ref(false)
  const createMode = ref<'manual' | 'ai'>('manual')
  const loading = ref(true)
  const searchQuery = ref('')
  const filterAction = ref<string>('all')

  const fetchRules = async () => {
    loading.value = true
    try {
      rules.value = await getRules()
    } catch (error) {
      console.error('Failed to fetch rules:', error)
      rules.value = []
    } finally {
      loading.value = false
    }
  }

  const filteredRules = computed(() => {
    let result = rules.value

    if (filterAction.value !== 'all') {
      result = result.filter((rule) => rule.action === filterAction.value)
    }

    if (searchQuery.value.trim()) {
      const query = searchQuery.value.toLowerCase()
      result = result.filter((rule) =>
        rule.name.toLowerCase().includes(query) ||
        rule.description.toLowerCase().includes(query)
      )
    }

    return result
  })

  const stats = computed(() => ({
    total: rules.value.length,
    block: rules.value.filter((rule) => rule.action === 'block').length,
    alert: rules.value.filter((rule) => rule.action === 'alert').length,
    allow: rules.value.filter((rule) => rule.action === 'allow').length,
    testing: rules.value.filter((rule) => rule.state === 'testing').length,
    production: rules.value.filter((rule) => rule.state === 'production').length
  }))

  const handleRuleSelect = (rule: Rule) => {
    selectedRule.value = rule
    createMode.value = 'manual'
    generatedRule.value = null
  }

  const handleManualRuleCreated = (rule: Partial<Rule>) => {
    generatedRule.value = {
      rule: rule as Rule,
      yaml: rule.yaml || '',
      reasoning: 'Manually created rule',
      confidence: 1,
      warnings: []
    }
    createMode.value = 'ai'
  }

  const handleManualRuleUpdated = async (rule: Partial<Rule>) => {
    if (!selectedRule.value) {
      return
    }

    try {
      const updatedRule = await updateRule(selectedRule.value.name, rule as Rule)
      await fetchRules()
      selectedRule.value = updatedRule
      selectedRule.value = null
    } catch (error) {
      console.error('Failed to update rule:', error)
      alert(`Failed to update rule: ${error instanceof Error ? error.message : 'Unknown error'}`)
    }
  }

  const handleEditCancel = () => {
    selectedRule.value = null
    generatedRule.value = null
    createMode.value = 'manual'
  }

  const handleRuleDeleted = async (ruleName: string) => {
    try {
      await deleteRule(ruleName)
      await fetchRules()
      selectedRule.value = null
      generatedRule.value = null
      createMode.value = 'manual'
    } catch (error) {
      console.error('Failed to delete rule:', error)
      alert(`Failed to delete rule: ${error instanceof Error ? error.message : 'Unknown error'}`)
    }
  }

  const handleConfirmDeploy = async () => {
    const ruleToDeploy = generatedRule.value?.rule || selectedRule.value
    if (!ruleToDeploy) {
      console.error('No rule to deploy')
      return
    }

    try {
      await createRule({
        ...ruleToDeploy,
        state: ruleToDeploy.state || 'testing'
      })

      showDeployConfirm.value = false
      await fetchRules()
      generatedRule.value = null
      createMode.value = 'manual'

      router.push({
        path: '/rule-validation',
        query: {
          rule: ruleToDeploy.name,
          from: 'deploy'
        }
      })
    } catch (error) {
      console.error('Failed to deploy rule:', error)
      alert(`Failed to deploy rule: ${error instanceof Error ? error.message : 'Unknown error'}`)
    }
  }

  onMounted(async () => {
    await fetchRules()
    const ruleParam = typeof routeQuery.rule === 'string' ? routeQuery.rule : null
    if (!ruleParam) {
      return
    }
    const match = rules.value.find((rule) => rule.name === ruleParam)
    if (match) {
      selectedRule.value = match
    }
  })

  return {
    rules,
    selectedRule,
    generatedRule,
    showDeployConfirm,
    createMode,
    loading,
    searchQuery,
    filterAction,
    filteredRules,
    stats,
    fetchRules,
    handleRuleSelect,
    handleManualRuleCreated,
    handleManualRuleUpdated,
    handleEditCancel,
    handleRuleDeleted,
    handleConfirmDeploy
  }
}
