import { ref, onMounted, onUnmounted } from 'vue'
import { getAlerts, subscribeToAlerts, type Alert } from '../lib/api'

export { Alert }

export function useAlerts() {
  const alerts = ref<Alert[]>([])
  const newAlertCount = ref(0)

  let unsubscribe: (() => void) | null = null
  let lastAlertIds = new Set<string>()

  const fetchAlerts = async () => {
    try {
      const result = await getAlerts()
      alerts.value = result || []
    } catch (e) {
      console.error('Failed to fetch alerts:', e)
    }
  }

  const handleAlertsUpdate = (newAlerts: Alert[]) => {
    const newIds = new Set(newAlerts.map(a => a.id))
    const addedCount = newAlerts.filter(a => !lastAlertIds.has(a.id)).length
    
    if (addedCount > 0) {
      newAlertCount.value += addedCount
    }
    
    alerts.value = newAlerts.slice(0, 100)
    lastAlertIds = newIds
  }

  const clearNewCount = () => {
    newAlertCount.value = 0
  }

  const getAlertsBySeverity = () => {
    const high = alerts.value.filter(a => a.severity === 'high').length
    const warning = alerts.value.filter(a => a.severity === 'warning').length
    const info = alerts.value.filter(a => a.severity === 'info').length
    return { high, warning, info }
  }

  onMounted(async () => {
    await fetchAlerts()
    lastAlertIds = new Set(alerts.value.map(a => a.id))
    unsubscribe = subscribeToAlerts(handleAlertsUpdate)
  })

  onUnmounted(() => {
    unsubscribe?.()
  })

  return {
    alerts,
    newAlertCount,
    clearNewCount,
    fetchAlerts,
    getAlertsBySeverity
  }
}
