import { ref, onMounted, onUnmounted } from 'vue'
import { promoteRule } from '../lib/api'
import { requestJSON } from '../lib/http'
import type { Insight } from '../types/sentinel'

const API_BASE = '/api/v1'

export function useSentinel() {
  const insights = ref<Insight[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const eventSource = ref<EventSource | null>(null)
  const connected = ref(false)

  const fetchInsights = async (limit = 50) => {
    loading.value = true
    error.value = null
    try {
      const data = await requestJSON<{ insights: Insight[], total: number }>(`${API_BASE}/sentinel/insights?limit=${limit}`)
      const uniqueMap = new Map<string, Insight>()
      data.insights.forEach((insight) => uniqueMap.set(insight.id, insight))
      insights.value = Array.from(uniqueMap.values())
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch insights'
    } finally {
      loading.value = false
    }
  }

  const subscribe = () => {
    unsubscribe()

    try {
      eventSource.value = new EventSource(`${API_BASE}/sentinel/stream`)

      eventSource.value.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          if (data.type === 'heartbeat') {
            connected.value = true
            error.value = null
            return
          }

          const insight = data as Insight
          connected.value = true
          error.value = null
          if (!insights.value.some(existing => existing.id === insight.id)) {
            insights.value.unshift(insight)
          }
          if (insights.value.length > 100) {
            insights.value = insights.value.slice(0, 100)
          }
        } catch (err) {
          console.error('[Sentinel] Failed to parse message:', err)
        }
      }

      eventSource.value.onerror = () => {
        connected.value = false
        if (eventSource.value?.readyState === EventSource.CLOSED) {
          error.value = 'Connection closed. EventSource will attempt to reconnect automatically.'
        }
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to connect to Sentinel'
      connected.value = false
    }
  }

  const unsubscribe = () => {
    if (eventSource.value) {
      eventSource.value.close()
      eventSource.value = null
    }
    connected.value = false
  }

  const executeAction = async (insight: Insight, actionId: string) => {
    const action = insight.actions.find(item => item.actionId === actionId)
    if (!action) {
      return
    }

    try {
      switch (actionId) {
        case 'promote':
          if (typeof action.params.ruleName === 'string') {
            await promoteRule(action.params.ruleName)
            insights.value = insights.value.filter(item => item.id !== insight.id)
          }
          break
        case 'investigate':
          if (typeof action.params.page === 'string') {
            window.location.href = `/${action.params.page}`
          }
          break
        case 'navigate':
          if (typeof action.params.page === 'string') {
            window.location.href = `/${action.params.page}`
          }
          break
        case 'dismiss':
          insights.value = insights.value.filter(item => item.id !== insight.id)
          break
        default:
          console.warn('Unknown action:', actionId)
      }
    } catch (err) {
      console.error('Failed to execute action:', err)
    }
  }

  onMounted(() => {
    fetchInsights()
    subscribe()
  })

  onUnmounted(() => {
    unsubscribe()
  })

  return {
    insights,
    loading,
    error,
    connected,
    fetchInsights,
    subscribe,
    unsubscribe,
    executeAction
  }
}

export type { Insight }
