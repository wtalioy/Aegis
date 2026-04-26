// AI Core Composable
import { ref } from 'vue'
import { analyzeContext, askAboutInsight, explainEvent, generateRule } from '../lib/api/analysis'
import type {
  AnalyzeRequest,
  AnalyzeResponse,
  AskInsightRequest,
  AskInsightResponse,
  ExplainRequest,
  ExplainResponse,
  RuleGenRequest,
  RuleGenResponse
} from '../types/ai'

export function useAI() {
  const loading = ref(false)
  const error = ref<string | null>(null)

  const runGenerateRule = async (req: RuleGenRequest): Promise<RuleGenResponse | null> => {
    loading.value = true
    error.value = null
    try {
      return await generateRule(req)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to generate rule'
      return null
    } finally {
      loading.value = false
    }
  }

  const runExplainEvent = async (req: ExplainRequest): Promise<ExplainResponse | null> => {
    loading.value = true
    error.value = null
    try {
      return await explainEvent(req)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to explain event'
      return null
    } finally {
      loading.value = false
    }
  }

  const runAnalyzeContext = async (req: AnalyzeRequest): Promise<AnalyzeResponse | null> => {
    loading.value = true
    error.value = null
    try {
      return await analyzeContext(req)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to analyze context'
      return null
    } finally {
      loading.value = false
    }
  }

  const runAskAboutInsight = async (req: AskInsightRequest): Promise<AskInsightResponse | null> => {
    loading.value = true
    error.value = null
    try {
      return await askAboutInsight(req)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to ask about insight'
      return null
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    error,
    generateRule: runGenerateRule,
    explainEvent: runExplainEvent,
    analyzeContext: runAnalyzeContext,
    askAboutInsight: runAskAboutInsight
  }
}
