// Sentinel Composable - Phase 4
import { ref, onMounted, onUnmounted } from 'vue'

export interface Insight {
  id: string
  type: 'testing_promotion' | 'anomaly' | 'optimization' | 'daily_report'
  title: string
  summary: string
  confidence: number
  severity: 'low' | 'medium' | 'high' | 'critical'
  data: Record<string, any>
  actions: Array<{
    label: string
    action_id: string
    params: Record<string, any>
  }>
  created_at: string
}

const API_BASE = '/api'

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
      const response = await fetch(`${API_BASE}/ai/sentinel/insights?limit=${limit}`)
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`)
      }
      const data = await response.json()
      insights.value = data.insights || []
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch insights'
    } finally {
      loading.value = false
    }
  }

  const subscribe = () => {
    // Close existing connection if any
    if (eventSource.value) {
      eventSource.value.close()
      eventSource.value = null
    }

    const sseUrl = `${API_BASE}/ai/sentinel/stream`
    
    try {
      eventSource.value = new EventSource(sseUrl)
      
      // EventSource connection is established when the object is created
      // Check readyState immediately and set connected
      const checkConnection = () => {
        if (eventSource.value) {
          if (eventSource.value.readyState === EventSource.OPEN || eventSource.value.readyState === EventSource.CONNECTING) {
            connected.value = true
            error.value = null
            console.log('[Sentinel] SSE connecting/connected, readyState:', eventSource.value.readyState)
          }
        }
      }
      
      // Check immediately
      checkConnection()
      
      // Also check after a short delay to ensure connection is established
      setTimeout(checkConnection, 100)
      
      eventSource.value.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          
          // Handle heartbeat - this confirms connection is alive
          if (data.type === 'heartbeat') {
            connected.value = true
            error.value = null
            return
          }
          
          // Handle insight
          const insight: Insight = data
          // Add to beginning of list
          insights.value.unshift(insight)
          // Keep only last 100
          if (insights.value.length > 100) {
            insights.value = insights.value.slice(0, 100)
          }
        } catch (err) {
          console.error('[Sentinel] Failed to parse message:', err)
        }
      }
      
      eventSource.value.onerror = (err) => {
        console.error('[Sentinel] SSE error:', err, 'readyState:', eventSource.value?.readyState)
        
        // EventSource readyState: 0=CONNECTING, 1=OPEN, 2=CLOSED
        if (eventSource.value?.readyState === EventSource.CLOSED) {
          connected.value = false
          error.value = 'Connection closed. EventSource will attempt to reconnect automatically.'
        } else if (eventSource.value?.readyState === EventSource.CONNECTING) {
          // Still connecting, don't mark as disconnected yet
          connected.value = false
        }
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to connect to Sentinel'
      connected.value = false
      console.error('[Sentinel] Failed to create EventSource:', err)
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
    const action = insight.actions.find(a => a.action_id === actionId)
    if (!action) {
      console.warn('Action not found:', actionId)
      return
    }

    try {
      switch (actionId) {
        case 'promote':
          // Call promote rule API
          const response = await fetch(`${API_BASE}/rules/${action.params.rule_name}/promote`, {
            method: 'POST'
          })
          if (response.ok) {
            // Remove insight or mark as resolved
            insights.value = insights.value.filter(i => i.id !== insight.id)
          }
          break

        case 'investigate':
          // Navigate to investigation page
          window.location.href = `/investigation?event=${action.params.event_id}`
          break

        case 'dismiss':
          // Remove insight
          insights.value = insights.value.filter(i => i.id !== insight.id)
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

