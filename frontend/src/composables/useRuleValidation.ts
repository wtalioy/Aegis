import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getRuleValidation, getSettings, getTestingRules, promoteRule } from '../lib/api'
import type { Rule, TestingRule } from '../types/rules'

export function useRuleValidation() {
  const route = useRoute()
  const router = useRouter()
  const routeQuery = route.query as Record<string, string | string[] | undefined>

  const testingRules = ref<TestingRule[]>([])
  const selectedRule = ref<TestingRule | null>(null)
  const validationData = ref<TestingRule | null>(null)
  const loading = ref(true)
  const promoting = ref(false)
  const error = ref<string | null>(null)
  const newlyDeployedRule = ref<string | null>(null)
  const promotionMinObservationMinutes = ref(1440)
  const promotionMinHits = ref(100)

  let refreshInterval: ReturnType<typeof setInterval> | null = null

  const fetchTestingRulesList = async () => {
    loading.value = true
    error.value = null
    try {
      testingRules.value = await getTestingRules()
      if (testingRules.value.length > 0 && !selectedRule.value) {
        selectedRule.value = testingRules.value[0]
        await loadValidation(testingRules.value[0])
      }
    } catch (fetchError) {
      console.error('Failed to fetch testing rules:', fetchError)
      error.value = 'Failed to load testing rules'
    } finally {
      loading.value = false
    }
  }

  const loadValidation = async (rule: Rule) => {
    if (!rule.name) {
      error.value = 'Invalid rule'
      return
    }
    try {
      validationData.value = await getRuleValidation(rule.name)
      error.value = null
    } catch (loadError) {
      console.error('Failed to load validation data:', loadError)
      error.value = `Failed to load validation data: ${loadError instanceof Error ? loadError.message : 'Unknown error'}`
    }
  }

  const stopAutoRefresh = () => {
    if (refreshInterval) {
      clearInterval(refreshInterval)
      refreshInterval = null
    }
  }

  const startAutoRefresh = () => {
    stopAutoRefresh()
    refreshInterval = setInterval(async () => {
      await fetchTestingRulesList()
      if (!selectedRule.value) {
        return
      }
      const refreshed = testingRules.value.find((rule) => rule.name === selectedRule.value?.name)
      if (refreshed) {
        selectedRule.value = refreshed
      }
      await loadValidation(selectedRule.value)
    }, 5000)
  }

  const handleSelectRule = async (rule: Rule) => {
    selectedRule.value = rule as TestingRule
    await loadValidation(rule)
    startAutoRefresh()
  }

  const handleAdjustRule = () => {
    if (!selectedRule.value) {
      return
    }
    router.push({ path: '/policy-studio', query: { rule: selectedRule.value.name } })
  }

  const handlePromote = async (_force = false) => {
    if (!selectedRule.value) {
      return
    }

    promoting.value = true
    error.value = null
    try {
      await promoteRule(selectedRule.value.name)
      await fetchTestingRulesList()
      if (selectedRule.value) {
        await loadValidation(selectedRule.value)
      }
    } catch (promoteError) {
      console.error('Failed to promote rule:', promoteError)
      error.value = promoteError instanceof Error ? promoteError.message : 'Failed to promote rule'
    } finally {
      promoting.value = false
    }
  }

  const isReady = computed(() => {
    const ruleStats = validationData.value?.stats
    const validation = validationData.value?.validation
    if (!ruleStats && !validation) {
      return false
    }
    const obsMinutes = ruleStats?.observationMinutes ?? 0
    const matches = ruleStats?.hits ?? 0
    const criteriaReady = obsMinutes >= promotionMinObservationMinutes.value && matches >= promotionMinHits.value
    return typeof validation?.isReady === 'boolean' ? validation.isReady : criteriaReady
  })

  const stats = computed(() => ({
    totalTesting: testingRules.value.length,
    readyToPromote: testingRules.value.filter((rule) => (rule.validation.score || 0) >= 0.7).length,
    avgObservationTime: testingRules.value.length > 0
      ? testingRules.value.reduce((sum, rule) => sum + rule.stats.observationMinutes, 0) / testingRules.value.length / 60
      : 0
  }))

  onMounted(async () => {
    try {
      const settings = await getSettings()
      if (settings.policy?.promotion_min_observation_minutes) {
        promotionMinObservationMinutes.value = settings.policy.promotion_min_observation_minutes
      }
      if (settings.policy?.promotion_min_hits) {
        promotionMinHits.value = settings.policy.promotion_min_hits
      }
    } catch (settingsError) {
      console.warn('Failed to fetch promotion config, using defaults:', settingsError)
    }

    const fromDeploy = routeQuery.from === 'deploy'
    if (fromDeploy) {
      await new Promise((resolve) => setTimeout(resolve, 500))
    }

    await fetchTestingRulesList()
    const ruleParam = typeof routeQuery.rule === 'string' ? routeQuery.rule : null

    if (ruleParam && testingRules.value.length > 0) {
      const match = testingRules.value.find((rule) => rule.name === ruleParam)
      if (match) {
        await handleSelectRule(match)
        if (fromDeploy) {
          newlyDeployedRule.value = ruleParam
          setTimeout(() => {
            newlyDeployedRule.value = null
          }, 8000)
        }
      }
    } else if (fromDeploy && testingRules.value.length === 0) {
      setTimeout(async () => {
        await fetchTestingRulesList()
        if (!ruleParam || testingRules.value.length === 0) {
          return
        }
        const match = testingRules.value.find((rule) => rule.name === ruleParam)
        if (match) {
          await handleSelectRule(match)
          newlyDeployedRule.value = ruleParam
          setTimeout(() => {
            newlyDeployedRule.value = null
          }, 8000)
        }
      }, 1000)
    }
  })

  onBeforeUnmount(() => {
    stopAutoRefresh()
  })

  return {
    testingRules,
    selectedRule,
    validationData,
    loading,
    promoting,
    error,
    newlyDeployedRule,
    isReady,
    stats,
    handleSelectRule,
    handleAdjustRule,
    handlePromote
  }
}
